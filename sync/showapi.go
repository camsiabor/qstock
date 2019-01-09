package sync

import (
	"encoding/json"
	"fmt"
	"github.com/camsiabor/qcom/qerr"
	"github.com/camsiabor/qcom/qlog"
	"github.com/camsiabor/qcom/qtime"
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qstock/dict"
	"github.com/camsiabor/qstock/sync/showSdk/normalRequest"
	"strings"
	"time"
)

func (o *Syncer) ShowAPI_request(
	work *ProfileWork,
	requestargs map[string]interface{}) (interface{}, error) {

	if !o.doContinue {
		return nil, nil
	}

	var profile = work.Profile
	var api = util.GetStr(profile, "", "api")
	var request = normalRequest.ShowapiRequest(o.domain+"/"+api, o.appid, o.appsecret)
	if requestargs != nil {
		for k, v := range requestargs {
			request.AddTextPara(k, v.(string))
		}
	}

	var timeout = util.GetInt64(profile, 20, "timeout")
	var nice = util.GetInt64(profile, 250, "nice")
	request.SetConnectTimeOut(time.Duration(timeout) * time.Second)
	request.SetConnectTimeOut(time.Duration(10) * time.Second)
	if nice >= 0 {
		time.Sleep(time.Millisecond * time.Duration(nice))
	}
	var jsonstr, err = request.Post()
	if err != nil {
		qlog.Log(qlog.FATAL, o.Name, err)
		return nil, err
	}

	var bytes = []byte(jsonstr)
	var m map[string]interface{}
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		qlog.Log(qlog.FATAL, o.Name, err)
		return nil, err
	}

	var showapi_res_code = m["showapi_res_code"].(float64)
	if showapi_res_code != 0 {
		var showapi_res_error = m["showapi_res_error"].(string)
		qlog.Log(qlog.ERROR, o.Name, "code", showapi_res_code, "msg", showapi_res_error)
		return nil, qerr.NewCError(int(showapi_res_code), o.Name, showapi_res_error)
	}
	var resbody, _ = m["showapi_res_body"].(map[string]interface{})
	var ret_code = resbody["ret_code"].(float64)
	if ret_code != 0 {
		var remark = resbody["remark"].(string)
		qlog.Log(qlog.ERROR, o.Name, "ret_code", ret_code, "remark", remark)
		return nil, qerr.NewCError(int(ret_code), o.Name, remark)
	}
	var list = resbody["list"].([]interface{})
	list, _, err = o.PersistAndCache(work, list)
	return list, err
}

func (o *Syncer) ShowAPI_snapshot(
	phrase string, work *ProfileWork) (err error) {

	if phrase != "work" {
		return nil
	}
	var profile = work.ProfileName
	var market = util.GetStr(profile, "", "marker")
	var fetcheach int64 = util.GetInt64(profile, 10, "each")
	var fetchlimit int64 = util.GetInt64(profile, 5000, "limit")
	var start int64 = 0
	var formatstr = ""
	if strings.Contains(market, "sz") {
		start = 0
		formatstr = "sz%6d"
	} else if strings.Contains(market, "sh") {
		start = 600000
		formatstr = "sh%6d"
	} else {
		return fmt.Errorf("unknown marker %s", market)
	}

	work.Dao.SelectDB(dict.DB_DEFAULT)

	var stockids = ""
	var rargs = make(map[string]interface{})
	for code := 0 + start; code <= fetchlimit+start; code++ {
		var stockid = fmt.Sprintf(formatstr, code)
		stockid = strings.Replace(stockid, " ", "0", -1)
		stockids = stockids + stockid + ","
		if code%fetcheach == 0 {
			stockids = strings.TrimRight(stockids, ",")
			rargs["stocks"] = stockids
			rargs["needIndex"] = "0"
			_, err = o.ShowAPI_request(work, rargs)
			stockids = ""
		}
	}
	return err
}

func (o *Syncer) ShowAPI_khistory(phrase string, work *ProfileWork) (err error) {

	if phrase != "work" {
		return nil
	}

	if err != nil {
		qlog.Log(qlog.ERROR, err)
		return err
	}

	var dao = work.Dao
	var profile = work.Profile
	var profilename = work.ProfileName
	var metatoken = o.GetMetaToken(profilename)
	//var profileRunInfo = o.GetProfileRunInfo(profilename);
	var market = util.GetStr(profile, "", "marker")
	var fetcheach = util.GetInt(profile, 30, "each")
	var from_date_str, _ = util.AsStrErr(dao.Get(dict.DB_HISTORY, metatoken, "fetch_last_date", 0, nil))
	if len(from_date_str) <= 0 {
		from_date_str = time.Now().AddDate(0, 0, -fetcheach).Format("2006-01-02")
	}
	var to_date = time.Now()
	var from_date, _ = time.Parse("2006-01-02", from_date_str)
	var interval = qtime.TimeInterval(&to_date, &from_date, time.Hour*24)
	if interval > 60 {
		from_date = to_date.AddDate(0, 0, -60)
		from_date_str = from_date.Format("2006-01-02")
	}
	var to_date_str = to_date.Format("2006-01-02")
	var keyprefix string
	if market == "sz" {
		keyprefix = "00*"
	} else {
		keyprefix = "60*"
	}
	codes, err := dao.Keys(dict.DB_DEFAULT, "", keyprefix, nil)

	if err != nil {
		qlog.Log(qlog.ERROR, "persist", "khistory", market, "fetch keys error", err)
		return err
	}

	var cachername = util.GetStr(profile, dict.CACHE_STOCK_KHISTORY, "cacher")
	var cacher *scache.SCache
	if len(cachername) > 0 {
		cacher = scache.GetManager().Get(cachername)
	}
	var rargs = make(map[string]interface{})
	for _, code := range codes {
		rargs["code"] = code
		rargs["begin"] = from_date_str
		rargs["end"] = to_date_str
		data, err := o.ShowAPI_request(work, rargs)
		if data == nil || err != nil {
			continue
		}
		var list = data.([]interface{})
		var listlen = len(list)
		if listlen == 0 {
			continue
		}
		var index = 0
		var dates = make([]interface{}, listlen)
		var infos = make([]interface{}, listlen)
		for _, info := range list {
			var minfo = info.(map[string]interface{})
			var datestr = util.AsStr(minfo["date"], "")
			//delete(minfo, "code");
			//delete(minfo, "market");
			//delete(minfo, "stockName");
			dates[index] = datestr
			infos[index] = minfo
			index = index + 1
			if cacher != nil {
				cacher.SetSubVal(true, minfo, code, datestr)
			}
		}
		var _, rerr = dao.Updates(dict.DB_HISTORY, code, dates, infos, true, -1, nil)
		if rerr != nil {
			qlog.Log(qlog.ERROR, "api", "showapi", "khistory", rerr)
		}
	}

	if err != nil {
		dao.Update(dict.DB_HISTORY, metatoken, "fetch_last_date", to_date_str, true, -1, nil)
	}

	return err
}
