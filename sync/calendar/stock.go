package calendar

import (
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qstock/dict"
	"sync"
	"time"
)

type StockCal struct {
	lock      sync.RWMutex
	pinday    int
	todaystr  string
	nextdelta int
	prevdelta int
	next      []string
	prev      []string
	cache     *scache.SCache
}

var _stock = &StockCal{
	next: make([]string, 1000),
	prev: make([]string, 1000),
}

func GetStockCalendar() *StockCal {
	return _stock
}

func (o *StockCal) List(iprev int, pin int, inext int, todayinclude bool) []string {

	var now = time.Now()

	// brutal reset
	if o.pinday != now.Day() {
		o.lock.Lock()
		if o.pinday != now.Day() {
			o.prevdelta = 0
			o.nextdelta = 0
			o.pinday = now.Day()
			o.todaystr = now.Format("20060102")
		}
		o.lock.Unlock()
	}

	if o.cache == nil {
		o.cache = scache.GetManager().Get(dict.CACHE_CALENDAR)
	}

	iprev = iprev - pin
	if iprev > o.prevdelta {
		var prev = time.Now()
		for i := 0; i < iprev; i++ {
			prev = prev.AddDate(0, 0, -1)
			var date = prev.Format("20060102")
			var isopen, _ = o.cache.Get(true, date)
			if util.AsInt(isopen, 0) > 0 {
				o.prev[i] = prev.Format("20060102")
			} else {
				o.prev[i] = ""
			}
		}
		o.prevdelta = iprev
	}

	inext = inext + pin
	if inext > o.nextdelta {
		var next = time.Now()
		for i := 0; i < inext; i++ {
			next = next.AddDate(0, 0, 1)
			var date = next.Format("20060102")
			var isopen, _ = o.cache.Get(true, date)
			if util.AsInt(isopen, 0) > 0 {
				o.next[i] = next.Format("20060102")
			} else {
				o.next[i] = ""
			}

		}
		o.nextdelta = inext
	}

	var rcount = 0
	var result = make([]string, inext+iprev+1)

	for i := iprev; i >= 0; i-- {
		var date = o.prev[i]
		if len(date) > 0 {
			result[rcount] = date
			rcount = rcount + 1
		}
	}

	if todayinclude {
		result[rcount] = o.todaystr
		rcount = rcount + 1
	}

	for i := 0; i < inext; i++ {
		var date = o.next[i]
		if len(date) > 0 {
			result[rcount] = date
			rcount = rcount + 1
		}
	}
	return result[:rcount]

}
