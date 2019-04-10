local ah_table = require("sync.th.mod_ah_table")
local ah_table_inst = ah_table:new()

local opts = { }
opts.loglevel = 0
opts.browser = "gorilla"

opts.request = true
opts.db = "group"
opts.datasrc = "th"

ah_table_inst:go(opts)