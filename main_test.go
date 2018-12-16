package main

import (
	"fmt"
	"github.com/camsiabor/qcom/global"
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/util"
	"testing"
	"time"
)

func myQuery() {

	var scachem = scache.GetCacheManager()
	var cache_code = scachem.Get("stock.code")
	var cache_stock = scachem.Get("stock.snapshot")
	var r, _ = cache_code.Get(false, "all")
	var codes = util.AsStringSlice(r, 0)
	var list = map[string]interface{}{}

	var codes_len = len(codes)
	for i := 0; i < codes_len; i++ {
		var code = codes[i]
		r, _ = cache_stock.Get(true, code)
		var stock = util.AsMap(r, false)
		var open0 = util.AsFloat64(stock["open"], 0)
		if stock != nil && open0 != 0 {
			var pb = util.AsFloat64(stock["pb"], 0)
			var change_rate = util.AsFloat64(stock["change_rate"], 0)
			if pb >= 1 && pb <= 10 && change_rate <= 7 {
				list[code] = code
			}
		}
	}

	fmt.Println(list)

	/*
		for i = 1, #codes do

			local code = codes[i];
			local stock = cache_stock.Get(true, code);
			local open0 = stock["open"];
			if (stock ~= nil and open0 ~= nil and (open0 + 0) > 0 ) then
				local pb = stock["pb"] + 0;
				local change_rate = stock["change_rate"] + 0;
				if pb >= 1 and pb <= 10 and change_rate <= 7 then
					list[#list + 1] = code;
				end
			end

		end
		return list;
	*/
}

func TestMainTest(t *testing.T) {
	var g = global.GetInstance()
	g.CycleHandler = func(cycle string, g *global.G, data interface{}) {
		time.Sleep(time.Second * 5)

		for i := 0; i < 10; i++ {
			var start = time.Now().UnixNano()
			myQuery()
			var end = time.Now().UnixNano()
			fmt.Println("finally", i, "--->", float64((end-start))/float64(time.Millisecond))
		}

	}
	main()

	time.Sleep(time.Hour)
}
