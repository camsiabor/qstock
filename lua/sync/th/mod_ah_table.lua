-- http://vis-free.10jqka.com.cn/hk/cache/ahld_yjl_desc_b.html


local M = {}
M.__index = M
M.url_ah_table = "http://vis-free.10jqka.com.cn/hk/cache/ahld_yjl_desc_b.html"

local global = require("q.global")
local json = require('common.json')
local simple = require("common.simple")

-------------------------------------------------------------------------------------------

function M:new()
    local inst = {}
    inst.__index = self
    setmetatable(inst, self)
    return inst
end

-------------------------------------------------------------------------------------------
function M:request(opts)

    local url_pattern = "http://vis-free.10jqka.com.cn/hk/cache/ahld_yjl_desc_b.html"

    local reqopts = {}

    local url = string.format(url_pattern, opts.request_board, opts.request_field, opts.request_order, 1)
    local reqopt = {}
    reqopt["url"] = url
    reqopts[1] = reqopt

    local err
    local browser = global[opts.browser]
    reqopts, err = browser.Get(reqopts, 0, 0, 1, 1)

    if err ~= nil then
        print("[request] fatal", err)
    end
    reqopt = reqopts[1]
    self:parse_html(opts, data, reqopt)
    return data

end

-------------------------------------------------------------------------------------------
function M:parse_html(opts, data, reqopt)
    local url = reqopt["url"]
    local html = reqopt["content"]
    if html == nil then
        print("[error] request failure", url)
        print(reqopt["err"])
        return
    end
    
    local tag_start = '<div class="table%-content"'
    local tag_end = '</div>'
    
    local istart = string.find(html, tag_start)
    local iend = string.find(html, tag_end, istart)
    html = string.sub(html, istart, iend + #tag_end)


    if self.xml == nil then
        self.xml = require("common.xml2lua.xml2lua")
        self.xml_tree_handler = require("common.xml2lua.tree")
    end

    local tree = self.xml_tree_handler:new()
    local parser = self.xml.parser(tree)
    parser:parse(html)

    local htable = tree.root.div.table
    

    local trs = htable.tr
    local tr_count = #trs
    local data = {}
    for i = 1, tr_count do
        local tr = trs[i]
        local tds = tr.td
        
        
        local one = {}
        one.index = tds[1][1]
        one.name_a = tds[2][1]
        one.code_a = tds[3][1]
        one.ch_a = tds[4][1]
        one.close_a = tds[5][1]
        one.pe_a = tds[6][1]
        one.pb_a = tds[7][1]
        one.vol_a = tds[8][1]
        one.code_h = tds[9][1]
        one.ch_h = tds[10][1]
        one.close_h = tds[11][1]
        one.pe_h = tds[12][1]
        one.pb_h = tds[13][1]
        one.vol_h = tds[14][1]

        print(one.name, one.code_a, one.code_h)
       
        local one = {}
        data[#data + 1] = one
        
    end -- for tr end
    
    print(html)
end


local opts = {}
opts.browser = "gorilla"
M:request(opts)

-------------------------------------------------------------------------------------------

function M:keygen(opts)
    local key = string.format("%s.stock.group.flow.%s.%s.%d", opts.datasrc, opts.request_field, opts.request_order, opts.request_board)
    return key
end

-------------------------------------------------------------------------------------------

function M:persist(opts, data)

    local dates = global.calendar.List(0, 0, 0, true)

    local db = opts.db
    local datestr = dates[1]
    local dao = global.daom.Get("main")

    local n = #data
    local key = self:keygen(opts)
    local jsonstr = json.encode(data)
    local _, err = dao.Update(db, datestr, key, jsonstr, true, 0, nil)
    if err == nil then
        print("[persist] success", datestr, key, n)
    else
        print("[persist] failure", datestr, key, err)
    end
end

-------------------------------------------------------------------------------------------

function M:reload(opts, data, as_array)

    if data == nil then
        data = { }
    end

    if opts.date_offset == nil then
        opts.date_offset = 0
    end

    if as_array == nil then
        as_array = true
    end

    local dates = global.calendar.List(0, opts.date_offset, 0, true)
    local datestr = dates[1]
    local db = opts.db
    local dao = global.daom.Get("main")

    local key = self:keygen(opts)
    local datastr, err = dao.Get(db, datestr, key, 0, nil)
    if err ~= nil then
        print("[reload] failure", datestr, key, err)
    end
    if datastr == nil or #datastr == 0 then
        print("[reload] empty", datestr, key)
    else
        data = json.decode(datastr)
    end
    print("[reload]", datestr, key, #data)
    return data

end


function M:reloads(opts)

    local date_offset_from = opts.date_offset_from
    local date_offset_to = opts.date_offset_to
    if (date_offset_from == nil or date_offset_from > 0) then
        date_offset_from = 0
    end
    if (date_offset_to == nil or date_offset_to < 0) then
        date_offset_to = 0
    end

    local from = opts.date_offset + date_offset_from
    local to = opts.date_offset + date_offset_to
    local currindex = -date_offset_from + 1

    if to > 0 then
        to = 0
    end

    local data_maps = { }
    local opts_clone = opts
    for date_offset = from, to do
        opts_clone = simple.table_clone(opts_clone)
        opts_clone.date_offset = date_offset
        local as_array = date_offset == opts.date_offset
        local data = self:reload(opts_clone, nil, as_array)
        data_maps[#data_maps + 1] = data
    end

    local data_curr = data_maps[currindex]

    local data_curr_count = #data_curr
    local data_maps_count = #data_maps

    local code_mapping

    if data_maps_count > 1 then

        code_mapping = { }

        for i = 1, data_curr_count do
            local one_curr = data_curr[i]
            local code = one_curr.code

            local mapping_array = { }
            code_mapping[code] = mapping_array

            for n = 1, data_maps_count do
                if n == currindex then
                    mapping_array[#mapping_array + 1] = one_curr
                else
                    local map = data_maps[n]
                    local one_near = map[code]
                    if one_near ~= nil then
                        mapping_array[#mapping_array + 1] = one_near
                    end
                end
            end
        end

    end

    --[[
    for k , v in pairs(code_mapping) do
        print(k, #v)
    end
    ]]--

    return data_curr, code_mapping
end


-------------------------------------------------------------------------------------------

function M:filter(opts, data_curr, code_mapping, result)

    if result == nil then
        result = { }
    end

    local filters = opts.filters
    if filters == nil or #filters == 0 then
        --print("no filter?")
        return data_curr
    end
    --print("filter?!", #filters)

    local data_curr_index = opts.date_offset_from
    if data_curr_index == nil then
        data_curr_index = 0
    else
        data_curr_index = -data_curr_index + 1
    end

    local filters_count = #filters
    local data_curr_count = #data_curr
    for i = 1, data_curr_count do
        local one_curr = data_curr[i]
        local code = one_curr.code
        local series
        if code_mapping ~= nil then
            series = code_mapping[code]
        end

        local include = true
        for f = 1, filters_count do
            local filter = filters[f]
            include = filter(one_curr, series, code, data_curr_index, opts)
            if not include then
                break
            end
        end
        if include then
            result[#result + 1] = one_curr
        end

    end

    return result

end

-------------------------------------------------------------------------------------------


function M:data_merge(opts, result_curr, code_mapping)

    print("[merge] result curr", #result_curr, "code mapping", code_mapping ~= nil)

    if code_mapping == nil then
        return result_curr
    end

    local result_merge = { }
    local result_curr_count = #result_curr
    for i = 1, result_curr_count do

        local one_curr = result_curr[i]
        local code = one_curr.code
        local series = code_mapping[code]
        simple.array_append(result_merge, series)
    end

    return result_merge

end


-------------------------------------------------------------------------------------------

function M:link_stock_group(opts, data)
    if self.mod_stock_group == nil then
        self.mod_stock_group = require("sync.th.mod_stock_group")
    end
    local mapping = self.mod_stock_group:code_group_mapping(false)
    local n = #data
    for i = 1, n do
        local one = data[i]
        local code = one.code
        one.group = mapping[code]
    end
    return data
end

-------------------------------------------------------------------------------------------
function M:print_data(opts, data)

    local fields =
    {
        "index", "code", "name", "change_rate", "turnover",
        "flow_io_rate", "flow_in_rate",
        "flow_big_in_rate", "flow_big_rate", "flow_big_rate_total", "flow_big_rate_compare",
        "flow_big_rate_cross", "flow_big_rate_cross_ex", "flow_big", "group"
    }

    local headers =
    {
        "i", "code", "name", "ch", "turn",
        "io", "in",
        "big_in", "big_r", "big_t", "big_c",
        "cross", "crossex", "big", "group"
    }

    local formatters = {
        group = function(one, field, v)
            local r = ""
            for groupname in pairs(v) do
                r = r .. groupname .. " "
            end
            return r
        end
    }

    if opts.print_fields ~= nil then
        fields = opts.print_fields
    end

    if opts.print_headers ~= nil then
        headers = opts.print_headers
    end

    local from = opts.print_data_from
    local to = opts.print_data_to
    if from == nil then
        from = 1
    end
    if to == nil then
        to = #data
    end

    local key
    local header_interval = opts.date_offset_to - opts.date_offset_from + 1
    if header_interval > 1 then
        key = "code"
    end
    simple.table_array_print_with_header(data, from, to, fields, headers, 10, "\n", "", formatters, key)

end



------------------------------------------------------------------------------------------

function M:go(opts)


    local data_curr = self:request(opts)
    
    return opts.data, opts.result
end

--return M