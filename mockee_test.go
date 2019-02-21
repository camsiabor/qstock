package main

import (
	"fmt"
	"github.com/camsiabor/golua/luar"
	"github.com/camsiabor/qcom/global"
	"github.com/camsiabor/qcom/qdao"
	"github.com/camsiabor/qcom/qlog"
	"github.com/camsiabor/qcom/qref"
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qstock/dict"
	"github.com/camsiabor/qstock/run/rlua"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"runtime/pprof"
	"testing"
	"time"
)

// https://colobu.com/
// https://github.com/camsiabor/golua
// https://github.com/camsiabor/golua/luar/

func BenchmarkGolua(b *testing.B) {

}

func BenchmarkGoluaRaw(b *testing.B) {
	var cacheManager = scache.GetManager()
	var cache = cacheManager.Get("test")
	cache.Set("power overwhelming", "power")
	for i := 0; i < 100; i++ {
		var val, _ = cache.Get(false, "power")
		fmt.Println(val)
	}
}

func myQuery() {

	var scachem = scache.GetManager()
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
}

func myQuery2() {

	var scachem = scache.GetManager()
	var cache_code = scachem.Get(dict.CACHE_STOCK_CODE)
	var cache_stock = scachem.Get(dict.CACHE_STOCK_SNAPSHOT)
	var cache_khistory = scachem.Get(dict.CACHE_STOCK_KHISTORY)
	var r, _ = cache_code.Get(false, "all")
	var codes = util.AsStringSlice(r, 0)
	var list = map[string]interface{}{}
	var missing = map[string]interface{}{}

	var codes_len = len(codes)
	for i := 0; i < codes_len; i++ {
		var code = codes[i]
		r, _ = cache_stock.Get(true, code)
		var k0 = util.AsMap(r, false)
		var open0 = util.AsFloat64(k0["open"], 0)
		if k0 != nil && open0 != 0 {
			var r, _ = cache_khistory.GetSubVal(true, code, "20181214")
			var k1 = util.AsMap(r, false)
			if k1 != nil {
				var cr0 = util.AsFloat64(k0["change_rate"], 0)
				var cr1 = util.AsFloat64(k1["change_rate"], 0)
				var low0 = util.AsFloat64(k0["low"], 0)
				var high1 = util.AsFloat64(k1["high"], 0)
				//var pb = util.AsFloat64(k0["pb"], 0)
				if cr0 >= 0 && cr1 >= 0 && low0 > high1 {
					list[code] = code
				}
			} else {
				missing[code] = code
			}
		}
	}
	//fmt.Println(list)
}

func testCycle() {
	var g = global.GetInstance()
	g.CycleHandler = func(cycle string, g *global.G, data interface{}) {
		time.Sleep(time.Second)

		var count = 100
		var total float64 = 0
		for i := 0; i < count; i++ {
			var start = time.Now().UnixNano()
			myQuery2()
			var end = time.Now().UnixNano()
			if i > 0 {
				total = total + float64((end-start))/float64(time.Millisecond)
			}
		}
		total = total / float64(count-1)
		fmt.Println("finally avg ", total)
	}
}

func TestLuaBenchmark(t *testing.T) {

	var L, err = rlua.InitState()
	if L != nil {
		defer L.Close()
	}
	if err != nil {
		panic(err)
	}

	var rets, rerr = rlua.RunFile(L, "test.lua", nil)
	if rerr == nil {
		fmt.Println(rets)
	} else {
		fmt.Println(rerr.Error())
		panic(rerr)
	}

	var stackinfo = L.GetData("err_stack")
	if stackinfo != nil {
		fmt.Println(stackinfo)
	}

	fmt.Println("done")
	os.Exit(0)
}

func TestLuaBenchmark10(t *testing.T) {
	var g = global.GetInstance()
	g.CycleHandler = func(cycle string, g *global.G, x interface{}) {

		var lua_path = util.GetStr(g.Config, "./lua/?.lua", "lua", "lua_path")
		var lua_cpath = util.GetStr(g.Config, "lua/clib/?", "lua", "lua_cpath")

		var L = luar.Init()
		defer L.Close()
		L.OpenBase()
		L.OpenLibs()
		L.OpenTable()
		L.OpenString()

		if len(lua_path) > 0 {
			L.PushString(lua_path)
			L.SetGlobal("LUA_PATH")
		}

		if len(lua_cpath) > 0 {
			L.PushString(lua_cpath)
			L.SetGlobal("LUA_CPATH")
		}

		var err = L.LoadFileEx("test.lua")
		if err == nil {
			fmt.Println("done")
		} else {
			qlog.Log(qlog.ERROR, err)
		}

		os.Exit(0)
	}
	main()
}

func TestLuaBenchmark4(t *testing.T) {

	//Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8
	//Accept-Encoding: gzip, deflate, br
	//Accept-Language: zh,zh-TW;q=0.9,en-US;q=0.8,en;q=0.7,zh-CN;q=0.6
	//Cache-Control: max-age=0
	//Connection: keep-alive
	//Cookie: lastSeen=0
	//Host: www.booking.com
	//Upgrade-Insecure-Requests: 1
	//User-Agent: Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36

	var checkin = "2019-01-25"
	var checkout = "2019-01-26"
	var url = "https://www.booking.com/hotel/mo/galaxy-macau.en-gb.html?aid=397617;label=gog235jc-1FCAEoggI46AdIM1gDaJcBiAEBmAEJuAEXyAEM2AEB6AEB-AEMiAIBqAID;sid=2e5b23a4ac97dfeaa5a1087e5d1e15f3;all_sr_blocks=30141306_89159273_2_2_0;checkin=" + checkin + ";checkout=" + checkout + ";dest_id=-1204094;dest_type=city;dist=0;hapos=1;highlighted_blocks=30141306_89159273_2_2_0;hp_group_set=0;hpos=1;room1=A%2CA;sb_price_type=total;sr_order=popularity;srepoch=1547982423;srpvid=89754e2b105d0188;type=total;ucfs=1&"

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh,zh-TW;q=0.9,en-US;q=0.8,en;q=0.7,zh-CN;q=0.6")
	req.Header.Set("Cookie", "lastSeen=0")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Host", "www.booking.com")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36")
	// req.Header.Set("Cookie", "cors_js=1; _ga=GA1.2.1212520393.1532961488; zz_cook_tms_seg1=1; zz_cook_tms_ed=1; cto_lwid=90e954dd-e522-4417-8504-a8b682fca7ec; zz_cook_tms_ep=1; zz_cook_tms_em=1; zz_cook_tms_eg=1; esadm=02UmFuZG9tSVYkc2RlIyh9YbxZGyl9Y5%2BPCQ%2Be6L1iyuiQmlDq6ydyWKALPUDlLlmpVsAwmz%2FLoOU%3D; he=02UmFuZG9tSVYkc2RlIyh9YbxZGyl9Y5%2BPCQ%2Be6L1iyuiIqKIjP5uh%2F%2BUGiDKk0Q8iPbTZpbGgypc%3D; _gcl_au=1.1.895656110.1545647297; zz_cook_tms_seg3=7; _gid=GA1.2.2019762290.1547982083; header_joinapp_prompt_retargeting=1; vpmss=1; BJS=-; has_preloaded=1; 11_srd=%7B%22features%22%3A%5B%7B%22id%22%3A16%7D%5D%2C%22score%22%3A3%2C%22detected%22%3Afalse%7D; zz_cook_tms_hlist=301413; utag_main=v_id:0164eba007f40062ae11f441708003072004806a00978$_sn:32$_ss:0$_st:1547984362921$4split:3$4split2:1$ses_id:1547982084055%3Bexp-session$_pn:11%3Bexp-session; lastSeen=1547984659586; bkng=11UmFuZG9tSVYkc2RlIyh9YSvtNSM2ADX0BnR0tqAEmjsc8vSgxGCDYesZRvY29jItmxhxeLNutqxXn9%2F%2B4iCE4c2sZ9zdrG0w0o2O6bSqDap%2F0hyVpKPzwJ2o0Ttz5PeL2tRFJ6JpXYcG4AGSTr1HDsQY63sh8LuHO7Yl44l5xkNlb8whIeNaZh7TuxLw7l5booev1kS%2BN3HyFILFosNfoQ%3D%3D");
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	var html = string(body)
	// strings.Index()

	log.Println(html)

}

// go tool pprof http://localhost:8080/debug/pprof/profile
func TestLuaBenchmark3(t *testing.T) {
	var g = global.GetInstance()

	g.CycleHandler = func(cycle string, g *global.G, x interface{}) {
		var dao, _ = qdao.GetManager().Get(dict.DAO_MAIN)

		var file, _ = os.OpenFile("cpu_profile", os.O_RDWR|os.O_CREATE, 0644)
		defer file.Close()
		pprof.StartCPUProfile(file)
		var start = time.Now().Nanosecond()
		for i := 1; i <= 10000; i++ {
			var v, _ = dao.Get(dict.DB_HISTORY, "ch000001", "20190115", 0, nil)
			if v != nil && i%500 == 0 {
				fmt.Printf("%v\n", v)
			}
		}
		var end = time.Now().Nanosecond()
		fmt.Println((end - start) / int(time.Millisecond))
		pprof.StopCPUProfile()
	}
	main()
	time.Sleep(time.Hour)
}

func TestLuaBenchmark2(t *testing.T) {
	var g = global.GetInstance()

	g.CycleHandler = func(cycle string, g *global.G, x interface{}) {
		var data = g.Data()
		for k, one := range data {
			var v = reflect.ValueOf(one)
			if !v.IsValid() {
				continue
			}
			var vptr = reflect.ValueOf(&one)
			var t = reflect.TypeOf(one)

			var vptrenum = vptr.Elem()
			if vptrenum.CanSet() {
				vptrenum.Set(reflect.ValueOf(t.Name()))
			}
			data[k] = t.Name()

			fmt.Printf("%v = %v\n", k, t)
			var kind = t.Kind()
			switch kind {
			case reflect.Map:
				fmt.Println("map")
			case reflect.Ptr:
				var pointto = t.Elem()
				fmt.Printf("ptr --> %v\n", pointto)
			case reflect.Struct:
				fmt.Println("struct")
			}
		}
	}

	main()
	time.Sleep(time.Hour)
}

func t1() {
	var g = global.GetInstance()
	var inter interface{} = g
	var v = reflect.ValueOf(inter)
	fmt.Printf("%v\n", v.Type())
	var nv = reflect.New(v.Type())
	fmt.Printf("%v\n", nv.Type())
	var velem = v.Elem()
	fmt.Printf("%v\n", velem.Type())
	var nvelem = nv.Elem()
	fmt.Printf("%v\n", nvelem.Type())
	nvelem.Set(v)

	var ng = nvelem.Interface()
	fmt.Printf("%v", ng)

}

func t2() {
	var g = global.GetInstance()
	var slice = []interface{}{
		g,
	}
	var v = reflect.ValueOf(slice)
	var mirrorv, _ = qref.IterateMapSlice(v, true, func(value reflect.Value, pvalue reflect.Value) error {
		fmt.Printf("%v  | %v ", value.Type(), value)
		pvalue.Elem().Set(reflect.Zero(value.Type()))
		return nil
	})
	fmt.Println(mirrorv)
}

func t3() {
	var g = global.GetInstance()
	var gtype = reflect.TypeOf(*g)
	g.Mode = "power"
	var v = reflect.ValueOf(map[string]interface{}{
		"g": g, "gp": &g,
	})
	var nv, _ = qref.IterateMapSlice(v, true, func(value reflect.Value, pvalue reflect.Value) error {
		fmt.Printf("%v = %v | %v = %v", value.Type(), value, pvalue.Type(), pvalue)
		//pvalue.Elem().Set(reflect.Zero(value.Type()))
		if value.CanInterface() {
			if value.Type().ConvertibleTo(gtype) {
				value.FieldByName("Mode").SetString(" i am G")
			}
		}
		return nil
	})

	fmt.Printf("G: %v = %v\n", v.Type(), v)
	fmt.Printf("N: %v = %v\n", nv.Type(), nv)

	var m = v.Interface().(map[string]interface{})
	var nm = nv.Interface().(map[string]interface{})

	m["over"] = "here"
	nm["xer"] = "aa"
	fmt.Printf("G: %v\n", m)
	fmt.Printf("N: %v\n", nm)

	if nm["g"] != nil {
		var ng = nm["g"].(*global.G)
		fmt.Printf("N: %v\n", *ng)
	}

	fmt.Printf("Global %v\n", g)
}
func TestSimple(t *testing.T) {

	var inter []interface{} = make([]interface{}, 10)
	var v = reflect.ValueOf(inter)
	fmt.Println(v.Type())
	fmt.Println(v.Index(1).Type())

	//t3()
	//t1()

	/*

	 */
}

func TestMockee(t *testing.T) {
	const script = `
local v;
local cache = Q.cachem.Get("test");
for i = 1, 1 do
	v, _ = cache.Get(false, "power");
	print(v);
end
--return { name = v };
return {v, v, v}, { "2" };
`
	//var g = global.GetInstance();
	var cacheManager = scache.GetManager()
	var cache = cacheManager.Get("test")
	cache.Set("power overwhelming", "power")

	L := luar.Init()
	defer L.Close()

	luar.Register(L, "", luar.Map{
		// Go functions may be registered directly.
		"print": fmt.Println,
		"Q": map[string]interface{}{
			"cachem": cacheManager,
		},
	})

	var err2 error
	var err = L.DoString(script)
	if err == nil {
		var inter interface{}
		var inter2 interface{}
		err = luar.LuaToGo(L, 1, &inter)
		err2 = luar.LuaToGo(L, 2, &inter2)

		fmt.Println("[inter1 type]", L.Type(1))
		fmt.Println("[inter2 type]", L.Type(2))

		fmt.Println("[ret]", inter)
		fmt.Println("[ret2]", inter2)

	}
	if err != nil {
		fmt.Println("[err]", err.Error())
	}
	if err2 != nil {
		fmt.Println("[err2]", err2.Error())
	}
}
