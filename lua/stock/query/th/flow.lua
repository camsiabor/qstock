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

opts.result_adapter = function(opts, result, mapping, currindex)
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
            one.customex = sums_up .. "/" .. sums_down .. "/" .. last
        end
        
    end
    print("[up]", up)
    print("[down]", down)
    print("[rate]", simple.numcon((up) / (up + down) * 100), "%")
end


opts.sort_field = "flow_io_rate"
--opts.sort_field = "flow_big_rate_cross_ex"

opts.request = false

opts.date_show = 12

opts.date_offset = -5
opts.date_offset_to = 10
--opts.date_offset_from = 0
opts.date_offset_from = -opts.date_show - opts.date_offset


opts.filters = {
    
    filters.no3(),
    
    
    --names
    --filters.names({  names = names_bought }),
    --filters.names({  names = names_tobe }),
    --filters.names({  names = { "拉夏贝尔"} }),
    
    
    
    -------------------------------------------------------------------------------------------------------------
    --codes
    --filters.codes({ codes = "603157" })
    
    --------------------------------------------------------------------------------------------------------------
    
    --groups
    --filters.groups( { groups = { "电力改革" } } ),

    
    
    --------------------------------------------------------------------------------------------------------------
    -- 很高的 IO,
    --filters.io({  io_lower = 1.75, io_upper = 100, ch_lower = 0, ch_upper = 11, big_in_lower = 0, date_offset = -3 }),
    
    --------------------------------------------------------------------------------------------------------------
    
    -- 高 IO, 高 CH
    --filters.io({  io_lower = 1.4, io_upper = 100, ch_lower = 4.5, ch_upper = 11, big_in_lower = 0, date_offset = 0 }),
    --filters.ma_diff({  ma_short_cycle = 3, ma_long_cycle = 6, ma_diff_lower = 50, ma_diff_upper = 150 }),
    
    --------------------------------------------------------------------------------------------------------------
    
    --H股
    --filters.groups( { groups = { "H股" } } ),
    --filters.io({  io_lower = 1.3, io_upper = 100, ch_lower = 0, ch_upper = 11, big_in_lower = 0, date_offset = -11 })
    --filters.io({  io_lower = 0, io_upper = 100, ch_lower = 4.5, ch_upper = 11, big_in_lower = 0, date_offset = 0 })
    
    --------------------------------------------------------------------------------------------------------------
    
    -- ST
    --filters.st({ }),
    --filters.io({  io_lower = 1.5, io_upper = 100, ch_lower = 0, ch_upper = 11, big_in_lower = 0, date_offset = -3 }),

    --------------------------------------------------------------------------------------------------------------
    
    -- 低吸
    filters.io({  io_lower = 1.35, io_upper = 100, ch_lower = 1.5, ch_upper = 4.5, big_in_lower = 0, date_offset = 0 }),
    filters.ma_diff({  ma_short_cycle = 3, ma_long_cycle = 6, ma_diff_lower = 0, ma_diff_upper = 150 }),
    
    --------------------------------------------------------------------------------------------------------------
    

    -- 中高 IO, 高 CH
    --filters.io({  io_lower = 1.8, io_upper = 100, ch_lower = 3, ch_upper = 11, big_in_lower = 0, date_offset = -5 }),
    --filters.io({  io_lower = 0, io_upper = 100, ch_lower = 8.5, ch_upper = 11, big_in_lower = 0, date_offset = -2 }),
    
    --------------------------------------------------------------------------------------------------------------

    -- narrow io
    --filters.io({  io_lower = 1.2, io_upper = 1.3, ch_lower = 5, ch_upper = 8.5, big_in_lower = 0, date_offset = -5}),
    
    --filters.io({  io_lower = 0, io_upper = 5, ch_lower = 5, ch_upper = 11, big_in_lower = 0, date_offset = -1}),
    
    
    
  
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