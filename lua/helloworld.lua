local simple = require("common.simple")
local logger = require("common.logger")

logger = logger:new()

logger:log("hello?", "power", "over", "whelming")

local debuginfo = debug.getinfo(1)
simple.table_print_all(debuginfo)
print(line)