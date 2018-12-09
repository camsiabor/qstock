package sync

import (
	"bytes"
	"encoding/json"
	"github.com/camsiabor/qcom/util/qlog"
	"github.com/camsiabor/qcom/util/qtime"
	"github.com/camsiabor/qcom/util/util"
	"github.com/camsiabor/qstock/dict"
	"github.com/camsiabor/qstock/sync/showSdk/httplib"
	"github.com/pkg/errors"
	"time"
)

// https://tushare.pro/document/2?doc_id=123
/*
curl -X POST -d '{"api_name": "trade_cal", "token": "xxxxxxxx", "params": {"exchange":"", "start_date":"20180901", "end_date":"20181001", "is_open":"0"}, "fields": "exchange,cal_date,is_open,pretrade_date"}' http://api.tushare.pro
 */


func (o Syncer) TuShare_request(
	work * ProfileWork,
	fields []string,
	requestargs map[string]interface{}) (ret interface{}, err error) {

	if (!o.doContinue) {
		return;
	}
	var profile = work.Profile;
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
	var timeout = util.GetInt64(profile, 20, "timeout");
	var nice = util.GetInt64(profile, 250, "nice");
	req.SetTimeout(time.Duration(10) * time.Second, time.Duration(timeout) * time.Second);
	if (nice > 0) {
		time.Sleep(time.Millisecond * time.Duration(nice));
	}
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

	var data = util.GetMap(m, false, "data");
	var cols = util.GetStringSlice(data, "fields");
	var rows = util.GetSlice(data, "items");
	var datalen = len(rows);
	if (datalen <= 0){
		return nil, nil;
	}
	maps, err := util.ColRowToMaps(cols, rows);
	if (err != nil) {
		return nil, err;
	}
	maps, _, err = o.PersistAndCache(work, maps);
	return maps, err;

}



func (o * Syncer) TuShare_khistory(phrase string, work * ProfileWork) (err error) {

	if (phrase != "work") {
		return nil;
	}


	if (err != nil) {
		qlog.Log(qlog.ERROR, err);
		return err;
	}
	var dao = work.Dao;
	var profile = work.Profile;
	var profilename = work.ProfileName;

	var db = util.GetStr(profile, dict.DB_HISTORY, "db");
	var metatoken = o.GetMetaToken(profilename);
	//var profileRunInfo = o.GetProfileRunInfo(profilename);
	var market= util.GetStr(profile, "", "marker");
	var fetcheach = util.GetInt(profile, 30, "each");
	var from_date_str, _ = util.AsStrErr(dao.Get(db, metatoken, "fetch_last_date", false));
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

	var perr error;
	var rargs= make(map[string]interface{});
	for retry := 1; retry <= 3; retry++ {
		var fails = make([]string, len(codes));
		var failcount = 0;
		for _, code := range codes {
			rargs["ts_code"] = code + keysuffix;
			rargs["start_date"] = from_date_str;
			rargs["end_date"] = to_date_str;
			_, err := o.TuShare_request(work, nil, rargs);
			if (err == nil) {
				qlog.Log(qlog.INFO, profilename, "persist", code, from_date_str, to_date_str);
			} else {
				qlog.Log(qlog.ERROR, profilename, "persist", "fail", code, from_date_str, to_date_str, err.Error());
				fails[failcount] = code;
				failcount++;
			}
		}
		if (failcount > 0) {
			codes = fails;
			qlog.Log(qlog.ERROR, profilename, "persist", "failcount", failcount, "retry", retry);
		} else {
			break;
		}
	}
	return perr;
}