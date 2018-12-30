package calendar

import (
	"github.com/camsiabor/qcom/global"
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qstock/dict"
	"sort"
	"sync"
	"time"
)

type StockCal struct {
	lock              sync.RWMutex
	todayDay          int
	lastTradeDayStr   string
	lastTradeDayIndex int
	dates             []string
	cache             *scache.SCache
}

var _stock = &StockCal{}

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

	var err error
	o.dates, err = o.cache.Keys()
	sort.Strings(o.dates)
	return err
}

func (o *StockCal) List(iprev int, pin int, inext int, reverse bool) []string {

	var now = time.Now()

	if o.dates == nil || len(o.dates) == 0 {
		if err := o.load(); err != nil {
			panic(err)
		}
	}

	// brutal reset
	if o.todayDay != now.Day() {
		func() {
			o.lock.Lock()
			defer o.lock.Unlock()
			if o.todayDay != now.Day() {
				o.load()
				o.todayDay = now.Day()
				for i := 0; i < 30; i++ {
					var lastTradeDay = now.AddDate(0, 0, -i)
					o.lastTradeDayStr = lastTradeDay.Format("20060102")
					if o.Is(o.lastTradeDayStr) {
						break
					}
				}

				var count = len(o.dates)
				for i := 0; i < count; i++ {
					if o.dates[i] == o.lastTradeDayStr {
						o.lastTradeDayIndex = i
						break
					}
				}
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
		for i := lower; i < upper; i++ {
			result[resultn] = o.dates[i]
			resultn = resultn + 1
		}
	}

	return result[:resultn]

}