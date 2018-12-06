package sync

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/camsiabor/qcom/qdao"
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qcom/util/qlog"
	"github.com/camsiabor/qcom/util/qtime"
	"showSdk/httplib"
	"github.com/camsiabor/qstock/dict"
	"time"
)

// https://tushare.pro/document/2?doc_id=123
/*
curl -X POST -d '{"api_name": "trade_cal", "token": "xxxxxxxx", "params": {"exchange":"", "start_date":"20180901", "end_date":"20181001", "is_open":"0"}, "fields": "exchange,cal_date,is_open,pretrade_date"}' http://api.tushare.pro
 */


func (o Syncer) TuShare_request(
	dao qdao.D,
	profile map[string]interface{},
	profilename string,
	fields []string,
	requestargs map[string]interface{},
	handler SyncAPIHandler) (ret interface{}, err error) {

	if (!o.doContinue) {
		return;
	}
	var req = httplib.Post(o.domain)
	var reqm = make(map[string]interface{});
	var api = util.GetStr(profile, "", "api");
	reqm["token"] = o.appsecret;
	reqm["api_name"] = api;
	if (fields != nil && len(fields) > 0) {
		reqm["fields"] = fields;
	}
	reqm["params"] = requestargs;
	ret, err = json.Marshal(reqm);
	if (err != nil) {
		return
	}
	req.Body(ret);
	httpresp, err := req.DoRequest();
	if err != nil {
		return;
	}
	var m map[string]interface{};
	var buffer = new(bytes.Buffer);
	buffer.ReadFrom(httpresp.Body);
	err = json.Unmarshal(buffer.Bytes(), &m)
	if (err != nil) {
		return nil, err;
	}

	var retcode = util.GetInt(m, 0, "code");
	if (retcode != 0) {
		var retmsg = util.GetStr(m, "", "msg");
		return nil, errors.New(retmsg);
	}
	var mapperName = util.GetStr(profile, "", "mapper");
	var data = util.GetMap(m, false, "data");
	var cols = util.GetStringSlice(data, "fields");
	var rows = util.GetSlice(data, "items");
	var mapper = util.GetMapperManager().Get(mapperName);
	maps, err := util.ColRowToMaps(cols, rows, mapper);
	return maps, err;
}



func (o * Syncer) TuShare_khistory(

	phrase string, dao qdao.D,
	profile map[string]interface{}, profilename string,
	arg1 interface{}, arg2 interface{}) (err error) {

	if (phrase != "work") {
		return nil;
	}


	if (err != nil) {
		qlog.Log(qlog.ERROR, err);
		return err;
	}

	var metatoken = o.GetMetaToken(profilename);
	//var profileRunInfo = o.GetProfileRunInfo(profilename);
	var market= util.GetStr(profile, "", "marker");
	var fetcheach = util.GetInt(profile, 30, "each");
	var from_date_str, _ = util.AsStrErr(dao.Get(dict.DB_HISTORY, metatoken, "fetch_last_date", false));
	if (len(from_date_str) <= 0) {
		from_date_str = time.Now().AddDate(0, 0, -fetcheach).Format("20060102");
	}
	var to_date = time.Now();
	var from_date, _ = time.Parse("20060102", from_date_str);
	var interval = qtime.TimeInterval(&to_date, &from_date, time.Hour * 24);
	if (interval > 60) {
		from_date = to_date.AddDate(0, 0, -60);
		from_date_str = from_date.Format("20060102");
	}
	var to_date_str = to_date.Format("20060102");
	var keyprefix string;
	var keysuffix string;
	if (market == "sz") {
		keyprefix = "00*";
		keysuffix = ".SZ";
	} else {
		keyprefix = "60*";
		keysuffix = ".SH";
	}
	codes, err := dao.Keys(dict.DB_DEFAULT, "",keyprefix);

	if (err != nil) {
		qlog.Log(qlog.ERROR, "persist", "khistory", market, "fetch keys error", err);
		return err;
	}

	var cachername = util.GetStr(profile, dict.CACHE_STOCK_KHISTORY, "cacher");
	var cacher * scache.SCache;
	if (len(cachername) > 0) {
		cacher = scache.GetCacheManager().Get(cachername);
	}
	var rargs= make(map[string]interface{});
	for _, code := range codes {
		rargs["ts_code"] = code + keysuffix;
		rargs["start_date"] = from_date_str;
		rargs["end_date"] = to_date_str;
		data, err := o.TuShare_request(dao, profile, profilename, nil, rargs, nil);
		if (data == nil || err != nil) {
			continue;
		}
		var list = data.([]interface{});
		var listlen = len(list);
		if (listlen == 0) {
			continue;
		}
		var index = 0;
		var dates = make([]interface{}, listlen);
		var infos = make([]interface{}, listlen);
		for _, info := range list {
			var minfo = info.(map[string]interface{});
			var datestr = util.AsStr(minfo["date"], "");
			//delete(minfo, "code");
			//delete(minfo, "market");
			//delete(minfo, "stockName");
			dates[index] = datestr;
			infos[index] = minfo;
			index = index + 1;
			if (cacher != nil){
				cacher.SetSubVal(minfo, code, datestr);
			}
		}
		var _, rerr = dao.Updates(dict.DB_HISTORY, code, dates, infos, true, true);
		if (rerr != nil) {
			qlog.Log(qlog.ERROR, "api", "showapi", "khistory", rerr);
		}
	}

	if (err != nil) {
		dao.Update(dict.DB_HISTORY, metatoken, "fetch_last_date", to_date_str, true, false);
	}


	return err;
}