local mod_snapshot = require("sync.th.mod_snapshot")

local dates = mod_snapshot:dates(3, 0, 0)
local ks = mod_snapshot:snapshot("ch000009", dates)

--print(ks[1])

for k, v in pairs(ks[1]) do
    print(k, v)
end

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