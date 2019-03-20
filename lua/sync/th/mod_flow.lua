-- http://data.10jqka.com.cn/funds/ggzjl/
-- http://data.10jqka.com.cn/funds/ggzjl/field/zjjlr/order/desc/page/1/ajax/1/



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

local xml, xml_tree_handler
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

    local url_prefix = "http://data.10jqka.com.cn/funds/ggzjl/field/"..opts.field.."/order/"..opts.order.."/page/"
    local url_suffix = "/ajax/1"

    local count = 1
    local reqopts = {}
    for page = opts.from, opts.to do
        local url = url_prefix..page..url_suffix
        local reqopt = {}
        reqopt["url"] = url
        reqopts[count] = reqopt
        count = count + 1
    end

    count = count - 1

    local err
    local browser = Q[opts.browser]
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
        print("[error] request failure")
        print(url)
        print(reqopt["err"])
        return
    end

    if xml == nil then
        xml = require("common.xml2lua.xml2lua")
        xml_tree_handler = require("common.xml2lua.tree")
    end

    local tree = xml_tree_handler:new()
    local parser = xml.parser(tree)
    parser:parse(html)


    local htable = tree.root.html.body.table
    if htable == nil then
        print("[error] response content invalid "..#html)
        print(url)
        print(html)
        print("")
        return
    end

    local tbody = htable.tbody

    local tr_count = #tbody.tr

    for i = 1, tr_count do

        local tr = tbody.tr[i]
        local index = tr.td[1][1]
        local code = tr.td[2].a[1]
        local name = tr.td[3].a[1]
        local change_rate = tr.td[5][1]
        local turnover = tr.td[6][1]
        local flow_in = tr.td[7][1]
        local flow_out = tr.td[8][1]
        local flow = tr.td[9][1]
        local amount = tr.td[10][1]
        local flow_big = tr.td[11][1]

        turnover = string.gsub(turnover, "%%", "") + 0
        change_rate = string.gsub(change_rate, "%%", "") + 0

        flow = simple.str2num(flow)
        flow_in = simple.str2num(flow_in)
        flow_out = simple.str2num(flow_out)
        flow_big = simple.str2num(flow_big)

        flow_in = simple.nozero(flow_in)
        flow_out = simple.nozero(flow_out)
        flow_big = simple.nozero(flow_big)

        amount = simple.str2num(amount)

        if amount <= 0 then
            amount = 0.0001
        end

        local flow_big_rate = flow_big / amount * 100
        local flow_big_rate_compare = flow / flow_big
        local flow_big_rate_total = turnover * flow_big_rate / 100

        local flow_in_rate = flow_in / amount * 100
        local flow_out_rate = flow_out / amount * 100
        local flow_io_rate = flow_in / flow_out

        local flow_big_in_rate = flow_big / flow_in * 100

        --local flow_big_rate_cross = (turnover * amount * flow_big_rate / 100) * flow_io_rate * flow_big_in_rate
        local flow_big_rate_cross = flow_io_rate * flow_big_rate_total * flow_big_rate / 100 * flow_big_in_rate
        local change_rate_ex = change_rate
        if change_rate_ex < 0 then
            change_rate_ex = 0.1
        end
        local flow_big_rate_cross_ex = flow_big_rate_cross / (change_rate_ex + 2.5)

        flow_big_rate = simple.numcon(flow_big_rate)
        flow_big_rate_compare = simple.numcon(flow_big_rate_compare)
        flow_big_rate_total = simple.numcon(flow_big_rate_total)
        flow_big_rate_cross = simple.numcon(flow_big_rate_cross)
        flow_big_rate_cross_ex = simple.numcon(flow_big_rate_cross_ex)

        flow_in_rate = simple.numcon(flow_in_rate)
        flow_out_rate = simple.numcon(flow_out_rate)
        flow_io_rate = simple.numcon(flow_io_rate)

        flow_big_in_rate = simple.numcon(flow_big_in_rate)

        local one = {}
        one.index = index
        one.code = code
        one.name = name
        one.flow = flow
        one.flow_in = flow_in
        one.flow_out = flow_out
        one.amount = amount
        one.turnover = turnover
        one.flow_big = flow_big
        one.change_rate = change_rate

        one.flow_big_rate = flow_big_rate
        one.flow_big_rate_total = flow_big_rate_total
        one.flow_big_rate_compare = flow_big_rate_compare
        one.flow_big_rate_cross = flow_big_rate_cross
        one.flow_big_rate_cross_ex = flow_big_rate_cross_ex

        one.flow_in_rate = flow_in_rate
        one.flow_out_rate = flow_out_rate
        one.flow_io_rate = flow_io_rate

        one.flow_big_in_rate = flow_big_in_rate

        data[#data + 1] = one


    end -- for tr end
end


-------------------------------------------------------------------------------------------

function M:keygen(opts, page)
    local key = string.format("%s.%s.%s.%d", opts.datasrc, opts.field, opts.order, page)
    return key
end



-------------------------------------------------------------------------------------------

function M:persist(opts, data)

    local dates = Q.calendar.List(0, 0, 0, true)

    local db = opts.db
    local datestr = dates[1]
    local dao = Q.daom.Get("main")

    local page = 1
    local pageone = {}
    local pagesize = 50

    local n = #data
    print("[persist]", datestr , "data count", n)
    for i = 1, n do
        pageone[#pageone + 1] = data[i]
        if (i % 50 == 0) or (i == n) then
            local jsonstr = json.encode(pageone)
            --print(jsonstr)
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

    local dates = Q.calendar.List(0, opts.date_offset, 0, true)
    local datestr = dates[1]
    local db = opts.db
    local dao = Q.daom.Get("main")

    print("[reload]", datestr)
    for page = opts.from, opts.to do
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

            for i = 1, n do
                local one = fragment[i]
                if as_array then
                    data[#data + 1] = one
                else
                    local code = one.code
                    data[code] = one
                end
            end
        end

    end

    return data

end


function M:reloads(opts)

    local date_offset_from = opts.date_offset_from
    local date_offset_to = opts.date_offset_to
    if (date_offset_from == nil or date_offset_from <= 0) then
        date_offset_from = 0
    end
    if (date_offset_to == nil or date_offset_to <= 0) then
        date_offset_to = 0
    end

    local from = opts.date_offset + date_offset_to
    local to = opts.date_offset + date_offset_from
    local currindex = (to - from) - date_offset_from + 1

    if to > 0 then
        to = 0
    end

    local opts_clone = opts
    local data_maps = { }
    for date_offset = from, to do
        opts_clone = simple.table_clone(opts_clone)
        opts_clone.date_offset = date_offset
        local as_array = date_offset == opts.date_offset
        local data = self:reload(opts_clone, nil, as_array)
        data_maps[#data_maps + 1] = data
    end

    local data_curr = data_maps[currindex]

    local data_curr_count = #data_curr
    local data_series_num = #data_maps


    local code_mapping

    if data_series_num > 1 then

        code_mapping = { }

        for i = 1, data_curr_count do
            local one_curr = data_curr[i]
            local code = one_curr.code

            local mapping_array = { }
            code_mapping[code] = mapping_array

            for n = 1, data_series_num do
                if n ~= currindex then
                    local map = data_maps[n]
                    local one_in_map = map[code]
                    if one_in_map ~= nil then
                        mapping_array[#mapping_array + 1] = one_in_map
                    end
                end
            end
        end

    end


    return data_curr, code_mapping
end


-------------------------------------------------------------------------------------------

function M:filter(opts, data_curr, code_mapping, result)

    if result == nil then
        result = { }
    end

    local filter = opts.filter
    local data_curr_count = #data_curr
    for i = 1, data_curr_count do
        local one_curr = data_curr[i]
        local code = one_curr.code
        local series
        if code_mapping ~= nil then
            series = code_mapping[code]
        end

        local include = filter(one_curr, series, code, opts)
        if include then
            result[#result + 1] = one_curr
        end
    end

    return result

end

-------------------------------------------------------------------------------------------


function M:merge_series(opts, result_curr, code_mapping)

    if code_mapping == nil then
        return result_curr
    end

    local result_curr_count = #result_curr
    for i = 1, result_curr_count do
        local one_curr = result_curr[i]
        local code = one_curr
    end

    if series ~= nil then
        for n = 1, #series do
            local one_serie = series[n]
            if one_serie ~= nil then
                result[#result + 1] = one_serie
            end
        end
    end

end


-------------------------------------------------------------------------------------------
function M:print_data(opts, data)

    local fields =
        {
            "index", "code", "name", "change_rate", "turnover",
            "flow_io_rate", "flow_in_rate",
            "flow_big_in_rate", "flow_big_rate", "flow_big_rate_total", "flow_big_rate_compare",
            "flow_big_rate_cross", "flow_big_rate_cross_ex", "flow_big"
        }

    local headers =
        {
            "i", "code", "name", "ch", "turn",
            "io", "in",
            "big_in", "big_r", "big_t", "big_c",
            "cross", "crossex", "big"
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

    simple.table_array_print_with_header(data, from, to, fields, headers, 10, "\n")
end



------------------------------------------------------------------------------------------

function M:go(opts)

    local data_curr, code_mapping
    if opts.dofetch then
        data_curr = self:request(opts)
        if opts.persist then
            self:persist(opts, data_curr)
        end
    else
        data_curr, code_mapping = self:reloads(opts)
    end

    local result_curr = self:filter(opts, data_curr, code_mapping)

    simple.table_sort(result_curr, opts.sort_field)

    local result = self:merge_series(opts, result_curr, code_mapping)

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