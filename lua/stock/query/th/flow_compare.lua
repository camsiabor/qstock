-- http://data.10jqka.com.cn/funds/ggzjl/
-- http://data.10jqka.com.cn/funds/ggzjl/field/zjjlr/order/desc/page/1/ajax/1/

local simple = require("common.simple")
local mod_th_flow = require("sync.th.mod_flow")
local mod_th_flow_inst = mod_th_flow:new()


local daycount = 2
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

opts.filter = function(opts, data, result)
    local n = #data
    for i = 1, n do
        local one = data[i]
        local critical =
                (
                    one.flow_io_rate >= 1.25
                    and one.flow_big_in_rate >= 35
                    and one.change_rate >= -1.5 and one.change_rate <= 6.5
                )
                or
                (
                    one.flow_io_rate >= 1.75
                )
        if critical then
            result[#result + 1] = one
        end
    end
end

-----------------------------------------------------------------------------------------

local results = {}
local results_map_array = {}
local date_offset = -daycount + 1
for i = 1, daycount do
    opts.data = {}
    opts.result = {}
    opts.dofetch = false
    opts.date_offset = date_offset
    
    if dofetchcurr and i == 1 then
        opts.dofetch = true
    end
    
    mod_th_flow_inst:go(opts)
    
    results[#results + 1] = opts.result
    results_map_array[#results_map_array + 1] = simple.array_to_map(opts.result, "code")
    
    print("date_offset", date_offset, "result count", #opts.result)
    
    date_offset = date_offset + 1
end

---------------------------------------------------------------------------------------------------
print(results_map_array[1]["002659"])
print(results_map_array[2]["002659"])

local complex = {}
simple.maps_intersect(results_map_array, function (maps, key)
    for i = 1, #maps do
        local map = maps[i]
        local v = map[key]
        complex[#complex + 1] = v
    end
end)

print("intersect", #complex / 2)

mod_th_flow_inst:print_data(opts, complex)