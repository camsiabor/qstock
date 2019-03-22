-- http://data.10jqka.com.cn/funds/ggzjl/
-- http://data.10jqka.com.cn/funds/ggzjl/field/zjjlr/order/desc/page/1/ajax/1/

local th_mod_flow = require("sync.th.mod_flow")
local th_mod_flow_inst = th_mod_flow:new()

local profile = Q.work.Profile

local opts = {}

opts.debug = false
opts.loglevel = 0
opts.browser = "firefox"

opts.from = 1
opts.to = 71
opts.nice = 0

opts.concurrent = profile["concurrent"]
if opts.concurrent == nil then
    opts.concurrent = 1
end

if Q.runtime.GOOS() == "windows" then
    opts.newsession = false
else
    opts.newsession = true
end

opts.persist = true

opts.dofetch = true
opts.date_offset = 0
opts.date_offset_from = 0
opts.date_offset_to = 0

opts.db = "flow"
opts.datasrc = "th"
opts.field = "zjjlr"
opts.order = "desc"

opts.sort_field = "flow_big_rate_cross_ex"

opts.print_data_from = -1
opts.print_data_to = -1

th_mod_flow_inst:go(opts)