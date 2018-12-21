package httpv

import (
	"fmt"
	"github.com/camsiabor/golua/lua"
	"github.com/camsiabor/golua/luar"
	"github.com/camsiabor/qcom/global"
	"github.com/camsiabor/qcom/qref"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qstock/run/rlua"
	"github.com/camsiabor/qstock/run/rscript"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"reflect"
	"strings"
	"time"
)

func getL() *lua.State {
	var L = luar.Init()
	defer L.Close()
	L.OpenLibs()
	L.OpenDebug()
	L.OpenTable()
	var Q = global.GetInstance().Data()
	luar.Register(L, "Q", Q)
	return L
}

func (o *HttpServer) getScriptParams(params interface{}) map[string]interface{} {
	if params == nil {
		return nil
	}
	var list = util.GetSlice(params, "list")
	if list == nil {
		return nil
	}
	var r = make(map[string]interface{})
	for _, one := range list {
		var key = util.GetStr(one, "", "key")
		if len(key) == 0 {
			continue
		}
		var value = util.Get(one, "", "value")
		r[key] = value
	}
	return r
}

// TODO arguments
func (o *HttpServer) handleLuaCmd(cmd string, m map[string]interface{}, c *gin.Context) {

	var name = util.GetStr(m, "", "name")
	if len(name) == 0 {
		o.RespJsonEx(0, errors.New("script name is null"), c)
		return
	}

	var hash = util.GetStr(m, "", "hash")
	var script = util.GetStr(m, "", "script")
	if len(script) == 0 && len(hash) == 0 {
		o.RespJsonEx(0, errors.New("script is null && hash is null"), c)
		return
	}
	var mode = util.GetStr(m, "debug", "mode")
	var debug = mode == "debug"

	var meta *rscript.Meta
	if len(hash) > 0 {
		var cache, _ = cacheScriptByHash.Get(false, hash)
		if cache == nil {
			if len(script) == 0 {
				o.RespJson(404, "no script", c)
				return
			} else {
				meta = &rscript.Meta{}
				func() {
					var L = luar.Init()
					defer L.Close()
					L.LoadString(script)
					if L.Dump() == nil {
						meta.Binary = L.ToBytes(-1)
						meta.Name = name
						meta.Script = script
						meta.Lines = strings.Split(meta.Script, "\n")
						cacheScriptByHash.Set(meta, hash)
					}
				}()
			}
		} else {
			meta = cache.(*rscript.Meta)
		}
	}

	var L = luar.Init()
	defer L.Close()
	L.OpenLibs()
	var Q = global.GetInstance().Data()
	Q["mode"] = mode
	luar.Register(L, "Q", Q)

	var params = o.getScriptParams(m["params"])
	if params != nil && len(params) > 0 {
		luar.Register(L, "A", params)
	}

	var trace interface{}
	luar.Register(L, "", map[string]interface{}{
		"R": func(data interface{}) {
			var interrupt = &lua.Interrupt{
				Data: data,
			}
			panic(interrupt)
		},
		"Trace": func(data interface{}) {
			trace = data
		},
	})

	var start, end int64
	var consume float64

	var code = 0
	var data interface{}
	var interrupt *lua.Interrupt
	var goStackInfo map[string]interface{}

	var errhandler = func(L *lua.State, pan interface{}) {
		if pan == nil {
			return
		}
		var ok bool
		interrupt, ok = pan.(*lua.Interrupt)
		if !ok {
			goStackInfo = qref.StackInfo(5)
			var stackstr = util.AsStr(goStackInfo["stack"], "")
			stackstr = strings.Replace(stackstr, "\t", "  ", -1)
			goStackInfo["stack"] = strings.Split(stackstr, "\n")
		}
	}

	if debug {
		start = time.Now().UnixNano()
	}
	var err error
	if meta == nil || meta.Binary == nil {
		err = L.DoStringHandle(script, errhandler)
	} else {
		err = L.LoadBuffer(meta.Binary, "chunk", "")
		if err == nil {
			err = L.CallHandle(0, lua.LUA_MULTRET, errhandler)
		}
	}

	if debug {
		end = time.Now().UnixNano()
		consume = float64((end - start)) / float64(time.Millisecond)
	}

	if err == nil {
		if interrupt == nil {
			data, err = rlua.LuaGetVal(L, 1)
			if err == nil {
				r2, err := rlua.LuaGetVal(L, 2)
				if err == nil && r2 != nil {
					code = util.AsInt(r2, 0)
				}
			}
		} else {
			code = interrupt.Code
			data = interrupt.Data
		}
	}

	if err == nil {

		if data != nil {
			var vdata = reflect.ValueOf(data)
			vdata, _ = qref.IterateMapSlice(vdata, true, func(val reflect.Value, pval reflect.Value) (err error) {
				fmt.Printf("%v | %v\n", val.Type(), pval.Type())
				switch pval.Interface().(type) {
				case *interface{}:
					pval.Elem().Set(reflect.ValueOf("power"))
				}
				return nil
			})
			data = vdata.Interface()
		}

		if debug {
			var wrap = map[string]interface{}{}
			wrap["mode"] = mode
			wrap["data"] = data
			wrap["consume"] = consume
			wrap["params"] = params
			data = wrap
		}
	} else {
		code = 500
		var luaStacks = L.StackTrace()
		var luaStackInfo = rlua.LuaFormatStackToMap(luaStacks)
		var r = make(map[string]interface{})

		var luaStockLen = len(luaStackInfo)
		if len(luaStackInfo) > 0 {
			r["func"] = luaStackInfo[luaStockLen-1]["func"]
			r["line"] = luaStackInfo[luaStockLen-1]["line"]
			r["linesrc"] = luaStackInfo[luaStockLen-1]["linesrc"]
		}
		if goStackInfo != nil {
			luaStackInfo = append([]map[string]interface{}{goStackInfo}, luaStackInfo...)
		}

		r["stack"] = luaStackInfo
		r["err"] = err.Error()
		r["type"] = "lua"

		if debug {
			r["cosume"] = consume
		}

		if trace != nil {
			r["a.trace"] = trace
		}

		data = r
	}

	o.RespJson(code, data, c)
}
