package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/camsiabor/qcom/qlog"
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

func (o *Syncer) TuShare_trade_calendar(phrase string, work *ProfileWork) (err error) {
	if phrase != "work" {
		return nil
	}
	return nil
}

func (o *Syncer) TuShare_khistory(phrase string, work *ProfileWork) (interface{}, error) {

	var codes []string
	var metatoken string
	var date_to_str string
	var date_from_str string
	if work.GCmd != nil {
		var cmdata = work.GCmd.Data()

		date_to_str = util.GetStr(cmdata, "", "to")
		date_from_str = util.GetStr(cmdata, "", "from")
		if len(date_from_str) > 0 {
			codes = util.GetStringSlice(cmdata, "codes")
			if codes == nil || len(codes) == 0 {
				return nil, fmt.Errorf("codes is null %s : %v", work.GCmd.GetServFunc(), cmdata)
			}
		}

	}

	var dao = work.Dao
	var profile = work.Profile
	var profilename = work.ProfileName
	var db = util.GetStr(profile, dict.DB_HISTORY, "db")
	if len(date_to_str) == 0 {
		var metatoken = o.GetMetaToken(profilename)
		var fetcheach = util.GetInt(profile, 60, "each")
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
		var date_to_str = to_date.Format("20060102")
		if date_from_str == date_to_str {
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

	var keyprefix string
	var keysuffix string
	var market = util.GetStr(profile, "", "marker")
	market = strings.ToLower(market)
	if market == "sz" {
		keyprefix = "00*"
		keysuffix = ".SZ"
	} else if market == "ms" {
		keyprefix = "3*"
		keysuffix = ".SZ"
	} else {
		keyprefix = "60*"
		keysuffix = ".SH"
	}

	var err error
	if codes == nil || len(codes) == 0 {
		codes, err = dao.Keys(dict.DB_DEFAULT, "", keyprefix, nil)
		if err != nil {
			qlog.Log(qlog.ERROR, "persist", "khistory", market, "fetch keys error", err)
			return nil, err
		}
	}

	var data []interface{}
	var rargs = make(map[string]interface{})
	for retry := 1; retry <= 3; retry++ {
		var fails = make([]string, len(codes))
		var failcount = 0
		for _, code := range codes {
			if len(code) == 0 {
				continue
			}
			rargs["ts_code"] = code + keysuffix
			rargs["start_date"] = date_from_str
			rargs["end_date"] = date_to_str
			data_part, err := o.TuShare_request(work, nil, rargs)
			if data_part != nil && len(data_part) > 0 {
				if data == nil {
					data = data_part
				}
				data = append(data, data_part...)
			}
			if err == nil {
				qlog.Log(qlog.INFO, profilename, "persist", code, date_from_str, date_to_str)
			} else {
				qlog.Log(qlog.ERROR, profilename, "persist", "fail", code, date_from_str, date_to_str, err.Error())
				fails[failcount] = code
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
