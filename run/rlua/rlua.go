package rlua

import (
	"fmt"
	"github.com/camsiabor/golua/lua"
	"github.com/camsiabor/golua/luar"
	"github.com/camsiabor/qcom/global"
	"github.com/camsiabor/qcom/qconfig"
	"github.com/camsiabor/qcom/qencode"
	"github.com/camsiabor/qcom/qlog"
	"github.com/camsiabor/qcom/qref"
	"github.com/camsiabor/qcom/scache"
	"github.com/camsiabor/qcom/util"
	"github.com/camsiabor/qstock/run/rscript"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var _mutex sync.Mutex
var _LUA_PATH string
var _LUA_PATH_FULL string
var _LUA_CPATH string
var _LUA_CPATH_FULL string

var _TOKEN_GLOBAL_MODULE = "__q_global"

func TokenGlobalModule() string {
	return _TOKEN_GLOBAL_MODULE
}

func GetVal(L *lua.State, idx int) (interface{}, error) {
	if L.IsNoneOrNil(idx) {
		return nil, nil
	}
	var ltype = int(L.Type(idx))
	switch ltype {
	case int(lua.LUA_TNUMBER):
		return L.ToNumber(idx), nil
	case int(lua.LUA_TSTRING):
		return L.ToString(idx), nil
	case int(lua.LUA_TBOOLEAN):
		return L.ToBoolean(idx), nil
	}
	var r interface{}
	var err = luar.LuaToGo(L, idx, &r)
	return r, err
}

func FormatStack(stacks []lua.LuaStackEntry) []lua.LuaStackEntry {
	var count = len(stacks)
	var clones = make([]lua.LuaStackEntry, count)
	for i := 0; i < count; i++ {
		var stack = stacks[i]
		var clone = lua.LuaStackEntry{
			Name: stack.Name,
		}
		var linenum = stack.CurrentLine
		if linenum >= 0 {
			var lines = strings.Split(stack.Source, "\n")
			if linenum < len(lines) {
				clone.ShortSource = lines[linenum-1]
			} else {
				clone.ShortSource = stack.ShortSource
			}
		}
		clone.Source = ""
		clone.CurrentLine = linenum
		clones[i] = clone
	}
	return clones
}

func FormatStackToString(stacks []lua.LuaStackEntry, prefix string, suffix string) string {
	var str = ""
	var count = len(stacks)

	for i := 0; i < count; i++ {
		var stack = stacks[i]
		var source = stack.ShortSource
		var funcname = stack.Name
		var linenum = stack.CurrentLine
		if linenum >= 0 {
			var lines = strings.Split(stack.Source, "\n")
			if linenum < len(lines) {
				source = lines[linenum-1]
			}
		}
		var one = fmt.Sprintf("%s%s %s %d\n%s", prefix, source, funcname, linenum, suffix)
		str = str + one
	}
	return str
}

func FormatStackToMap(stacks []lua.LuaStackEntry) []map[string]interface{} {
	var count = len(stacks)
	var clones = make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		var stack = stacks[i]
		var clone = make(map[string]interface{})
		var linenum = stack.CurrentLine
		if linenum >= 0 {
			var lines = strings.Split(stack.Source, "\n")
			if linenum < len(lines) {
				clone["linesrc"] = lines[linenum-1]
			} else {
				clone["linesrc"] = stack.ShortSource
			}
		}
		clone["line"] = linenum
		clone["func"] = stack.Name
		clones[i] = clone
	}
	return clones
}

func Compile(meta *rscript.Meta, cache *scache.SCache) error {
	var L = lua.NewState()
	defer L.Close()
	var retcode = L.LoadString(meta.Script)
	if retcode != 0 {
		return fmt.Errorf("lua load string retcode %d != 0", retcode)
	}
	var err = L.Dump()
	if err == nil {
		meta.Binary = L.ToBytes(-1)
		meta.Lines = strings.Split(meta.Script, "\n")
		if len(meta.Hash) == 0 {
			meta.Hash = qencode.Md5Str(meta.Script)
		}
		if cache != nil {
			cache.Set(meta, meta.Hash)
		}
	}
	return err
}

func GetLuaPath() string {
	if len(_LUA_PATH) == 0 {
		_mutex.Lock()
		defer _mutex.Unlock()

		var g = global.GetInstance()
		if g.Config == nil {
			g.Config, _ = qconfig.ConfigLoad("config.json", "includes")
		}

		var err error
		var lua_version = lua.GetVersionNumber()
		var lua_version_without_dot = lua.GetVersionNumberWithoutDot()
		var lua_path = util.GetStr(g.Config, "../../src/github.com/camsiabor/qstock/lua/", "lua", "lua_path")
		var lua_cpath = util.GetStr(g.Config, lua_path, "lua", "lua_cpath")

		if lua_path, err = filepath.Abs(lua_path); err != nil {
			panic(err)
		}
		if lua_cpath, err = filepath.Abs(lua_cpath); err != nil {
			panic(err)
		}

		var lua_lib_suffix = "so"
		if runtime.GOOS == "windows" {
			lua_path = strings.Replace(lua_path, "\\", "/", -1)
			lua_cpath = strings.Replace(lua_cpath, "\\", "/", -1)
			lua_lib_suffix = "dll"
		}

		if lua_path[:len(lua_path)-1] != "/" {
			lua_path = lua_path + "/"
		}

		if lua_cpath[:len(lua_cpath)-1] != "/" {
			lua_cpath = lua_cpath + "/"
		}

		var lua_path_full = lua_path + "?.lua;" +
			lua_path + "?init.lua;" +
			lua_path + "?;" +
			lua_path + "lib/?.lua;" +
			lua_path + "lib/?init.lua;" +
			lua_path + "lib/?;"

		var lua_cpath_full = lua_cpath + "lib/?." + lua_lib_suffix + ";" +
			lua_cpath + "lib/?" + lua_version + "." + lua_lib_suffix + ";" +
			lua_cpath + "lib/?" + lua_version_without_dot + "." + lua_lib_suffix + ";" +
			lua_cpath + "lib/load.all." + lua_lib_suffix + ";" +
			lua_cpath + "lib/?;"

		_LUA_PATH = lua_path
		_LUA_PATH_FULL = lua_path_full

		_LUA_CPATH = lua_cpath
		_LUA_CPATH_FULL = lua_cpath_full

		qlog.Log(qlog.INFO, "lua", "LUA_PATH", _LUA_PATH)
		qlog.Log(qlog.INFO, "lua", "LUA_CAPTH", _LUA_CPATH)
		qlog.Log(qlog.INFO, "lua", "LUA_PATH full", _LUA_PATH_FULL)
		qlog.Log(qlog.INFO, "lua", "LUA_CAPTH full", _LUA_CPATH_FULL)
	}
	return _LUA_PATH
}

func InitState() (L *lua.State, err error) {

	defer func() {
		var pan = recover()
		if pan != nil {
			if L != nil {
				L.Close()
			}
			var ok bool
			err, ok = pan.(error)
			if !ok {
				err = fmt.Errorf("lua init state error %v", pan)
			}
		}
	}()

	GetLuaPath()

	L = luar.Init()

	L.OpenBase()
	L.OpenLibs()
	L.OpenTable()
	L.OpenString()
	L.OpenPackage()
	L.OpenOS()
	L.OpenMath()
	L.OpenDebug()
	L.OpenBit32()
	L.OpenDebug()

	L.PushString(_LUA_PATH_FULL)
	L.SetGlobal("LUA_PATH")

	L.PushString(_LUA_CPATH_FULL)
	L.SetGlobal("LUA_CPATH")

	L.GetGlobal("package")
	if !L.IsTable(-1) {
		return nil, fmt.Errorf("package is not a table? why? en? @_@?")
	}

	L.PushString(_LUA_PATH_FULL)
	L.SetField(-2, "path")

	L.PushString(_LUA_CPATH_FULL)
	L.SetField(-2, "cpath")

	var g = global.GetInstance()
	var gmodule = g.Data()
	luar.Register(L, _TOKEN_GLOBAL_MODULE, gmodule)

	L.Pop(-1)

	return L, err
}

func DefaultErrHandler(L *lua.State, pan interface{}) {
	if pan == nil {
		return
	}
	var ok bool
	L.Notice, ok = pan.(*lua.Interrupt)
	if !ok {
		var stackinfo = qref.StackInfo(5)
		var stackstr = util.AsStr(stackinfo["stack"], "")
		stackstr = strings.Replace(stackstr, "\t", "  ", -1)
		stackinfo["stack"] = strings.Split(stackstr, "\n")
		L.SetData("err_stack", stackinfo)
	}
}

func RunFile(L *lua.State, filename string, errhandler lua.LuaGoErrHandler) (rets []interface{}, err error) {

	GetLuaPath()

	var top_before = L.GetTop()
	var fpath = _LUA_PATH + filename
	if err = L.LoadFileEx(fpath); err != nil {
		return rets, err
	}

	luar.Register(L, "", map[string]interface{}{
		"Qrace": func(data interface{}) {
			L.SetData("qrace", data)
		},
	})

	if errhandler == nil {
		errhandler = DefaultErrHandler
	}

	err = L.CallHandle(0, lua.LUA_MULTRET, errhandler)
	var top_after = L.GetTop()
	var return_num = top_after - top_before
	if err == nil {
		if return_num > 0 {
			rets = make([]interface{}, return_num)
			for i := 0; i < return_num; i++ {
				rets[i], err = GetVal(L, i+1)
				if err != nil {
					break
				}
			}
		}
	}
	top_after = L.GetTop()
	if top_after-top_before > 0 {
		for i := 0; i < return_num; i++ {
			L.Pop(-1)
		}
	}

	if err != nil {
		var qrace = L.GetData("qrace")
		if qrace != nil {
			rets = make([]interface{}, 1)
			rets[0] = qrace
		}
	}

	return rets, err
}
