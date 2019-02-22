--local mobdebug = require("mobdebug")
-- mobdebug.start()

--[[
local http = require("socket.http")
local b, c, h = http.request("http://www.baidu.com")

print(b)
print(c)
prin1t(h)
]]--


--local x = add(10,20)

function add(a, b, c)
    return a + b + c
end

local a = 10
local b = 12
local c = add(a, b, 102)

return c, c + 1, c + 2