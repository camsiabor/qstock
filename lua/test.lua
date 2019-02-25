--local mobdebug = require("mobdebug")
-- mobdebug.start()

-- local http = require("socket.http")
-- local b, c, h = http.request("http://127.0.0.1/h/run/r.html")


print(Q.http ~= nil)

local api = Q.global.Config["api"]
local tushare = api["tushare"]["profiles"]
local khistory = tushare["k.history"]
local d = khistory["nice"]

print(d)
print(type(d))

return 0