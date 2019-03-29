local global = require("q.global")
local json = require("common.json")
local mod_snapshot = require("sync.th.mod_snapshot")

local cache_code = global.cachem.Get("stock.code");
local cache_khistory = global.cachem.Get("stock.khistory");
local dates = global.calendar.List(3, 0, 0, true)
local codes = cache_code.Get(false, "sz.sh");
local map = {}
for i = 1, #codes do
    local code = codes[i];
    local ks = mod_snapshot:snapshot(code, dates)
    map[code] = ks
end

local str = json:encode(map)

print(str)




if 1 == 1 then
    return
end


local money = 80000
for i = 1, 36 do 
    print(money)
    money = money * 1.1
end
print(money)
    






if 1 == 1 then
    return 
end

local simple = require("common.simple")
local loggerm = require("q.logger")
local logger = loggerm:newstdout()
for i = 1, 3 do 
    logger:info("hello?")
end

--print(logger)
--simple.table_print_all(logger)