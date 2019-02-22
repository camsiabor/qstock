--local mobdebug = require("mobdebug")
-- mobdebug.start()


local http = require("socket.http")
-- local b, c, h = http.request("http://127.0.0.1/h/run/r.html")



--local x = add(10,20)
--[[
function add(a, b, c)
    return a + b + c
end

local a = 10
local b = 12
local c = add(a, b, 102)
]]--

return "power"