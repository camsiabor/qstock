package test

import (
	"fmt"
	"github.com/camsiabor/golua/luar"
	"github.com/camsiabor/qcom/scache"
	"testing"
)

// https://colobu.com/
// https://github.com/camsiabor/golua
// https://github.com/camsiabor/golua/luar/

func BenchmarkGolua(b *testing.B) {
	TestTryLua(nil)
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

func testRegister() {
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

func TestTryLua(t *testing.T) {
	L := luar.Init()
	//L.LoadString("xxx")
	if L.Dump() == nil {
		var bytes = L.ToBytes(2)
		fmt.Println(len(bytes))
	}
	L.Close()
}
