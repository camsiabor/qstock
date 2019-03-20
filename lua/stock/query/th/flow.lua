-- http://data.10jqka.com.cn/funds/ggzjl/
-- http://data.10jqka.com.cn/funds/ggzjl/field/zjjlr/order/desc/page/1/ajax/1/


local filter_anti = function(one, series, code, opts)
    return
        --one.flow_io_rate >= 1 and one.flow_big_in_rate >= 60
        -- one.flow_io_rate >= 2 and one.turnover >= 0.1 and one.change_rate <= 10
        one.flow_io_rate < 0.9
        and one.change_rate > 0  and one.change_rate <= 5
        and one.flow_big_in_rate >= 10
        and one.turnover >= 1
end

local filter_high_io = function(one, series, code, opts)
    return
        --one.flow_io_rate >= 1 and one.flow_big_in_rate >= 60
        -- one.flow_io_rate >= 2 and one.turnover >= 0.1 and one.change_rate <= 10
        one.flow_io_rate >= 1.75 and one.turnover >= 0.1 and one.change_rate <= 10
end


local filter_single_force = function(one, series, code, opts)
    return
        ( one.change_rate >= -1.5 and one.change_rate <= 6.5 )
            and
            (
                    ( one.flow_io_rate >= 1.25  and one.flow_big_in_rate >= 35 )
                            or
                            ( one.flow_io_rate >= 1.75 )
            )
end

local filter_multi_force = function(one, series, code, opts)
    return
        ( one.change_rate >= 1.5 and one.change_rate <= 6.5 )
        and
        (
            one.flow_io_rate >= 1.1
            and one.flow_big_in_rate >= 35
            and one.flow_big_rate_total >= 1
        )
end

local filter_moderate = function(one, series, code, opts)
    return
        ( one.change_rate >= 1.5 and one.change_rate <= 6.5 )
        and
        (
            one.flow_io_rate >= 1.25
            and one.flow_big_in_rate >= 25
        )
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
opts.date_offset = 0
opts.date_offset_from = 0
opts.date_offset_to = 0

opts.db = "flow"
opts.datasrc = "th"
opts.field = "zjjlr"
opts.order = "desc"

opts.sort_field = "flow_big_rate_cross_ex"

--opts.filter = filter_moderate
--opts.filter = filter_high_io
opts.filter = filter_anti
--opts.filter = fitler_single_force
--opts.filter = fitler_multi_force

th_mod_flow_inst:go(opts)