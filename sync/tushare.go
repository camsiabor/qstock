package sync

import (
	"bytes"
	"encoding/json"
	"github.com/camsiabor/qcom/qlog"
	"github.com/camsiabor/qcom/qtime"
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qstock/dict"
	"github.com/camsiabor/qstock/sync/showSdk/httplib"
	"github.com/pkg/errors"
	"strings"
	"time"
)

// https://tushare.pro/document/2?doc_id=123
/*
curl -X POST -d '{"api_name": "trade_cal", "token": "xxxxxxxx", "params": {"exchange":"", "start_date":"20180901", "end_date":"20181001", "is_open":"0"}, "fields": "exchange,cal_date,is_open,pretrade_date"}' http://api.tushare.pro
*/

func (o Syncer) TuShare_request(
	work *ProfileWork,
	fields []string,
	requestargs map[string]interface{}) (ret []interface{}, err error) {

	if !o.doContinue {
		return
	}
	var profile = work.Profile
	var req = httplib.Post(o.domain)
	var reqm = make(map[string]interface{})
	var api = util.GetStr(profile, "", "api")

	reqm["token"] = o.appsecret
	reqm["api_name"] = api
	if fields != nil && len(fields) > 0 {
		reqm["fields"] = fields
	}
	reqm["params"] = requestargs
	reqbody, err := json.Marshal(reqm)
	if err != nil {
		return
	}
	req.Body(reqbody)
	var timeout = util.GetInt64(profile, 20, "timeout")
	var nice = util.GetInt64(profile, 250, "nice")
	req.SetTimeout(time.Duration(10)*time.Second, time.Duration(timeout)*time.Second)
	if nice > 0 {
		time.Sleep(time.Millisecond * time.Duration(nice))
	}
	httpresp, err := req.DoRequest()
	if err != nil {
		return
	}
	var m map[string]interface{}
	var buffer = new(bytes.Buffer)
	buffer.ReadFrom(httpresp.Body)
	err = json.Unmarshal(buffer.Bytes(), &m)
	if err != nil {
		return nil, err
	}

	var retcode = util.GetInt(m, 0, "code")
	if retcode != 0 {
		var retmsg = util.GetStr(m, "", "msg")
		return nil, errors.New(retmsg)
	}

	var data = util.GetMap(m, false, "data")
	var cols = util.GetStringSlice(data, "fields")
	var rows = util.GetSlice(data, "items")
	var datalen = len(rows)
	if datalen <= 0 {
		return nil, nil
	}
	maps, err := util.ColRowToMaps(cols, rows)
	if err != nil {
		return nil, err
	}
	_, _, err = o.PersistAndCache(work, maps)
	return maps, err
}

func (o *Syncer) TuShare_trade_calendar(phrase string, work *ProfileWork) error {
	if phrase != "work" {
		return nil
	}

	var err error
	var calendar []interface{}
	var each = util.GetInt(work.Profile, 365, "each")
	var start_date = time.Now().AddDate(0, 0, -each).Format("20060102")
	var end_date = time.Now().AddDate(0, 0, +each).Format("20060102")
	var rargs = make(map[string]interface{})
	rargs["exchange"] = "SSE"
	rargs["start_date"] = start_date
	rargs["end_date"] = end_date
	rargs["is_open"] = 1
	retry := util.GetInt(work.Profile, 3, "retry")
	var cacher = scache.GetManager().Get(dict.CACHE_CALENDAR)
	for i := 1; i <= retry; i++ {
		calendar, err = o.TuShare_request(work, nil, rargs)
		if err == nil {
			var list = util.AsSlice(calendar, 0)
			var dates = make([]string, len(list))
			var is_opens = make([]interface{}, len(list))
			for i, one := range list {
				dates[i] = util.GetStr(one, "", "date") // after mapping cal_date -> date
				is_opens[i] = util.GetInt(one, 0, "is_open")
			}
			cacher.Sets(is_opens, dates)
			break
		}
	}
	return err
}

func (o *Syncer) TuShare_khistory(phrase string, work *ProfileWork) (interface{}, error) {

	var date_to_str string
	var date_from_str string
	var codes []string = util.GetStringSlice(work.Profile, "codes")
	var addsuffix = util.GetBool(work.Profile, true, "addsuffix")
	if work.GCmd != nil {
		var cmdata = work.GCmd.Data()
		date_to_str = util.GetStr(cmdata, "", "to")
		date_from_str = util.GetStr(cmdata, "", "from")

		var cmdcodes = util.GetStringSlice(cmdata, "codes")
		if cmdcodes != nil {
			codes = cmdcodes
		}
	}

	var api = util.AsStr(work.Profile["api"], "")
	var dao = work.Dao
	var profile = work.Profile
	var profilename = work.ProfileName
	var metatoken = o.GetMetaToken(profilename)
	var db = util.GetStr(profile, dict.DB_HISTORY, "db")
	if len(date_to_str) == 0 {

		var fetcheach_default int
		if api == "daily" {
			fetcheach_default = 90
		} else {
			fetcheach_default = 180
		}

		var fetcheach = util.GetInt(profile, fetcheach_default, "each")
		date_from_str = time.Now().AddDate(0, 0, -fetcheach).Format("20060102")
		var fetch_last_date_from, _ = util.AsStrErr(dao.Get(db, metatoken, "fetch_last_date_from", 0, nil))
		if len(fetch_last_date_from) > 0 {
			var date_from_str_num = util.AsInt64(date_from_str, 0)
			var fetch_last_date_from_num = util.AsInt64(fetch_last_date_from, 0)
			if date_from_str_num >= fetch_last_date_from_num {
				var fetch_last_date, _ = util.AsStrErr(dao.Get(db, metatoken, "fetch_last_date", 0, nil))
				if len(fetch_last_date) > 0 {
					date_from_str = fetch_last_date
				}
			}
		}

		var to_date = time.Now()
		date_to_str = to_date.Format("20060102")
		if date_from_str == date_to_str {
			if work.GCmd == nil || !strings.Contains(work.GCmd.SFlag, "force") {
				if to_date.Weekday() == time.Saturday || to_date.Weekday() <= time.Sunday {
					qlog.Log(qlog.DEBUG, o.Name, "sunday & saturday need a rest")
					return nil, nil
				}
				if to_date.Hour() < 15 {
					qlog.Log(qlog.DEBUG, o.Name, "wait for the market to rest ", to_date.Hour())
					return nil, nil
				}
			}
		}
	}

	if len(date_to_str) == 0 {
		date_to_str = time.Now().Format("20060102")
	}

	var err error
	var targets []string
	var fetch_by_date bool = (codes == nil || len(codes) == 0)
	if fetch_by_date {
		var date_to, _ = time.Parse("20060102", date_to_str)
		var date_from, _ = time.Parse("20060102", date_from_str)
		if api == "weekly" {
			targets, err = qtime.GetTimeFormatIntervalArray(&date_from, &date_to, "20060102",
				false, time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Sunday, time.Saturday)
		} else if api == "monthly" {
			targets, err = qtime.GetTimeFormatIntervalArray(&date_from, &date_to, "20060102",
				true)
		} else {
			targets, err = qtime.GetTimeFormatIntervalArray(&date_from, &date_to, "20060102",
				false, time.Sunday, time.Saturday)
		}

	} else {
		targets = codes
	}

	var data []interface{}
	var data_part []interface{}
	var rargs = make(map[string]interface{})
	var retry = util.GetInt(work.Profile, 3, "retry")

	for i := 1; i <= retry; i++ {
		var fails = make([]string, len(codes))
		var failcount = 0
		for _, target := range targets {
			if len(target) == 0 {
				continue
			}
			if fetch_by_date {
				rargs["trade_date"] = target
			} else {
				var keysuffix string
				if addsuffix {
					if target[0] == '6' {
						keysuffix = ".SH"
					} else {
						keysuffix = ".SZ"
					}
				}
				rargs["ts_code"] = target + keysuffix
				rargs["start_date"] = date_from_str
				rargs["end_date"] = date_to_str
			}
			data_part, err = o.TuShare_request(work, nil, rargs)
			if data_part != nil && len(data_part) > 0 {
				if data == nil {
					data = data_part
				}
				data = append(data, data_part...)
			}
			if err == nil {
				qlog.Log(qlog.INFO, profilename, "persist", target, date_from_str, date_to_str)
			} else {
				qlog.Log(qlog.ERROR, profilename, "persist", "fail", target, date_from_str, date_to_str, err.Error())
				fails[failcount] = target
				failcount++
			}
		}
		if failcount > 0 {
			codes = fails
			qlog.Log(qlog.ERROR, profilename, "persist", "failcount", failcount, "retry", retry)
		} else {
			if len(metatoken) > 0 {
				dao.Update(db, metatoken, "fetch_last_date", date_to_str, true, 0, nil)
				dao.Update(db, metatoken, "fetch_last_date_from", date_from_str, true, 0, nil)
			}
			break
		}
	}

	return data, err
}

func (o *Syncer) TuShare_trade_concept(phrase string, work *ProfileWork) (interface{}, error) {
	return nil, nil
}
