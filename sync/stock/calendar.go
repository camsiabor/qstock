package stock

import (
	"fmt"
	"github.com/camsiabor/qcom/global"
	"github.com/camsiabor/qcom/qlog"
	"github.com/camsiabor/qcom/qroutine"
	"github.com/camsiabor/qcom/qtime"
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qstock/dict"
	"sort"
	"strconv"
	"sync"
	"time"
)

type StockCalInterval int

const (
	Day StockCalInterval = iota
	Week
	Month
)

type StockCal struct {
	lock sync.RWMutex

	openhm int

	today          time.Time
	todayDay       int
	todayStr       string
	todayNum       int
	todayIndex     int
	todayTrade     bool
	todayNeedReset bool

	weekStr   string
	weekNum   int
	weekIndex int

	monthStr   string
	monthNum   int
	monthIndex int

	lastTradeDay      time.Time
	lastTradeDayNum   int
	lastTradeDayStr   string
	lastTradeDayIndex int

	dates  []string
	datesn []int

	weeks  []string
	weeksn []int

	months  []string
	monthsn []int

	cache *scache.SCache
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

func (o *StockCal) IsByInt(date int) bool {
	return o.Is(strconv.Itoa(date))
}

func (o *StockCal) load() error {
	var g = global.GetInstance()
	var cmd = &global.Cmd{
		Service:  "sync",
		Function: "trade.calendar",
		SFlag:    "force",
	}
	g.SendCmd(cmd, time.Second*5)

	if cmd.RetErr != nil {
		return cmd.RetErr
	}

	o.cache.Init(true)
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
	qlog.Log(qlog.INFO, "loaded", len(o.dates))
	return err
}

func (o *StockCal) calDay(hm int, now time.Time) {

	if hm >= o.openhm {
		o.todayNeedReset = false
	} else {
		o.todayNeedReset = true
		now = now.AddDate(0, 0, -1)
	}
	o.today = now
	o.todayDay = now.Day()
	o.todayStr = now.Format("20060102")
	o.todayNum, _ = strconv.Atoi(o.todayStr)
	for i := 0; i < 30; i++ {
		o.lastTradeDay = now.AddDate(0, 0, -i)
		o.lastTradeDayStr = o.lastTradeDay.Format("20060102")
		if o.Is(o.lastTradeDayStr) {
			o.lastTradeDayNum, _ = strconv.Atoi(o.lastTradeDayStr)
			qtime.TruncateHMS(&o.lastTradeDay)
			qlog.Log(qlog.INFO, "last trade day", o.lastTradeDay.Format("2006-01-02 15:04:05"))
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

func (o *StockCal) calWeek() {
	var datecount = len(o.dates)
	var date_end string
	var date_start string
	var time_end time.Time
	var time_start time.Time
	for i := datecount - 1; i >= 0; i-- {
		date_end = o.dates[i]
		if date_end != "" {
			var err error
			time_end, err = time.Parse("20060102", date_end)
			if err == nil {
				if time_end.Weekday() == time.Friday {
					break
				}
			} else {
				qlog.Log(qlog.ERROR, "parse time error", date_end, err)
			}

		}
	}
	for i := 0; i < datecount; i++ {
		if o.dates[i] != "" {
			var err error
			date_start = o.dates[i]
			time_start, err = time.Parse("20060102", date_start)
			if err == nil {
				break
			} else {
				qlog.Log(qlog.ERROR, "parse time error", date_start, err)
			}

		}
	}

	var last_trade_unix = o.lastTradeDay.Unix()
	var time_end_unix = time_end.Unix()
	var time_start_unix = time_start.Unix()

	var count = 0
	var date string
	var thisweekindex = -1
	var capacity = datecount / 3
	var weeks = make([]string, capacity)
	var weeksn = make([]int, capacity)
	for time_end_unix >= time_start_unix {
		var delta = -1
		var found = false

		var weekday = time_end.Weekday()
		if weekday >= time.Monday && weekday <= time.Friday {
			date = time_end.Format("20060102")
			if o.Is(date) {
				var daten, _ = strconv.Atoi(date)
				found = true
				count = count + 1
				weeks[capacity-count] = date
				weeksn[capacity-count] = daten
				delta = -int(weekday) - 2
			}
		}

		var before = time_end_unix
		time_end = time_end.AddDate(0, 0, delta)
		time_end_unix = time_end.Unix()
		if thisweekindex < 0 && found {
			if last_trade_unix >= time_end_unix && last_trade_unix <= before {
				o.weekStr = date
				o.weekNum, _ = strconv.Atoi(date)
				thisweekindex = count
			}
		}
	}

	o.weeks = weeks[capacity-count:]
	o.weeksn = weeksn[capacity-count:]
	o.weekIndex = len(o.weeks) - thisweekindex

	qlog.Log(qlog.INFO, o.weekStr, "week index", o.weekIndex)
}

func (o *StockCal) calMonth() {
	var datecount = len(o.datesn)
	var capacity = datecount / 10
	var months = make([]string, capacity)
	var monthsn = make([]int, capacity)
	var count = 0
	var i = datecount - 1
	var prevmonth = 0

	var thisyear = o.lastTradeDay.Year()
	var thismonth = int(o.lastTradeDay.Month())
	var thismonthindex = -1

	for i >= 0 {
		var daten = o.datesn[i]
		var day = daten % 100
		if day >= 28 {
			var month = (daten % 10000) / 100
			if month != prevmonth {

				var date = o.dates[i]
				prevmonth = month
				count = count + 1
				months[capacity-count] = date
				monthsn[capacity-count] = daten

				if thismonthindex < 0 {
					var year = daten / 10000
					if month == thismonth && year == thisyear {
						o.monthNum = daten
						o.monthStr = date
						thismonthindex = count
					}
				}
			}
		}
		i--
	}
	o.months = months[capacity-count:]
	o.monthsn = monthsn[capacity-count:]
	o.monthIndex = len(o.months) - thismonthindex

	qlog.Log(qlog.INFO, o.monthStr, "month index", o.monthIndex)
}

func (o *StockCal) check() {
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
				o.calDay(hm, now)
				qroutine.Exec(
					0,
					qroutine.NewBox(func(arg interface{}) interface{} {
						o.calWeek()
						return nil
					}, nil),
					qroutine.NewBox(func(arg interface{}) interface{} {
						o.calMonth()
						return nil
					}, nil),
				)
			}
		}()
	}
}

func (o *StockCal) ListEx(interval StockCalInterval, iprev int, pin int, inext int, reverse bool) []string {

	o.check()

	var resultn = 0

	var offset int
	var domain []string
	switch interval {
	case Day:
		domain = o.dates
		offset = o.lastTradeDayIndex + pin
	case Week:
		domain = o.weeks
		offset = o.weekIndex + pin
	case Month:
		domain = o.months
		offset = o.monthIndex + pin
	}
	var lower = offset - iprev
	var upper = offset + inext
	if lower < 0 {
		lower = 0
	}
	if upper >= len(domain) {
		upper = len(domain) - 1
	}

	var result = make([]string, upper-lower+1)
	if reverse {
		for i := upper; i >= lower; i-- {
			result[resultn] = domain[i]
			resultn = resultn + 1
		}
	} else {
		for i := lower; i <= upper; i++ {
			result[resultn] = domain[i]
			resultn = resultn + 1
		}
	}

	return result[:resultn]
}

func (o *StockCal) List(iprev int, pin int, inext int, reverse bool) []string {
	return o.ListEx(Day, iprev, pin, inext, reverse)
}

func (o *StockCal) ListWeek(iprev int, pin int, inext int, reverse bool) []string {
	return o.ListEx(Week, iprev, pin, inext, reverse)
}

func (o *StockCal) ListMonth(iprev int, pin int, inext int, reverse bool) []string {
	return o.ListEx(Month, iprev, pin, inext, reverse)
}

func (o *StockCal) ListByDateEx(interval StockCalInterval, from string, to string, reverse bool) ([]string, error) {

	o.check()

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

	var current int
	var targets []string
	var targetsn []int
	switch interval {
	case Day:
		targets = o.dates
		targetsn = o.datesn
		current = o.lastTradeDayIndex
	case Week:
		targets = o.weeks
		targetsn = o.weeksn
		current = o.weekIndex
	case Month:
		targets = o.months
		targetsn = o.monthsn
		current = o.monthIndex
	}

	var to_i int = -1
	var from_i int = -1
	for i := current; i >= 0; i-- {
		var daten = targetsn[i]
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

	if from_i < 0 {
		from_i = 0
	}

	if to_i < 0 {
		to_i = current
	}
	var count = 0
	var result = make([]string, to_i-from_i+1)
	if reverse {
		for i := to_i; i >= from_i; i-- {
			result[count] = targets[i]
			count++
		}
	} else {
		for i := from_i; i <= to_i; i++ {
			result[count] = targets[i]
			count++
		}
	}
	return result, nil
}

func (o *StockCal) ListByDate(from string, to string, reverse bool) ([]string, error) {
	return o.ListByDateEx(Day, from, to, reverse)
}

func (o *StockCal) ListWeekByDate(from string, to string, reverse bool) ([]string, error) {
	return o.ListByDateEx(Week, from, to, reverse)
}

func (o *StockCal) ListMonthByDate(from string, to string, reverse bool) ([]string, error) {
	return o.ListByDateEx(Month, from, to, reverse)
}
