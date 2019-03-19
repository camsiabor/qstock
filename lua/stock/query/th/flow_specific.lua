-- http://data.10jqka.com.cn/funds/ggzjl/
-- http://data.10jqka.com.cn/funds/ggzjl/field/zjjlr/order/desc/page/1/ajax/1/

local simple = require("common.simple")
local mod_th_flow = require("sync.th.mod_flow")
local mod_th_flow_inst = mod_th_flow:new()


local daycount = 3
local dofetchcurr = false

local opts = {}
opts.data = {}
opts.result = {}

opts.debug = false
opts.loglevel = 0
opts.browser = "firefox"

opts.from = 1
opts.to = 71
opts.nice = 0
opts.concurr = 1
opts.newsession = false
opts.persist = true

opts.dofetch = true
opts.date_offset = 0

opts.pagesize = 71
opts.ch_lower = -2.5
opts.ch_upper = 6
opts.big_c_lower = 0.2
opts.big_c_upper = 10

opts.db = "flow"
opts.datasrc = "th"
opts.field = "zjjlr"
opts.order = "desc"

opts.print_data = false
opts.sort_field = "flow_big_rate_cross_ex"


mod_th_flow_inst:reload(opts, opts.data, opts.result)

mod_th_flow_inst:print_data(opts, complex)