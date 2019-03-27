local simple = require("common.simple")
local loggerm = require("q.logger")


local logger = loggerm:newstdout()
for i = 1, 3 do 
    logger:info("hello?")
end

--print(logger)
--simple.table_print_all(logger)