-- http://data.10jqka.com.cn/funds/ggzjl/
-- http://data.10jqka.com.cn/funds/ggzjl/field/zjjlr/order/desc/page/1/ajax/1/




local th_mod_flow = require("sync.th.mod_flow")
local filters = require("sync.th.mod_flow_filters")

local th_mod_flow_inst = th_mod_flow:new()

local opts = {}

opts.debug = false
opts.loglevel = 0
opts.browser = "firefox"

opts.from = 1
opts.to = 71
opts.nice = 0

opts.concurrent = 3
opts.newsession = false
opts.persist = true

opts.dofetch = false
opts.date_offset = 0
opts.date_offset_from = -2
opts.date_offset_to = 1

opts.db = "flow"
opts.datasrc = "th"
opts.field = "zjjlr"
opts.order = "desc"

opts.sort_field = "flow_io_rate"

opts.filters =  {
    -- codes
    --filters.codes({  codes = { "601929" } })
    
    -- moderate
    --filters.io({  io_lower = 1.5, io_upper = 100, ch_lower = -1.5, ch_upper = 3.5, big_in_lower = 3  })
    
    -- high io
    -- filters.io({  io_lower = 1.75, io_upper = 100, ch_lower = -1.5, ch_upper = 6.5, big_in_lower = 10  })
    
    -- flow in increase
    filters.io_increase({ in_lower = 55, in_upper = 100, in_swing = 3, ch_lower = -3, ch_upper = 7 })
}

th_mod_flow_inst:go(opts)