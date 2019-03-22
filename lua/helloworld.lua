local simple = require("common.simple")
local logger = require("common.logger")


function try()
    logger = logger:new()
    logger:trace("hello?", "power", "over", "whelming")
end

try()

local debuginfo = debug.getinfo(1)
simple.table_print_all(debuginfo)
print(line)