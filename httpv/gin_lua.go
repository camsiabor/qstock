package httpv

import (
	"github.com/camsiabor/golua/lua"
	"github.com/camsiabor/golua/luar"
	"github.com/camsiabor/qcom/global"
	"github.com/camsiabor/qcom/qref"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qstock/run"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
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

	var hash = util.GetStr(m, "", "hash")
	var script = util.GetStr(m, "", "script")
	if len(script) == 0 && len(hash) == 0 {
		o.RespJsonEx(0, errors.New("script is null && hash is null"), c)
		return
	}

	var debug = util.GetBool(m, false, "debug")

	var chunk []byte

	if len(hash) > 0 {
		var cache = o.GetData(hash)
		if cache == nil {
			func() {
				var L = luar.Init()
				defer L.Close()
				L.LoadString(script)
				if L.Dump() == nil {
					chunk = L.ToBytes(-1)
					o.SetData(hash, chunk)
				}
			}()
		} else {
			chunk = cache.([]byte)
		}
	}

	var L = luar.Init()
	defer L.Close()
	L.OpenLibs()
	var Q = global.GetInstance().Data()
	luar.Register(L, "Q", Q)

	var params = o.getScriptParams(m["params"])
	if params != nil {
		luar.Register(L, "A", params)
	}

	var start, end int64
	var consume float64

	var code = 0
	var data interface{}
	var goStackInfo map[string]interface{}

	var errhandler = func(L *lua.State, pan interface{}) {
		goStackInfo = qref.StackInfo(5)
		var stackstr = util.AsStr(goStackInfo["stack"], "")
		stackstr = strings.Replace(stackstr, "\t", "  ", -1)
		goStackInfo["stack"] = strings.Split(stackstr, "\n")
	}

	if debug {
		start = time.Now().UnixNano()
	}
	var err error
	if chunk == nil {
		err = L.DoStringHandle(script, errhandler)
	} else {
		err = L.LoadBuffer(chunk, "chunk", "")
		if err == nil {
			err = L.CallHandle(0, lua.LUA_MULTRET, errhandler)
		}
	}

	if debug {
		end = time.Now().UnixNano()
		consume = float64((end - start)) / float64(time.Millisecond)
	}

	if err == nil {
		data, err = run.LuaGetVal(L, 1)
		if err == nil {
			r2, err := run.LuaGetVal(L, 2)
			if err == nil && r2 != nil {
				code = util.AsInt(r2, 0)
			}
		}
	}

	if err == nil {
		if debug {
			var wrap = map[string]interface{}{}
			wrap["debug"] = true
			wrap["data"] = data
			wrap["consume"] = consume
			data = wrap
		}
	} else {
		code = 500
		var luaStacks = L.StackTrace()
		var luaStackInfo = run.LuaFormatStackToMap(luaStacks)
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

		data = r
	}
	o.RespJson(code, data, c)
}
