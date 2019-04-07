-- http://data.10jqka.com.cn/funds/ggzjl/
-- http://data.10jqka.com.cn/funds/ggzjl/field/zjjlr/order/desc/page/1/ajax/1/
--http://news.10jqka.com.cn/20190322/c610418561.shtml#refCountId=pop_50ab41b8_259

local th_mod_flow = require("sync.th.mod_flow")

local th_mod_flow_inst = th_mod_flow:new()

local opts = {}

opts.debug = false
opts.loglevel = 0
opts.browser = "firefox"

opts.request = false
opts.request_from = 1
opts.request_to = 71
opts.nice = 0

opts.concurrent = 5
opts.newsession = true
opts.persist = true

opts.date_show = 10

opts.date_offset = 0
opts.date_offset_to = 10
opts.date_offset_from = -opts.date_show - opts.date_offset

opts.db = "flow"
opts.datasrc = "th"
opts.field = "zjjlr"
opts.order = "desc"

opts.link_stock_group = true
opts.link_stock_snapshot = false

--opts.sort_field = "flow_io_rate"
opts.sort_field = "flow_big_rate_cross_ex"

th_mod_flow_inst:go_stock_group_profile(opts)