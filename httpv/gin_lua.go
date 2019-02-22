package httpv

import (
	"fmt"
	"github.com/camsiabor/golua/lua"
	"github.com/camsiabor/golua/luar"
	"github.com/camsiabor/qcom/global"
	"github.com/camsiabor/qcom/qio"
	"github.com/camsiabor/qcom/qref"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qstock/run/rlua"
	"github.com/camsiabor/qstock/run/rscript"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
	"time"
)

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

func embedLuaScript(L *lua.State) int {
	var name = L.ToString(1)
	var data, err = cacheScriptByName.Get(true, name)
	if err != nil {
		panic(err)
	}
	if data == nil {
		L.PushString("no data for script " + name)
		return 1
	}
	var meta = data.(*rscript.Meta)
	if meta.Binary == nil {
		err = rlua.Compile(meta, cacheScriptByHash)
		if err != nil {
			panic(err)
		}
	}
	err = L.LoadBuffer(meta.Binary, meta.Name, "")
	if err != nil {
		panic(err)
	}
	L.MustCall(0, lua.LUA_MULTRET)
	//L.CallHandle(0, lua.LUA_MULTRET, nil)

	L.PushString("loaded " + name)
	return 1
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
				meta.Name = name
				meta.Hash = hash
				meta.Script = script
				rlua.Compile(meta, cacheScriptByHash)
			}
		} else {
			meta = cache.(*rscript.Meta)
		}
	}

	var L = luar.Init()
	defer L.Close()
	L.OpenBase()
	L.OpenLibs()
	L.OpenTable()
	L.OpenString()

	var Q = global.GetInstance().Data()
	Q["mode"] = mode
	luar.Register(L, "Q", Q)

	var params = o.getScriptParams(m["params"])
	if params != nil && len(params) > 0 {
		luar.Register(L, "A", params)
	}

	var trace interface{}
	luar.Register(L, "", map[string]interface{}{
		"Qr": func(data interface{}) {
			var interrupt = &lua.Interrupt{
				Data: data,
			}
			panic(interrupt)
		},
		"Qrace": func(data interface{}) {
			trace = data
		},
	})

	L.Register("Qembed", embedLuaScript)

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
		L.Notice, ok = pan.(*lua.Interrupt)
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
			data, err = rlua.GetVal(L, 1)
			if err == nil {
				r2, err := rlua.GetVal(L, 2)
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

		//if data != nil {
		//	var vdata = reflect.ValueOf(data)
		//	vdata, _ = qref.IterateMapSlice(vdata, true, func(val reflect.Value, pval reflect.Value) (err error) {
		//		//fmt.Printf("%v | %v\n", val.Type(), pval.Type())
		//		switch pval.Interface().(type) {
		//		case *interface{}:
		//			pval.Elem().Set(reflect.ValueOf("power"))
		//		}
		//		return nil
		//	})
		//	data = vdata.Interface()
		//}

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
		var luaStackInfo = rlua.FormatStackToMap(luaStacks)
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
			r["a.qtrace"] = trace
		}

		data = r
	}

	o.RespJson(code, data, c)
}

func (o *HttpServer) handleLuaFileCmd(cmd string, m map[string]interface{}, c *gin.Context) {

	var scriptname = util.GetStr(m, "", "path")
	if len(scriptname) == 0 {
		var err = fmt.Errorf("no script name in profile %v", m)
		o.RespJsonEx(nil, err, c)
		return
	}

	var stdout_redirect = util.GetBool(m, true, "stdout_redirect")
	var L, err = rlua.InitState()
	if L != nil {
		defer L.Close()
	}
	if err != nil {
		o.RespJsonEx(nil, err, c)
		return
	}

	var stdout *qio.Buffer
	if stdout_redirect {
		var stdout_type = util.GetStr(m, "memory", "stdout_type")
		var bufferType qio.BufferType
		if stdout_type == "memory" {
			bufferType = qio.BufferMemory
		} else {
			bufferType = qio.BufferTempFile
		}
		stdout = qio.NewBuffer(bufferType)
		defer stdout.Close()
	}

	if stdout != nil {
		L.SetStdout(stdout)
		L.SetDoCloseStdout(false)
	}

	var Q = global.GetInstance().Data()
	luar.Register(L, "Q", Q)
	var start = time.Now().UnixNano()
	var rets, rerr = rlua.RunFile(L, scriptname, nil)
	var end = time.Now().UnixNano()

	if rerr != nil {
		var luaerr, ok = rerr.(*lua.LuaError)
		if ok {
			var stacktrace = rlua.FormatStackToString(luaerr.StackTrace(), "\t", "")
			var m = make(map[string]interface{})
			m["code"] = luaerr.Code()
			m["error"] = luaerr.Error()
			m["stacktrace"] = stacktrace
			o.RespJson(luaerr.Code(), m, c)
		} else {
			o.RespJsonEx(nil, rerr, c)
		}
		return
	}

	var data interface{} = nil
	if rets != nil {
		if len(rets) >= 1 {
			data = rets[0]
		}
		if len(rets) >= 2 && rets[1] != nil {
			rerr = fmt.Errorf("%v", rets[1])
		}
	}

	var stdoutstr string
	if stdout != nil {
		bytes, _ := stdout.Bytes()
		stdoutstr = string(bytes[:])
	}

	var wrap = make(map[string]interface{})
	wrap["iamwrap"] = true
	wrap["data"] = data
	if rerr != nil {
		wrap["error"] = rerr.Error()
	}
	if len(stdoutstr) > 0 {
		wrap["stdout"] = stdoutstr
	}
	var consume = float64((end - start)) / float64(time.Millisecond)
	wrap["consume"] = consume
	o.RespJsonEx(wrap, nil, c)
}
