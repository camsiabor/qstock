-- http://data.10jqka.com.cn/funds/ggzjl/
-- http://data.10jqka.com.cn/funds/ggzjl/field/zjjlr/order/desc/page/1/ajax/1/


local filter_anti = function(opts, data, result)
    print("[filter] anti")
    local n = #data
    for i = 1, n do
        local one = data[i]
        local critical = 
            --one.flow_io_rate >= 1 and one.flow_big_in_rate >= 60
            -- one.flow_io_rate >= 2 and one.turnover >= 0.1 and one.change_rate <= 10
            one.flow_io_rate < 0.9 
            and one.change_rate > 0  and one.change_rate <= 5 
            and one.flow_big_in_rate >= 10
            and one.turnover >= 1
        if critical then
            result[#result + 1] = one
        end
    end
end

local filter_high_io = function(opts, data, result)
    print("[filter] high io")
    
    local n = #data
    for i = 1, n do
        local one = data[i]
        local critical = 
            --one.flow_io_rate >= 1 and one.flow_big_in_rate >= 60
            -- one.flow_io_rate >= 2 and one.turnover >= 0.1 and one.change_rate <= 10
            one.flow_io_rate >= 1.75 and one.turnover >= 0.1 and one.change_rate <= 10
        if critical then
            result[#result + 1] = one
        end
    end
end


local filter_single_force = function(opts, data, result)
    print("[filter] single force")
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

local filter_multi_force = function(opts, data, result)
    print("[filter] multi force")
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

local filter_moderate = function(opts, data, result)
    print("[filter] moderate")
    local n = #data
    for i = 1, n do
        local one = data[i]
        local critical = 
            ( one.change_rate >= 1.5 and one.change_rate <= 6.5 )
            and
            (
                one.flow_io_rate >= 1.25
                and one.flow_big_in_rate >= 25
            )
        
        if critical then
            result[#result + 1] = one
        end
    end
end


local th_mod_flow = require("sync.th.mod_flow")
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
opts.date_offset = -3
opts.reload_thereafter = 10

opts.pagesize = 71
opts.ch_lower = -2
opts.ch_upper = 6
opts.big_c_lower = 0.2
opts.big_c_upper = 10

opts.db = "flow"
opts.datasrc = "th"
opts.field = "zjjlr"
opts.order = "desc"

opts.sort_field = "flow_big_rate_cross_ex"

local filter_as_single = false


--opts.filter = opts.filter_moderate
--opts.filter = filter_high_io
opts.filter = filter_anti
--opts.filter = opts.fitler_single_force
--opts.filter = opts.fitler_multi_force

th_mod_flow_inst:go(opts)