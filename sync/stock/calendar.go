package stock

import (
	"fmt"
	"github.com/camsiabor/qcom/global"
	"github.com/camsiabor/qcom/qlog"
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qstock/dict"
	"sort"
	"strconv"
	"sync"
	"time"
)

type StockCal struct {
	lock sync.RWMutex

	openhm int

	todayDay       int
	todayStr       string
	todayNum       int
	todayIndex     int
	todayTrade     bool
	todayNeedReset bool

	thisWeekStr   string
	thisWeekNum   int
	thisWeekIndex int

	thisMonthStr   string
	thisMonthNum   int
	thisMonthIndex int

	lastTradeDayStr   string
	lastTradeDayIndex int
	datesn            []int
	dates             []string
	weeks             []string
	months            []string
	cache             *scache.SCache
}

var _stock = &StockCal{
	openhm: 925,
}

func GetStockCalendar() *StockCal {
	if _stock.cache == nil {
		_stock.cache = scache.GetManager().Get(dict.CACHE_CALENDAR)
	}
	return _stock
}

func (o *StockCal) Is(date string) bool {
	var one, _ = o.cache.Get(true, date)
	return one != nil && util.AsInt(one, 0) > 0
}

func (o *StockCal) load() error {
	var g = global.GetInstance()
	var cmd = &global.Cmd{
		Service:  "sync",
		Function: "trade.calendar",
		SFlag:    "force",
	}
	g.SendCmd(cmd, time.Second*15)

	if cmd.RetErr != nil {
		return cmd.RetErr
	}

	dates, err := o.cache.Keys(true)
	o.dates = util.AsStringSlice(dates, len(dates))
	sort.Strings(o.dates)

	var count = len(o.dates)
	o.datesn = make([]int, count)
	for i := 0; i < count; i++ {
		var date = o.dates[i]
		if len(date) > 0 {
			o.datesn[i], err = strconv.Atoi(o.dates[i])
			if err != nil {
				qlog.Log(qlog.ERROR, "date cannot convert to int", o.dates[i], err)
			}
		}
	}
	return err
}

func (o *StockCal) calInternal(hm int, now time.Time) {
	if hm >= o.openhm {
		o.todayNeedReset = false
	} else {
		o.todayNeedReset = true
		now = now.AddDate(0, 0, -1)
	}
	o.todayDay = now.Day()
	o.todayStr = now.Format("20060102")
	o.todayNum, _ = strconv.Atoi(o.todayStr)
	for i := 0; i < 30; i++ {
		var lastTradeDay = now.AddDate(0, 0, -i)
		o.lastTradeDayStr = lastTradeDay.Format("20060102")
		if o.Is(o.lastTradeDayStr) {
			break
		}
	}

	o.todayTrade = (o.todayStr == o.lastTradeDayStr)

	var count = len(o.dates)
	for i := 0; i < count; i++ {
		if o.dates[i] == o.lastTradeDayStr {
			o.lastTradeDayIndex = i
			break
		}
	}

	for i := 0; i < count; i++ {
		if o.dates[i] == o.todayStr {
			o.todayIndex = i
			break
		}
	}
}

func (o *StockCal) List(iprev int, pin int, inext int, reverse bool) []string {

	var now = time.Now()
	var hour = now.Hour()
	var minute = now.Minute()
	var hm = hour*100 + minute

	if o.dates == nil || len(o.dates) == 0 {
		if err := o.load(); err != nil {
			panic(err)
		}
	}

	var needReset bool
	if o.todayNeedReset {
		needReset = hm >= o.openhm
	} else {
		needReset = o.todayDay != now.Day()
	}
	// brutal reset
	if needReset {
		func() {
			o.lock.Lock()
			defer o.lock.Unlock()
			if o.todayDay != now.Day() {
				o.load()
				o.calInternal(hm, now)
			}
		}()
	}

	var resultn = 0
	var offset = o.lastTradeDayIndex + pin
	var lower = offset - iprev
	var upper = offset + inext
	if lower < 0 {
		lower = 0
	}
	if upper >= len(o.dates) {
		upper = len(o.dates) - 1
	}

	var result = make([]string, upper-lower+1)
	if reverse {
		for i := upper; i >= lower; i-- {
			result[resultn] = o.dates[i]
			resultn = resultn + 1
		}
	} else {
		for i := lower; i <= upper; i++ {
			result[resultn] = o.dates[i]
			resultn = resultn + 1
		}
	}

	return result[:resultn]

}

func (o *StockCal) ListByDate(from string, to string, reverse bool) ([]string, error) {

	to_num, err := strconv.Atoi(to)
	if err != nil {
		return nil, err
	}
	from_num, err := strconv.Atoi(from)
	if err != nil {
		return nil, err
	}

	if from_num > to_num {
		return nil, fmt.Errorf("from > to :  %d > %d", from_num, to_num)
	}

	var to_i int = -1
	var from_i int = -1
	for i := o.todayNum; i >= 0; i-- {
		var daten = o.datesn[i]
		if to_num == daten {
			to_i = i
		}

		if to_num == -1 && to_num > daten {
			to_i = i + 1
		}

		if from_num == daten {
			from_i = i
			break
		}
		if from_num > daten {
			from_i = i + 1
			break
		}
	}
	if from_i < 0 || to_i < 0 {
		return nil, fmt.Errorf("date not found: from %d to %d", from_num, to_num)
	}
	var count = 0
	var result = make([]string, from_i-to_i+1)
	if reverse {
		for i := to_i; i >= from_i; i-- {
			result[count] = o.dates[i]
			count++
		}
	} else {
		for i := from_i; i <= to_i; i++ {
			result[count] = o.dates[i]
			count++
		}
	}
	return result, nil
}

func (o *StockCal) ListWeek(iprev int, pin int, inext int, reverse bool) []string {
	panic("implement")
}

func (o *StockCal) ListMonth(iprev int, pin int, inext int, reverse bool) []string {
	panic("implement")
}
