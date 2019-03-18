-- http://data.10jqka.com.cn/funds/ggzjl/
-- http://data.10jqka.com.cn/funds/ggzjl/field/zjjlr/order/desc/page/1/ajax/1/


local thflow = require("sync.th.flow")
local inst = thflow:new()

local opts = {}

opts.debug = false
opts.loglevel = 0
opts.browser = "firefox"

opts.from = 1
opts.to = 71
opts.nice = 0
opts.concurrent = 1
opts.newsession = false
opts.persist = true

opts.dofetch = false
opts.date_offset = -1

opts.pagesize = 71
opts.ch_lower = -2.5
opts.ch_upper = 6
opts.big_c_lower = 0.2
opts.big_c_upper = 10

opts.db = "flow"
opts.datasrc = "th"
opts.field = "zjjlr"
opts.order = "desc"

opts.sort_field = "flow_big_rate_cross_ex"

opts.filter_balance_io_rate = function(opts, data, result)
    print("[filter] low")
    local n = #data
    for i = 1, n do
        local one = data[i]
        local critical = 
            one.flow_io_rate >= 1 
            and one.flow_big_in_rate >= 30
            and one.change_rate >= -1.5 and one.change_rate <= 5
        
        if critical then
            result[#result + 1] = one
        end
    end
end

opts.filter = opts.filter_balance_io_rate

inst:go(opts)