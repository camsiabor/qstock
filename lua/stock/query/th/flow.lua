-- http://data.10jqka.com.cn/funds/ggzjl/
-- http://data.10jqka.com.cn/funds/ggzjl/field/zjjlr/order/desc/page/1/ajax/1/


local th_mod_flow = require("sync.th.mod_flow")
local th_mod_flow_inst = th_mod_flow:new()

local opts = {}

opts.debug = false
opts.loglevel = 0
opts.browser = "firefox"

opts.from = 1
opts.to = 71
opts.nice = 0
opts.concurrent = 5
opts.newsession = false
opts.persist = true

opts.dofetch = false
opts.date_offset = 0

opts.pagesize = 71
opts.ch_lower = -1
opts.ch_upper = 6
opts.big_c_lower = 0.2
opts.big_c_upper = 10

opts.db = "flow"
opts.datasrc = "th"
opts.field = "zjjlr"
opts.order = "desc"

opts.sort_field = "flow_big_rate_cross_ex"

local filter_as_single = false

opts.filter_single_force = function(opts, data, result)
    local n = #data
    for i = 1, n do
        local one = data[i]
        local critical = 
            ( one.change_rate >= -1.5 and one.change_rate <= 6.5 )
            and
            (
                ( one.flow_io_rate >= 1.25  and one.flow_big_in_rate >= 35 )
                or 
                ( one.flow_io_rate >= 1.75 )
            )
        
        if critical then
            result[#result + 1] = one
        end
    end
end

opts.filter_single_force = function(opts, data, result)
    local n = #data
    for i = 1, n do
        local one = data[i]
        local critical = 
            ( one.change_rate >= 1.5 and one.change_rate <= 6.5 )
            and
            (
                one.flow_io_rate >= 1.1  
                and one.flow_big_in_rate >= 35
                and one.flow_big_rate_total >= 1
            )
        
        if critical then
            result[#result + 1] = one
        end
    end
end

if filter_as_single then
    opts.filter = opts.filter_single_force
else
    opts.filter = opts.filter_multi_force
end
th_mod_flow_inst:go(opts)