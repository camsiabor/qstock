--[[

    14:15 -->
        (1) 1.4 <= io <= 100, 4.5 <= ch <= 10
        (2) 1.0 <= io <= 1.4, 1 <= ch <= 4.5
        (3) H, 1.4 <= io <= 100

]]--


local cal = require("common.cal")
local simple = require("common.simple")
local th_mod_flow = require("sync.th.mod_flow")
local filters = require("sync.th.mod_flow_filters")

local global = require("q.global")

--[[
local dates = global.calendar.List(15, 0, 0, false)
print(dates)
if 1 == 1 then
    return
end
]]--

local th_mod_flow_inst = th_mod_flow:new()

local opts = {}

opts.debug = false
opts.loglevel = 0
opts.browser = "firefox"

opts.request_from = 1
opts.request_to = 71
opts.request_each = 1
opts.concurrent = 1
opts.newsession = 5
opts.nice = 0

opts.persist = true

opts.db = "flow"
opts.datasrc = "th"
opts.field = "zjjlr"
opts.order = "desc"

opts.link_stock_group = true
opts.link_stock_snapshot = false


local names_bought = {
    "长城电工"

}


local adapt_ch_sum = function(opts, result, mapping, currindex)
    if mapping == nil then
        print("no mapping")
        return
    end
    local up = 0
    local down = 0
    local moderate = 0
    local n = #result
    for i = 1, n do
        local one = result[i]
        local code = one.code
        local series = mapping[code]
        if series ~= nil then
            local series_chs = simple.table_field_to_array(series, "change_rate", currindex + 1, #series)
            local sums = cal.array_step_sum(series_chs)
            local sums_up, sums_down = cal.array_up_down_count(sums, 0)
            if sums_up > 0 then
                up = up + 1
            else
                down = down + 1
            end
            local last = sums[#sums]
            if last == nil then
                last = "nil"
            end
            one.custom8 = sums_up .. "/" .. sums_down .. "/" .. last
            if sums_up > 0 then
                one.custom8 = one.custom8 .. " @"
            else
                one.custom8 = one.custom8 .. " FFF"
            end
        end

    end
    print("[adapt] [ch sum] [u]", up)
    print("[adapt] [ch sum] [d]", down)
    if up > 0 or down > 0 then
        print("[adapt] [ch sum] [rate]", simple.numcon((up) / (up + down) * 100), "%")
    end
end

--opts.date_ignore = { "0501", "0502", "0503" }
opts.sort_field = "flow_io_rate"
--opts.sort_field = "flow_big_rate_cross_ex"

opts.request = 1

opts.date_show = 15

opts.date_offset = -5
opts.date_offset_to = 7
--opts.date_offset_from = 0
opts.date_offset_from = -opts.date_show - opts.date_offset

opts.adapters = {
    adapt_ch_sum
}

local TROLL = 0

opts.filters = {

    filters.code_head( { head = "3", include = false } ),
    --filters.name_head( { head = "五", include = true } ),


    --names
    --filters.names({  names = names_bought }),
    --filters.names({  names = names_tobe }),
    --filters.names({  names = { "英联股份"} }),

    --filters.codes({  codes = { "002467", "600662", "603218", "600371" } }),

    --[[
    filters.avg_diff({  field = "turnover", set = "custom", short_cycle = 2, long_cycle = 4 , diff_lower = -10000, diff_upper = 10000 }),
    filters.avg_diff({  field = "change_rate", set = "custom2", short_cycle = 2, long_cycle = 4, diff_lower = -10000, diff_upper = 10000 }),
    filters.ratio({  field1 = "custom", field2 = "custom2", set = "custom3", ratio_lower = -10000, ratio_upper = 10000, date_offset = 0 }),
    ]]--

    --filters.names_contain({  names = { ""} }),

    -------------------------------------------------------------------------------------------------------------
    --codes
    --[[
    filters.codes({ codes = codes0 }),
    filters.avg_diff({  field = "turnover", set = "custom", short_cycle = 2, long_cycle = 4 , diff_lower = -10000, diff_upper = 10000 }),
    filters.avg_diff({  field = "change_rate", set = "custom2", short_cycle = 2, long_cycle = 4, diff_lower = -10000, diff_upper = 10000 }),
    filters.ratio({  field1 = "custom2", field2 = "custom", set = "custom3", absolute = true, ratio_lower = -10000, ratio_upper = 10000, date_offset = 0 }),
    filters.avg_diff({  field = "turnover", set = "custom4", short_cycle = 3, long_cycle = 6 , diff_lower = -5000, diff_upper = 5000 }),
    filters.avg_diff({  field = "change_rate", set = "custom5", short_cycle = 3, long_cycle = 6, diff_lower = -5000, diff_upper = 5000 }),
    filters.ratio({  field1 = "custom5", field2 = "custom4", set = "custom6", absolute = true, ratio_lower = 0, ratio_upper = 5000, date_offset = 0 }),
    ]]--

    --------------------------------------------------------------------------------------------------------------

    --groups
    --filters.groups( { groups = { "华为概念" } } ),

    --filters.io({  io_lower = 0, io_upper = 100, ch_lower = 3.5, ch_upper = 11, big_in_lower = 0, date_offset = 0 }),

    --------------------------------------------------------------------------------------------------------------


    -- (A) 高 IO, 高 CH


    --filters.io({  io_lower = 1.3, io_upper = 10, ch_lower = -4.5, ch_upper = -0.01, big_in_lower = 0, date_offset = 0 }),
    --filters.io({  io_lower = 0.9, io_upper = 1.1, ch_lower = 1, ch_upper = 3.5, big_in_lower = 0, date_offset = 0 }),
    --filters.io({  io_lower = 1.6, io_upper = 5, ch_lower = 0, ch_upper = 3.5, big_in_lower = 0, date_offset = 0 }),

    filters.io({  io_lower = 1.3, io_upper = 1.6, ch_lower = 0, ch_upper = 3.5, big_in_lower = 0, date_offset = 0 }),
    --filters.io({  io_lower = 1.35, io_upper = 11, ch_lower = 4, ch_upper = 11, big_in_lower = 0, date_offset = 0 }),

    filters.io_any_simple({  io_lower = 1.3, io_upper = 10, date_offset_from = -12, date_offset_to = -1, tag = true }),

    filters.avg_diff({  field = "turnover", set = "custom",
                        short_cycle = 2, long_cycle = 4 , per = 1, diff_lower = TROLL, diff_upper = 10 }),
    filters.avg_diff({  field = "change_rate", set = "custom2",
                        short_cycle = 2, long_cycle = 4, per = 1, diff_lower = 0, diff_upper = 10, deduce = "close" }),
    filters.ratio({  field1 = "custom", field2 = "custom2", set = "custom3",
                     absolute = true, ratio_lower = 0, ratio_upper = 1000, date_offset = 0 }),

    filters.avg_diff({  field = "turnover", set = "custom4",
                        short_cycle = 4, long_cycle = 8 , per = 1, diff_lower = TROLL, diff_upper = 10 }),
    filters.avg_diff({  field = "change_rate", set = "custom5",
                        short_cycle = 4, long_cycle = 8, per = 1, diff_lower = 0, diff_upper = 10, deduce = "close" }),
    filters.ratio({  field1 = "custom4", field2 = "custom5", set = "custom6",
                     absolute = true, ratio_lower = 0, ratio_upper = 1000, date_offset = 0 }),

    filters.ratio({  field1 = "custom3", field2 = "custom6", set = "custom7",
                     absolute = true, ratio_lower = 0, ratio_upper = 1000, date_offset = 0 }),
    --[[
     ]]--

    --------------------------------------------------------------------------------------------------------------
    -- 低级
    --[[
    filters.io({  io_lower = 1.75, io_upper = 5, ch_lower = -4.5, ch_upper = 11, big_in_lower = 0, date_offset = 0 }),
    --filters.io({  io_lower = 1.6, io_upper = 5, ch_lower = -4.5, ch_upper = 3.5, big_in_lower = 0, date_offset = 0 }),
    filters.io_any_simple({  io_lower = 1.3, io_upper = 10, date_offset_from = -12, date_offset_to = -1, tag = true, count = 4 }),
    ]]--


    --------------------------------------------------------------------------------------------------------------



    --[[
    filters.io({  io_lower = 1.6, io_upper = 10, ch_lower = -4.5, ch_upper = 5, big_in_lower = 0, date_offset = 0 }),
    filters.io_any_simple({  io_lower = 1.3, io_upper = 10, date_offset_from = -12, date_offset_to = -1, tag = true }),
    ]]--

    -- anti io serial
    --[[
    filters.io({  io_lower = 0, io_upper = 1, ch_lower = 0, ch_upper = 5, big_in_lower = 0, date_offset = -1 }),
    filters.io({  io_lower = 1, io_upper = 1.1, ch_lower = 0, ch_upper = 5, big_in_lower = 0, date_offset = 0 }),
    --filters.io({  io_lower = 0, io_upper = 1, ch_lower = 0, ch_upper = 5, big_in_lower = 0, date_offset = 0 }),
    filters.io_any_simple({  io_lower = 1.3, io_upper = 10, date_offset_from = -12, date_offset_to = -1, tag = true }),

    ]]--


    --------------------------------------------------------------------------------------------------------------

    -- (B) H股
    --[[
    filters.groups( { groups = { "H股" } } ),
    filters.io({  io_lower = 1.3, io_upper = 100, ch_lower = 0, ch_upper = 11, big_in_lower = 0, date_offset = 0 }),

    ]]--


    --------------------------------------------------------------------------------------------------------------

    -- (C)低吸
    --[[

    filters.io({  io_lower = 1.3, io_upper = 10, ch_lower = 1, ch_upper = 3.5, big_in_lower = 0, date_offset = 0 }),
    --filters.io_any_simple({  io_lower = 1.3, io_upper = 10, date_offset_from = -12, date_offset_to = -1, tag = false }),

    filters.avg_diff({  field = "turnover", set = "custom",
        short_cycle = 2, long_cycle = 4 , per = 1, diff_lower = 0, diff_upper = 10 }),
    filters.avg_diff({  field = "change_rate", set = "custom2",
        short_cycle = 2, long_cycle = 4, per = 1, diff_lower = 0, diff_upper = 10, deduce = "close" }),
    filters.ratio({  field1 = "custom", field2 = "custom2", set = "custom3",
        absolute = true, ratio_lower = 0, ratio_upper = 10, date_offset = 0 }),

    filters.avg_diff({  field = "turnover", set = "custom4",
        short_cycle = 4, long_cycle = 8 , per = 1, diff_lower = 0, diff_upper = 10 }),
    filters.avg_diff({  field = "change_rate", set = "custom5",
        short_cycle = 4, long_cycle = 8, per = 1, diff_lower = 0, diff_upper = 10, deduce = "close" }),
    filters.ratio({  field1 = "custom4", field2 = "custom5", set = "custom6",
        absolute = true, ratio_lower = 0, ratio_upper = 10, date_offset = 0 }),

    filters.ratio({  field1 = "custom3", field2 = "custom6", set = "custom7",
        absolute = true, ratio_lower = 0.9, ratio_upper = 10, date_offset = 0 }),
    --]]

    --------------------------------------------------------------------------------------------------------------

    -- (D)非常低吸

    --[[
    filters.io({  io_lower = 1.35, io_upper = 100, ch_lower = 1, ch_upper = 10, big_in_lower = 0, date_offset = 0 }),
    --filters.io_any_simple({  io_lower = 1.2, io_upper = 10, date_offset_from = -5, date_offset_to = -1 }),
    filters.avg_diff({  field = "turnover", set = "custom", short_cycle = 2, long_cycle = 4 , diff_lower = -100, diff_upper = 100 }),
    filters.avg_diff({  field = "change_rate", set = "custom2", short_cycle = 2, long_cycle = 4, diff_lower = -200, diff_upper = 200 }),
    filters.ratio({  field1 = "custom2", field2 = "custom", set = "custom3", absolute = true, ratio_lower = 3, ratio_upper = 500, date_offset = 0 }),
    ]]--


    --------------------------------------------------------------------------------------------------------------

    -- ST
    --filters.st({ }),
    --filters.io({  io_lower = 1.5, io_upper = 100, ch_lower = 0, ch_upper = 11, big_in_lower = 0, date_offset = -3 }),



    --------------------------------------------------------------------------------------------------------------


    -- 中高 IO, 高 CH
    --filters.io({  io_lower = 1.8, io_upper = 100, ch_lower = 3, ch_upper = 11, big_in_lower = 0, date_offset = -5 }),
    --filters.io({  io_lower = 0, io_upper = 100, ch_lower = 8.5, ch_upper = 11, big_in_lower = 0, date_offset = -2 }),

    --------------------------------------------------------------------------------------------------------------

    -- narrow io
    --filters.io({  io_lower = 1.2, io_upper = 1.3, ch_lower = 5, ch_upper = 8.5, big_in_lower = 0, date_offset = -5}),
    --filters.io({  io_lower = 1.4, io_upper = 5, ch_lower = 0, ch_upper = 11, big_in_lower = 0, date_offset = 0}),
    --filters.avg_diff({  field = "turnover", set = "custom", short_cycle = 2, long_cycle = 4 , diff_lower = 5, diff_upper = 10000 }),
    --filters.avg_diff({  field = "change_rate", set = "custom2", short_cycle = 2, long_cycle = 4, diff_lower = 5, diff_upper = 10000 }),
    --filters.ratio({  field1 = "custom", field2 = "custom2", set = "custom3", ratio_lower = 0.1, ratio_upper = 10000, date_offset = 0 }),




    --------------------------------------------------------------------------------------------------------------

    -- 非常低 IO, 正 CH, 蓄力股
    --filters.io({  io_lower = 0.5, io_upper = 1.2, ch_lower = -0.5, ch_upper = 5, big_in_lower = 0, date_offset = -4 }),
    --filters.io({  io_lower = 1, io_upper = 1.25, ch_lower = 5, ch_upper = 8, big_in_lower = 0, date_offset = -3 }),


    --------------------------------------------------------------------------------------------------------------

    -- ch 0
    --filters.io({  io_lower = 0.8, io_upper = 1.2, ch_lower = -0.1, ch_upper = 0.1, big_in_lower = 0, date_offset = -1 })


    ---------------------------------------------------------------------------------------------

    -- 迥 CH
    --filters.io({  io_lower = 0, io_upper = 100, ch_lower = 9, ch_upper = 11, big_in_lower = 0, date_offset = 0 }),

    --------------------------------------------------------------------------------------------------------------

    --anti io series
    --filters.io({  io_lower = 0.8, io_upper = 100, ch_lower = -0.1, ch_upper = 0.15, big_in_lower = 0, date_offset = -1  }),
    --filters.io({  io_lower = 1.3, io_upper = 100, ch_lower = 3, ch_upper = 11, big_in_lower = 0, date_offset = 0  }),
    --filters.io_all({  io_lower = 0, io_upper = 100, ch_lower = -3, ch_upper = 11, big_in_lower = 0, date_offset = 0  }),

    --------------------------------------------------------------------------------------------------------------
    -- anti io series 2
    --filters.io({  io_lower = 1.1, io_upper = 100, ch_lower = -1.5, ch_upper = 0, big_in_lower = 0, date_offset = -1  }),
    --filters.io({  io_lower = 0.5, io_upper = 3, ch_lower = 1, ch_upper = 11, big_in_lower = 0, date_offset = 0  }),
    --filters.io_all({  io_lower = 0, io_upper = 100, ch_lower = -3, ch_upper = 11, big_in_lower = 0, date_offset = 0  }),


    -- two days anti io io >= 2 plus high io
    --filters.io({  io_lower = 0.6, io_upper = 0.97, ch_lower = -0.1, ch_upper = 3, big_in_lower = 0, date_offset = -1 }),
    --filters.io({  io_lower = 0.6, io_upper = 0.97, ch_lower = -0.1, ch_upper = 3, big_in_lower = 0, date_offset = 0 }),



    --------------------------------------------------------------------------------------------------------------

    -- high io, ch == 0
    --filters.io({  io_lower = 1.05, io_upper = 100, ch_lower = -0.1, ch_upper = 0.1, big_in_lower = 0 })

    --------------------------------------------------------------------------------------------------------------

    -- serial anti io
    --filters.io({  io_lower = 0.7, io_upper = 1, ch_lower = -0.1, ch_upper = 8, big_in_lower = 0, date_offset = -2 }),
    --filters.io({  io_lower = 0.7, io_upper = 1, ch_lower = -0.1, ch_upper = 8, big_in_lower = 0, date_offset = -1 }),
    -- filters.io({  io_lower = 0.7, io_upper = 1, ch_lower = -0.1, ch_upper = 8, big_in_lower = 0, date_offset = 0 }),




    --------------------------------------------------------------------------------------------------------------

    -- high io, io >= 1.4 && ch >= 0 && ch <= 2
    --filters.io({  io_lower = 1.4, io_upper = 100, ch_lower = 0, ch_upper = 2, big_in_lower = 0 })

    --------------------------------------------------------------------------------------------------------------

    -- moderate io and low ch
    --filters.io({  io_lower = 1.35, io_upper = 100, ch_lower = 0, ch_upper = 2, big_in_lower = 0 })

    -------------------------------------------------------------------------------------------------------------

    -- anti io, 0.5 <= io <= 0.98, ch >= 0, ch <= 3
    --filters.io({  io_lower = 0.8, io_upper = 0.98, ch_lower = 1, ch_upper = 3, big_in_lower = 0, date_offset = 0 }),

    -------------------------------------------------------------------------------------------------------------

    --moderate
    --filters.io({  io_lower = 1.2, io_upper = 100, ch_lower = 5, ch_upper = 11, big_in_lower = 0  }),

    -- low io
    --filters.io({  io_lower = 0.8, io_upper = 1.25, ch_lower = 0, ch_upper = 5, big_in_lower = 0  })




    --very high io
    --filters.io({  io_lower = 1.5, io_upper = 100, ch_lower = -1, ch_upper = 10.5, big_in_lower = 0 }),

    -- io ceil
    --filters.io({  io_lower = 0, io_upper = 100, ch_lower = 8, ch_upper = 11, big_in_lower = 0 }),

    -- flow in increase
    --filters.io_increase({ in_lower = 30, in_upper = 100, in_swing = 5, ch_lower = -10, ch_upper = 10 })

    -- chase high
    --filters.io({  io_lower = 1.2, io_upper = 100, ch_lower = 5, ch_upper = 10, big_in_lower = 0 }),

    -- find underline
    --filters.io_any({  io_lower = 1.85, io_upper = 100, ch_lower = 0, ch_upper = 10, big_in_lower = 0  }),

    --filters.io_any({  io_lower = 0.5, io_upper = 0.75, ch_lower = -1, ch_upper = 5, big_in_lower = 0  }),
    --filters.io_any({  io_lower = 1.75, io_upper = 100, ch_lower = 0, ch_upper = 10.5, big_in_lower = 0, ch_avg_lower = 0, ch_avg_upper = 10  }),

    --all
    --filters.io_all({  io_lower = 1, io_upper = 100, ch_lower = -2.5, ch_upper = 10.5, big_in_lower = 0, ch_avg_lower = 0, ch_avg_upper = 10  }),



}

th_mod_flow_inst:go(opts)