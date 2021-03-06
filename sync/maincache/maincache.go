package maincache

import (
	"fmt"
	"github.com/camsiabor/qcom/global"
	"github.com/camsiabor/qcom/qdao"
	"github.com/camsiabor/qcom/qerr"
	"github.com/camsiabor/qcom/qlog"
	"github.com/camsiabor/qcom/qtime"
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qstock/dict"
	"github.com/camsiabor/qstock/sync/stock"
	"strconv"
	"strings"
	"time"
)

func InitMainCache(g *global.G) {
	var cache_manager = scache.GetManager()
	g.SetData("cachem", cache_manager)

	var cache_trade_calendar = cache_manager.Get(dict.CACHE_CALENDAR)
	cache_trade_calendar.Dao = dict.DAO_MAIN
	cache_trade_calendar.Db = dict.DB_CALENDAR
	cache_trade_calendar.Loader = func(cache *scache.SCache, factor int, timeout time.Duration, lock bool, keys ...interface{}) (interface{}, error) {
		var dao, _ = qdao.GetManager().Get(cache.Dao)
		var date = keys[0]
		var r, err = dao.Get(cache.Db, cache.Group, date, 1, nil)
		var is_open = util.Get(r, 0, "is_open")
		return is_open, err
	}
	cache_trade_calendar.Initer = func(cache *scache.SCache, lock bool) (interface{}, error) {
		dao, _ := qdao.GetManager().Get(cache.Dao)
		dates, err := dao.Keys(cache.Db, cache.Group, "*", nil)
		if err == nil && dates != nil {
			var count = len(dates)
			for i := 0; i < count; i++ {
				var date = dates[i]
				var _, perr = strconv.Atoi(date)
				if perr == nil {
					cache.SetEx(1, dates[i], lock)
				}
			}
		}
		return dates, err
	}
	//cache_trade_calendar.Initer(cache_trade_calendar, true)

	var cache_timestamp = cache_manager.Get(dict.CACHE_TIMESTAMP)
	cache_timestamp.Loader = func(cache *scache.SCache, factor int, timeout time.Duration, lock bool, keys ...interface{}) (v interface{}, err error) {
		var key = keys[0]
		var skey = util.AsStr(key, "")
		if strings.Contains(skey, "@") {
			return qtime.Time2Int64(nil), nil
		}
		return time.Now().Format("20060102150405"), nil
	}

	var scache_code = cache_manager.Get(dict.CACHE_STOCK_CODE)
	scache_code.ArrayLimitInit = 1000000
	scache_code.Dao = dict.DAO_MAIN
	scache_code.Db = dict.DB_DEFAULT
	scache_code.Loader = func(cache *scache.SCache, factor int, timeout time.Duration, lock bool, keys ...interface{}) (interface{}, error) {
		var dao, err = qdao.GetManager().Get(cache.Dao)
		if err != nil {
			qlog.Error(0, err)
			go func() {
				defer qerr.SimpleRecover(1)
				time.Sleep(time.Duration(10) * time.Second)
				cache.Loader(cache, factor, timeout, lock, keys...)
			}()
			return nil, err
		}
		var codes, _ = dao.Keys(cache.Db, "", "*", nil)
		var sz, szn = make([]string, 5000), 0
		var sh, shn = make([]string, 5000), 0
		var su, sun = make([]string, 5000), 0
		var ch, chn = make([]string, 15000), 0
		var hk, hkn = make([]string, 10000), 0
		for _, code := range codes {
			var include = true

			var first = code[0]
			var third = code[2]

			if first == 'c' {
				switch third {
				case '0':
					sz[szn] = code
					szn++
				case '3':
					su[szn] = code
					sun++
				case '6':
					sh[shn] = code
					shn++
				default:
					include = false
				}
				if include {
					ch[chn] = code
					chn++
				}
			} else if first == 'h' {
				hk[hkn] = code
				hkn++
			}
		}

		var sz_sh = make([]string, 0)
		sz_sh = append(sz_sh, sz[:szn]...)
		sz_sh = append(sz_sh, sh[:shn]...)

		cache.SetEx(ch[:chn], dict.CHINA, lock)
		cache.SetEx(sz[:szn], dict.SHENZHEN, lock)
		cache.SetEx(sh[:shn], dict.SHANGHAI, lock)
		cache.SetEx(su[:sun], dict.STARTUP, lock)
		cache.SetEx(sz_sh, dict.SHENZHEN+"."+dict.SHANGHAI, lock)
		cache.SetEx(hk[:hkn], dict.HONGKONG, lock)
		qlog.Log(qlog.INFO, "cache", "code", "shenzhen", szn, "shanghai", shn, "startup", sun, "china", chn, "hongkong", hkn)
		return cache, nil
	}
	var apim = util.GetMap(g.Config, false, "api")
	if apim != nil || len(apim) > 0 {
		scache_code.Loader(scache_code, 1, 0, true)
	}

	var scache_snapshot = scache.GetManager().Get(dict.CACHE_STOCK_SNAPSHOT)
	scache_snapshot.ArrayLimitInit = 1000000
	scache_snapshot.Dao = dict.DAO_MAIN
	scache_snapshot.Db = dict.DB_DEFAULT
	scache_snapshot.Loader = func(cache *scache.SCache, factor int, timeout time.Duration, lock bool, keys ...interface{}) (interface{}, error) {
		conn, err := qdao.GetManager().Get(cache.Dao)
		if err != nil {
			return nil, err
		}
		var code = keys[0]
		data, err := conn.Get(cache.Db, "", code, 1, nil)
		if data != nil {
			data = util.MapStringToFloat64(data)
		}
		return data, err
	}

	var cache_khistory_loader_generator = func(prefix string, profilesuffix string) scache.Loader {
		return func(cache *scache.SCache, factor int, timeout time.Duration, lock bool, keys ...interface{}) (interface{}, error) {
			if len(keys) < 1 {
				return nil, fmt.Errorf("keys len invalid for this cache Loader %s", cache.Name)
			}
			conn, err := qdao.GetManager().Get(cache.Dao)
			if err != nil {
				return nil, err
			}

			var code string
			if len(prefix) == 0 {
				code = util.AsStr(keys[0], "")
			} else {
				var skey = util.AsStr(keys[0], "")
				code = prefix + skey
			}
			var datestr = keys[1]
			data, err := conn.Get(cache.Db, code, datestr, 1, nil)

			if data != nil {
				data = util.MapStringToFloat64(data)
				return data, err
			}

			if factor <= 0 {
				return data, err
			}

			var g = global.GetInstance()
			var cmd = &global.Cmd{
				Service:  dict.SERVICE_SYNC,
				Function: "k.history" + profilesuffix,
				SFlag:    "go",
			}
			cmd.SetData("from", datestr)
			cmd.SetData("codes", []string{code})
			var reply, _ = g.SendCmd(cmd, timeout)
			if reply != nil {
				err = reply.RetErr
				data = reply.RetVal
			}
			return data, err
		}
	}

	var cache_stock_khistory = scache.GetManager().Get(dict.CACHE_STOCK_KHISTORY)
	cache_stock_khistory.ArrayLimitInit = 1000000
	cache_stock_khistory.Dao = dict.DAO_MAIN
	cache_stock_khistory.Db = dict.DB_HISTORY
	cache_stock_khistory.Timeout = -1 //time.Second * time.Duration(20);
	cache_stock_khistory.Loader = cache_khistory_loader_generator("", "")

	var cache_stock_khistory_week = scache.GetManager().Get(dict.CACHE_STOCK_KHISTORY_WEEK)
	cache_stock_khistory_week.ArrayLimitInit = 1000000
	cache_stock_khistory_week.Dao = dict.DAO_MAIN
	cache_stock_khistory_week.Db = dict.DB_HISTORY
	cache_stock_khistory_week.Timeout = -1 //time.Second * time.Duration(20);
	cache_stock_khistory_week.Loader = cache_khistory_loader_generator("w.", "week")

	var cache_stock_khistory_month = scache.GetManager().Get(dict.CACHE_STOCK_KHISTORY_MONTH)
	cache_stock_khistory_month.ArrayLimitInit = 1000000
	cache_stock_khistory_month.Dao = dict.DAO_MAIN
	cache_stock_khistory_month.Db = dict.DB_HISTORY
	cache_stock_khistory_month.Timeout = -1 //time.Second * time.Duration(20);
	cache_stock_khistory_month.Loader = cache_khistory_loader_generator("m.", "month")

	var cache_stock_khistory_flat = scache.GetManager().Get(dict.CACHE_STOCK_KHISTORY_FLAT)
	cache_stock_khistory_flat.Dao = dict.DAO_MAIN
	cache_stock_khistory_flat.Db = dict.DB_HISTORY
	cache_stock_khistory_flat.Loader = func(cache *scache.SCache, factor int, timeout time.Duration, lock bool, keys ...interface{}) (interface{}, error) {
		var mapping = make(map[string]interface{})
		var codesInterf, _ = scache_code.Get(true, "sz.sh")
		var codes = codesInterf.([]string)
		for i, n := 0, len(codes); i < n; i++ {
			// var code = codes[i]
		}
		cache_stock_khistory.Get(true, "")
		return mapping, nil
	}

	g.SetData("calendar", stock.GetStockCalendar())
}
