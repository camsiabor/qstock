

-- http://data.10jqka.com.cn/funds/gnzjl/field/tradezdf/order/desc/page/2/ajax/1/


--[[
     1序号
    2股票代码
    3股票简称
    4最新价
    5涨跌幅
    6换手率
    7流入资金(元)
    8流出资金(元)
    9净额(元)
    10成交额(元)
    11大单流入(元)
]]--

local M = {}
M.__index = M

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

    local url_pattern = "http://data.10jqka.com.cn/funds/gnzjl/board/%d/field/%s/order/%s/page/%d/ajax/1/"
    local count = 1
    local reqopts = {}
    for page = opts.request_from, opts.request_to do
        local url = string.format(url_pattern, opts.request_board, opts.request_field, opts.request_order, page)
        local reqopt = {}
        reqopt["url"] = url
        reqopts[count] = reqopt
        count = count + 1
    end

    count = count - 1

    local err
    local browser = global[opts.browser]
    reqopts, err = browser.Get(reqopts, opts.nice, opts.newsession, opts.concurrent, opts.loglevel)

    if err ~= nil then
        print("[request] fatal", err)
    end

    local data = { }
    for i = 1, count do
        local reqopt = reqopts[i]
        self:parse_html(opts, data, reqopt)
    end
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
    if self.htmlparser == nil then
        self.htmlparser = require("common.htmlparser.htmlparser")
    end

    --[[
    1 序号
    2 行业
    3 行业指数
    4 涨跌幅
    5 流入资金(亿)
    6 流出资金(亿)
    7 净额(亿)
    8 公司家数
    9 领涨股
    10 涨跌幅
    11 当前价(元)
    ]]--

    local root = self.htmlparser.parse(html)
    if root == nil then
        print("[parse] failure", url)
        return
    end
    local tbody = root:select("tbody")[1]
    local tr_count = #tbody.nodes
    for i = 1, tr_count do
        local tr = tbody.nodes[i]
        local tds = tr.nodes
        local index = tds[1]:getcontent()
        local code = tds[2].nodes[1]:getcontent()
        local name = tds[3]:getcontent()
        local ch = tds[4]:getcontent()
        local flow_in = tds[5]:getcontent()
        local flow_out = tds[6]:getcontent()
        local flow = tds[7]:getcontent()

        local one = {}
        one.index = index + 0
        one.code = code
        one.name = name
        one.ch = simple.percent2num(ch)
        one.flow_in = flow_in + 0
        one.flow_out = flow_out + 0
        one.flow = flow + 0

        if opts.request_board == 1 then
            local company = tds[8]:getcontent()
            local head = tds[9].nodes[1]:getcontent()
            local head_ch = tds[10]:getcontent()
            local head_close = tds[11]:getcontent()
            one.company = company + 0
            one.head = head
            one.head_ch = simple.percent2num(head_ch)
            one.head_close = head_close + 0
        end

        data[#data + 1] = one
    end -- for tr end
end

-------------------------------------------------------------------------------------------

function M:keygen(opts, page)
    local key = string.format("%s.%s.%s.%d.%d", opts.datasrc, opts.request_field, opts.request_order, opts.request_board, page)
    return key
end

-------------------------------------------------------------------------------------------

function M:persist(opts, data)

    local dates = global.calendar.List(0, 0, 0, true)

    local db = opts.db
    local datestr = dates[1]
    local dao = global.daom.Get("main")

    local page = 1
    local pageone = {}

    local n = #data
    print("[persist]", datestr , "data count", n)
    for i = 1, n do
        pageone[#pageone + 1] = data[i]
        if (i % 50 == 0) or (i == n) then
            local jsonstr = json.encode(pageone)

            local key = self:keygen(opts, page)
            local _, err = dao.Update(db, datestr, key, jsonstr, true, 0, nil)
            if err == nil then
                if opts.debug then
                    print("[persist]", datestr, key, #pageone)
                end
            else
                print("[persist] failure", err)
            end
            page = page + 1
            pageone = {}
        end
    end
    print("[persist] fin")
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


    local total = 0
    for page = opts.request_from, opts.request_to do
        local key = self:keygen(opts, page)
        local datastr, err = dao.Get(db, datestr, key, 0, nil)
        if err ~= nil then
            print("[reload] failure", datestr, key, err)
        end
        if datastr == nil or #datastr == 0 then
            print("[reload] empty", datestr, key)
        else
            local fragment = json.decode(datastr)
            local n = #fragment

            if as_array then
                for i = 1, n do
                    local one = fragment[i]
                    data[#data + 1] = one
                end
            else
                for i = 1, n do
                    local one = fragment[i]
                    local code = one.code
                    data[code] = one
                end
            end
            total = total + n
        end
    end
    print("[reload]", datestr, total, "as array", as_array)
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
    local mapping = self.mod_stock_group:code_group_mapping(true)
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

    local data_curr, code_mapping
    if opts.request then
        data_curr = self:request(opts)
        if opts.persist then
            self:persist(opts, data_curr)
        end
    else
        data_curr, code_mapping = self:reloads(opts)
    end

    if simple.is(opts.link_stock_group) then
        data_curr = M:link_stock_group(opts, data_curr)
    end

    local result_curr = self:filter(opts, data_curr, code_mapping)

    simple.table_sort(result_curr, opts.sort_field)

    local result = self:data_merge(opts, result_curr, code_mapping)

    if opts.print_data == nil then
        self:print_data(opts, result)
    else
        simple.func_call(opts.print_data, opts, result)
    end

    opts.data = data_curr
    opts.result = result

    return opts.data, opts.result
end

return M