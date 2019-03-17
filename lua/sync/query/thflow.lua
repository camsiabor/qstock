-- http://data.10jqka.com.cn/funds/ggzjl/
-- http://data.10jqka.com.cn/funds/ggzjl/field/zjjlr/order/desc/page/1/ajax/1/


local thflow = require("sync.th.flow")
local inst = thflow:new()

local opts = {}

opts.debug = false

opts.from = 1
opts.to = 5
opts.nice = 0
opts.concurrent = 1
opts.newsession = false

opts.dofetch = true
opts.persist = true
opts.pagesize = 50

opts.ch_lower = -2.5
opts.ch_upper = 6
opts.big_c_lower = 0.2
opts.big_c_upper = 10

opts.db = "flow"
opts.datasrc = "th"
opts.field = "zjjlr"
opts.order = "desc"

opts.sort_field = "flow_big_rate_cross_ex"

inst:go(opts)