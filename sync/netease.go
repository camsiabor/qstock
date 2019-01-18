package sync

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/camsiabor/qcom/qlog"
	"github.com/camsiabor/qcom/qtime"
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qstock/dict"
	"github.com/camsiabor/qstock/sync/showSdk/httplib"
	"github.com/camsiabor/qstock/sync/stock"
	"strings"
	"time"
)

// http://img1.money.126.net/data/hs/kline/day/history/2015/1399001.json
// http://img1.money.126.net/data/hk/kline/day/history/2018/00700.json

// http://blog.sina.com.cn/s/blog_afae4ee50102wu8a.html

func (o Syncer) Netease_request(
	work *ProfileWork,
	fields []string,
	requestargs map[string]interface{}) (data []interface{}, ids []interface{}, err error) {

	if !o.doContinue {
		return
	}
	var profile = work.Profile
	var api = util.GetStr(profile, "", "api")
	var url = o.domain + "/" + api
	var req = httplib.Post(url)
	var timeout = util.GetInt64(profile, 60, "timeout")
	var nice = util.GetInt64(profile, 200, "nice")
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
		return nil, nil, err
	}

	var retcode = util.GetInt(m, 0, "error_code")
	if retcode != 0 {
		var retmsg = util.GetStr(m, "", "reason")
		return nil, nil, errors.New(retmsg)
	}

	data = util.GetSlice(m, "data")
	return o.PersistAndCache(work, data)
}

func (o *Syncer) Netease_snapshot(
	phrase string, work *ProfileWork) (err error) {

	if phrase != "work" {
		return nil
	}

	var cache_khistory_name = util.GetStr(work.Profile, dict.CACHE_STOCK_KHISTORY, "cache_khistory")
	var cache_khistory = scache.GetManager().Get(cache_khistory_name)
	var now = time.Now()
	var hm = now.Hour()*100 + now.Minute()
	var todaystr = now.Format("20060102")
	var todaytrade = false
	if hm >= 925 {
		var cal = stock.GetStockCalendar()
		if cal.Is(todaystr) {
			todaytrade = true
		}
	}
	for i := 0; i < 2; i++ {
		data, _, err := o.ShenJian_request(work, nil, nil)
		if todaytrade {
			var datalen = len(data)
			for n := 0; n < datalen; n++ {
				var stock = data[n].(map[string]interface{})
				stock["date"] = todaystr
				var clone = util.MapCloneShallow(stock)
				var code = util.GetStr(stock, "", "code")
				cache_khistory.SetSubVal(true, clone, code, todaystr)
			}
		}
		if err == nil {
			qlog.Log(qlog.INFO, work.ProfileName, "persist", "success")
			break
		} else {
			qlog.Log(qlog.ERROR, work.ProfileName, "persist", "fail", err)
		}
	}
	return err
}

func (o *Syncer) Netease_khistory(phrase string, work *ProfileWork) (interface{}, error) {

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
