package httpv

import (
	"github.com/aarzilli/golua/lua"
	"github.com/gin-gonic/gin"
)

func (o HttpServer) handleLuaCmd(cmd string, m map[string]interface{}, c * gin.Context) {
	var L = lua.NewState()
	// TODO

	defer L.Close();

}


