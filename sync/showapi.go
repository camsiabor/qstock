package sync

import (
	"encoding/json"
	"fmt"
	"github.com/camsiabor/qcom/qdao"
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/util/qerr"
	"github.com/camsiabor/qcom/util/util"
	"github.com/camsiabor/qcom/util/qlog"
	"github.com/camsiabor/qcom/util/qtime"
	"github.com/camsiabor/qstock/dict"
	"showSdk/normalRequest"
	"strings"
	"time"
)

func (o * Syncer) ShowAPI_request(
	dao qdao.D,
	profile map[string]interface{},
	profilename string,
	requestargs map[string]interface{},
	handler SyncAPIHandler) (interface{}, error) {

	if (!o.doContinue) {
		return nil, nil;
	}

	var nice = util.GetInt64(profile, 100, "nice" );
	if (nice > 0) {
		time.Sleep(time.Duration(nice) * time.Millisecond);
	}

	var api = util.GetStr(profile, "", "api");
	var request = normalRequest.ShowapiRequest(o.domain+  "/" + api, o.appid, o.appsecret)
	if (requestargs != nil) {
		for k, v := range requestargs {
			request.AddTextPara(k, v.(string))
		}
	}

	var jsonstr, err = request.Post()
	if err != nil {
		qlog.Log(qlog.FATAL, o.Name, err)
		return nil, err
	}

	var bytes = []byte(jsonstr)
	var m map[string]interface{};
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		qlog.Log(qlog.FATAL, o.Name, err)
		return nil, err
	}

	var showapi_res_code = m["showapi_res_code"].(float64)
	if showapi_res_code != 0 {
		var showapi_res_error = m["showapi_res_error"].(string)
		qlog.Log(qlog.ERROR, o.Name, "code", showapi_res_code, "msg", showapi_res_error)
		return nil, qerr.NewCError(int(showapi_res_code), o.Name, showapi_res_error);
	}

	var resbody, _ = m["showapi_res_body"].(map[string]interface{})
	var ret_code = resbody["ret_code"].(float64)
	if ret_code != 0 {
		var remark = resbody["remark"].(string)
		qlog.Log(qlog.ERROR, o.Name, "ret_code", ret_code, "remark", remark)
		return nil, qerr.NewCError(int(ret_code), o.Name, remark);
	}

	var prefix = util.GetStr(profile, "", "prefix");
	var primarykey = util.GetStr(profile, "code", "key");



	var mappername = util.GetStr(profile, "", "mapper");
	var mapper = util.GetMapperManager().Get(mappername);

	//var hastable = len(table) > 0;


	var updatetime = time.Now().Format("02-1504"); // updateimte
	var list = resbody["list"].([]interface{})
	var ids = make([]interface{}, len(list));
	for i, one := range list {
		var info = one.(map[string]interface{});
		if (mapper != nil) {
			_, err = mapper.Map(info, false);
			if (err != nil) {
				qlog.Log(qlog.ERROR, err);
				break;
			}
		}
		var stockcode string = info[primarykey].(string);
		var id string = prefix + stockcode;
		ids[i] = id;
		info["_u"] = updatetime;
	}


	if (err == nil) {

		var db = util.GetStr(profile, "", "db");
		var group = util.GetStr(profile, "", "group");
		var cachername = util.GetStr(profile, "", "cacher");
		var cacher = scache.GetCacheManager().Get(cachername)

		var idsss = util.AsStringSlice(ids, 0);
		if (len(group) == 0) {
			if (cacher != nil) {
				cacher.Sets(list, idsss);
			}
			_, err= dao.Updates(db, group, ids, list, true, false);
		} else {
			if (cacher != nil) {
				cacher.SetSubVals(list, idsss, group);
			}
			_, err= dao.Updates(db, group, ids, list, true, true);
		}
	}

	return list, nil;
}

func (o * Syncer) ShowAPI_snapshot(
	phrase string, dao qdao.D,
	profile map[string]interface{}, profilename string,
	arg1 interface{}, arg2 interface{} ) (err error) {

	if (phrase != "work") {
		return nil;
	}
	var market = util.GetStr(profile, "", "marker");
	var fetcheach int64 = util.GetInt64(profile, 10, "each");
	var fetchlimit int64 = util.GetInt64(profile, 5000, "limit");
	var start int64 = 0
	var formatstr=  ""
	if strings.Contains(market, "sz") {
		start = 0
		formatstr = "sz%6d"
	} else if strings.Contains(market, "sh") {
		start = 600000
		formatstr = "sh%6d"
	} else {
		return fmt.Errorf("unknown marker %s", market);
	}

	dao.SelectDB(dict.DB_DEFAULT);

	var stockids= ""
	var rargs= make(map[string]interface{});
	for code := 0 + start; code <= fetchlimit+start; code++ {
		var stockid = fmt.Sprintf(formatstr, code)
		stockid = strings.Replace(stockid, " ", "0", -1)
		stockids = stockids + stockid + ","
		if code%fetcheach == 0 {
			stockids = strings.TrimRight(stockids, ",")
			rargs["stocks"] = stockids;
			rargs["needIndex"] = "0";
			_, err = o.ShowAPI_request(dao, profile, profilename, rargs, nil)
			stockids = ""
		}
	}
	return err;
}



func (o * Syncer) ShowAPI_khistory(

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
		from_date_str = time.Now().AddDate(0, 0, -fetcheach).Format("2006-01-02");
	}
	var to_date = time.Now();
	var from_date, _ = time.Parse("2006-01-02", from_date_str);
	var interval = qtime.TimeInterval(&to_date, &from_date, time.Hour * 24);
	if (interval > 60) {
		from_date = to_date.AddDate(0, 0, -60);
		from_date_str = from_date.Format("2006-01-02");
	}
	var to_date_str = to_date.Format("2006-01-02");
	var keyprefix string;
	if (market == "sz") {
		keyprefix = "00*";
	} else {
		keyprefix = "60*";
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
		rargs["code"] = code;
		rargs["begin"] = from_date_str;
		rargs["end"] = to_date_str;
		data, err := o.ShowAPI_request(dao, profile, profilename, rargs, nil);
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