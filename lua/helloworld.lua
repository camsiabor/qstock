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