package run

import (
	"github.com/camsiabor/golua/lua"
	"github.com/camsiabor/golua/luar"
	"strings"
)

func LuaGetVal(L *lua.State, idx int) (interface{}, error) {
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

func LuaFormatStack(stacks []lua.LuaStackEntry) []lua.LuaStackEntry {
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

func LuaFormatStackToMap(stacks []lua.LuaStackEntry) []map[string]interface{} {
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
