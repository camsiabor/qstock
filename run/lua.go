package run

import (
"github.com/aarzilli/golua/lua"
"github.com/stevedonovan/luar"
)

func LuaGetVal(L * lua.State, idx int) (interface{}, error) {
	if (L.IsNoneOrNil(idx)) {
		return nil, nil;
	}
	var ltype = int(L.Type(idx));
	switch ltype {
	case int(lua.LUA_TNUMBER):
		return L.ToNumber(idx), nil;
	case int(lua.LUA_TSTRING):
		return L.ToString(idx), nil;
	case int(lua.LUA_TBOOLEAN):
		return L.ToBoolean(idx), nil;
	}
	var r interface{};
	var err = luar.LuaToGo(L, idx, &r);
	return r, err;
}