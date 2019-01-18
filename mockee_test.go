package main

import (
	"fmt"
	"github.com/camsiabor/golua/luar"
	"github.com/camsiabor/qcom/global"
	"github.com/camsiabor/qcom/qdao"
	"github.com/camsiabor/qcom/qref"
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qstock/dict"
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

// go tool pprof http://localhost:8080/debug/pprof/profile
func TestLuaBenchmark(t *testing.T) {
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
