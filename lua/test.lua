

--local p = "../../src/github.com/camsiabor/qstock/lua/?.lua"
--local m_package_path = package.path
--package.path = string.format("%s;%s?.lua;%s?/init.lua", m_package_path, p, p)

--require("common.util")

print("[lua] i am test")

require("helloworld")
require("lib.common.util")

--local http = require("lib.luasocket.socket")

local socket = require("socket.core")

for k, v in pairs(socket) do
    print(k, v)
end

print(LUA_PATH)
print(package.path)
print(package.cpath)


local http = require("socket.http")
local b, c, h = http.request("http://www.baidu.com")

print(b)
print(c)
print(h)

--local x = add(10,20)
return c, c + 1, c + 2, #b