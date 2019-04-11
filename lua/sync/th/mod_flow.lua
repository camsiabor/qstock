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


local global = require("q.global")
local json = require('common.json')
local cal = require("common.cal")
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

    local each = opts.request_each
    if each == nil or each <= 1 then
        each = opts.request_to
    end
    local url_prefix = "http://data.10jqka.com.cn/funds/ggzjl/field/"..opts.field.."/order/"..opts.order.."/page/"
    local url_suffix = "/ajax/1"
    local data = { }
    local from = opts.request_from
    local limit = opts.request_to
    while from <= limit do
        local to = from + each - 1
        if to > limit then
            to = limit
        end
        local count = 1
        local reqopts = {}
        for page = from, to do
            local url = url_prefix..page..url_suffix
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
        for i = 1, count do
            local reqopt = reqopts[i]
            self:parse_html(opts, data, reqopt)
        end
        from = from + each
    end
    print("[data] count", #data)
    return data
end

-------------------------------------------------------------------------------------------
function M:parse_html(opts, data, reqopt)
    local url = reqopt["url"]
    local html = reqopt["content"]

    if html == nil or #html <= 500 then
        print("[error] request failure ---------------------------------")
        print(url)
        print(html)
        local err = reqopt["err"]
        if err ~= nil then
            print(err)
        end
        return
    end

    if self.xml == nil then
        self.xml = require("common.xml2lua.xml2lua")
        self.xml_tree_handler = require("common.xml2lua.tree")
    end

    local tree = self.xml_tree_handler:new()
    local parser = self.xml.parser(tree)
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

        flow = cal.str2num(flow)
        flow_in = cal.str2num(flow_in)
        flow_out = cal.str2num(flow_out)
        flow_big = cal.str2num(flow_big)

        flow_in = simple.nozero(flow_in)
        flow_out = simple.nozero(flow_out)
        flow_big = simple.nozero(flow_big)

        amount = cal.str2num(amount)

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



    local dates = global.calendar.List(0, 0, 0, true)

    local db = opts.db
    local datestr = dates[1]
    local dao = global.daom.Get("main")

    local page = 1
    local pageone = {}

    local slicify = false
    local timeslice = nil
    local datestr_cur = simple.now_date()
    if datestr_cur == datestr then
        local hm_cur = simple.now_hour_min() + 0
        slicify = (hm_cur >= 930 and hm_cur <= 1130) or (hm_cur >= 1300 and hm_cur <= hm_cur <= 1500)
        if slicify then
            local timearray = opts.persist_time_array
            if timearray == nil then
                timearray = { "0930", "1000", "1030", "1100", "1130", "1315", "1330", "1400", "1430", "1445", "1500" }
            end
            local timeslice = simple.num_array_align(timearray, hm_cur)
            if timeslice == nil then
                slicify = false
            end
        end
    end

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
                    print("[persist]", datestr, key, page, #pageone)
                end
            else
                print("[persist] failure", datestr, key, page, err)
            end

            if slicify then
                key = key .. "." .. timeslice
                local _, sliceerr = dao.Update(db, datestr, key, jsonstr, true, 0, nil)
                if sliceerr == nil then
                    print("[persist] slice", datestr, key, page, #pageone)
                else
                    print("[persist] slice failure", datestr, key, page, err)
                end
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
    local emptystr = ""
    for page = opts.request_from, opts.request_to do
        local key = self:keygen(opts, page)
        local datastr, err = dao.Get(db, datestr, key, 0, nil)
        if err ~= nil then
            print("[reload] failure", datestr, key, err)
        end
        if datastr == nil or #datastr == 0 then
            emptystr = string.format("[reload] empty %s %s", datestr, key)
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
    if #emptystr > 0 then
        print(emptystr)
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
                    if one_near == nil then
                        one_near = { code = code, name = one_curr.name, empty = true }
                    end
                    mapping_array[#mapping_array + 1] = one_near
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
        return data_curr
    end

    local data_curr_index = opts.date_offset_from
    if data_curr_index == nil then
        data_curr_index = 0
    else
        data_curr_index = -data_curr_index + 1
    end


    local includes
    local scans = data_curr
    local filters_count = #filters
    for f = 1, filters_count do
        includes = {}
        local n = #scans
        local filter = filters[f]
        for i = 1, n do
            local series
            local one = scans[i]
            local code = one.code
            if code_mapping ~= nil then
                series = code_mapping[code]
            end
            if filter(one, series, code, data_curr_index, opts) then
                includes[#includes + 1] = one
            end
        end
        scans = includes
    end

    local n = #includes
    for i = 1, n do
        result[#result + 1] = includes[i]
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

        if series ~= nil then
            local nseries = #series
            for n = 1, nseries do
                local serie = series[n]
                if serie ~= nil then
                    serie.group = one_curr.group
                end
            end
        end

        simple.array_append(result_merge, series)
    end

    return result_merge

end



-------------------------------------------------------------------------------------------

function M:link_stock_group(opts, data)
    if self.mod_stock_group == nil then
        self.mod_stock_group = require("sync.th.mod_stock_group")
    end
    if self.mod_ah_table == nil then
        self.mod_ah_table = require("sync.th.mod_ah_table_east")
    end
    local ah_table = self.mod_ah_table:reload({ reload_as_map = true } )

    local mapping = self.mod_stock_group:code_group_mapping(opts.stock_group_types, false)
    local n = #data
    for i = 1, n do
        local one = data[i]
        local code = one.code
        one.group = mapping[code]
        if one.group == nil then
            one.group = {}
        end
        local ah = ah_table[code]
        if ah ~= nil then
            local code_h = ah.code_h
            one.group["H股"] = ah
            one.group["H."..code_h] = code_h
        end
    end
    return data
end


-------------------------------------------------------------------------------------------
function M:link_stock_snapshot(opts, data, code_mapping)
    if self.mod_stock_snapshot == nil then
        self.mod_stock_snapshot = require("sync.th.mod_snapshot")
    end

    local dates = self.mod_stock_snapshot:dates(-opts.date_offset_from, opts.date_offset, opts.date_offset_to)
    local n = #data
    for i = 1, n do
        local one = data[i]
        local code = one.code
        local key = "ch" .. code
        local ks = self.mod_stock_snapshot:snapshot_by_cache(key, dates)
        local series
        if code_mapping == nil then
            series = { one }
        else
            series = code_mapping[code]
        end
        self.mod_stock_snapshot:merges(series, ks)
    end -- for

end

-------------------------------------------------------------------------------------------

function M:profile_result_stock_group(opts, data, code_mapping)
    local profiles = {}
    local n = #data
    for i = 1, n do
        local one = data[i]
        local io = one.flow_io_rate
        local ch = one.change_rate
        local big_in = one.flow_big_in_rate
        local groups = one.group
        if groups ~= nil then
            for gname in pairs(groups) do
                local profile = profiles[gname]
                if profile == nil then
                    profile = { name = gname, count = 0, sum_io = 0, sum_ch = 0, sum_big_in = 0 }
                    profiles[gname] = profile
                end
                profile.count = profile.count + 1
                profile.sum_io = profile.sum_io + io
                profile.sum_ch = profile.sum_ch + ch
                profile.sum_big_in = profile.sum_big_in + big_in
            end
        end
    end

    for _, profile in pairs(profiles) do
        local count = profile.count
        if count > 0 then
            profile.avg_io = simple.numcon(profile.sum_io / count)
            profile.avg_ch = simple.numcon(profile.sum_io / count)
            profile.avg_big_in = simple.numcon(profile.sum_big_in / count)
        end
    end
    return profiles
end



-------------------------------------------------------------------------------------------
function M:print_data(opts, data)

    local fields =
        {
            "index", "code", "name", "change_rate", "turnover",
            "flow_io_rate", "flow_in_rate",
            "flow_big_in_rate", "flow_big_rate",
            "flow_big_rate_cross", "flow_big_rate_cross_ex", "flow_big", "group"
        }

    local headers =
        {
            "i", "code", "name", "ch", "turn",
            "io", "in",
            "big_in", "big_r",
            "cross", "crossex", "big", "group"
        }

    local formatters = {
        --[[
        date = function(one, field, v)
            if v == nil or #v == 0 then
                return ""
            end
            return string.sub(v, 5)
        end,
        ]]--
        group = function(one, field, v, ikey)
            if ikey > 1 then
               return ""
            end
            local r = ""
            for groupname in pairs(v) do
                r = r .. groupname .. " "
            end
            return r
        end,
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

function M:print_stock_group_profile(opts, data)

    local fields =
    {
        "count", "avg_io", "avg_ch", "avg_big_in", "name"
    }

    local headers =
    {
        "count", "avg_io", "avg_ch", "avg_big", "name"
    }

    if opts.print_fields ~= nil then
        fields = opts.print_fields
    end

    if opts.print_headers ~= nil then
        headers = opts.print_headers
    end
    local from = 1
    local to = #data
    simple.table_array_print_with_header(data, from, to, fields, headers, 10, "\n", "")
end

------------------------------------------------------------------------------------------

function M:go(opts)

    if opts.stock_group_types == nil then
        opts.stock_group_types =  { "concept" }
    end

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
        self:link_stock_group(opts, data_curr, code_mapping)
    end

    if simple.is(opts.link_stock_snapshot) then
        --self:link_stock_snapshot(opts, data_curr, code_mapping)
    end

    local result_curr = self:filter(opts, data_curr, code_mapping)

    if opts.result_adapter ~= nil then
        local currindex = -opts.date_offset_from + 1
        local adapted = simple.func_call(opts.result_adapter, opts, result_curr, code_mapping, currindex)
        if adapted ~= nil then
            result_curr = adapted
        end
    end

    local group_profiles = self:profile_result_stock_group(opts, result_curr, code_mapping)
    group_profiles = simple.map_to_array(group_profiles)
    simple.table_sort(group_profiles, "count")
    print("[result.group.profile] ^ ---------------------------------------------------------------------------------------------")
    self:print_stock_group_profile(opts, group_profiles)
    print("[result.group.profile] # ---------------------------------------------------------------------------------------------")
    print("")


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


------------------------------------------------------------------------------------------


function M:profile_stock_group(opts)
    if opts.stock_group_types == nil then
        opts.stock_group_types =  { "concept" }
    end

    if opts.cal_limit == nil then
        opts.cal_limit = 300
    end

    if opts.n_io_criteria == nil then
        opts.n_io_criteria = {
            0.5, 0.75, 1, 1.2, 1.5, 1.7
        }
    end
    if opts.n_ch_criteria == nil then
        opts.n_ch_criteria = {
            -5, -2.5, 0, 2.5, 5, 7.5
        }
    end
    opts.date_offset_to = 0

    local data_curr, code_mapping = self:reloads(opts)
    if self.mod_stock_group == nil then
        self.mod_stock_group = require("sync.th.mod_stock_group")
    end
    -- TODO multi stock group request type
    local dates = global.calendar.List(-opts.date_offset_from, opts.date_offset, opts.date_offset_to, false)
    for i = 1, #dates do
        dates[i] = string.sub(dates[i], 5)
    end

    local daycount = opts.date_offset_to - opts.date_offset_from + 1
    local groups = self.mod_stock_group:list_reload( { request_type = opts.stock_group_types[1] } )

    for _, group in pairs(groups) do
        local gname = group.name
        local list = group.list
        if list == nil then
            list = {}
        end

        local nlist = 0
        for _ in pairs(list) do
            nlist = nlist + 1
        end

        group.nlist = nlist
        group.profiles = { }

        for i = 1, daycount do
            local profile = {
                name = gname,
                date = dates[i],
                count = 0,
                io = 0, fin = 0, ch = 0, big_in = 0,
                avg_io = 0, avg_fin = 0, avg_ch = 0, avg_big_in = 0,
                n_io = { }, n_ch = { }
            }
            cal.num_level_criteria_init(opts.n_io_criteria, profile.n_io)
            cal.num_level_criteria_init(opts.n_ch_criteria, profile.n_ch)
            group.profiles[#group.profiles + 1] = profile
        end

        if nlist <= opts.cal_limit then
            for code in pairs(list) do
                local series = code_mapping[code]
                if series == nil then
                    series = {}
                end
                for i = 1, daycount do
                    local serie = series[i]

                    if serie ~= nil then
                        local profile = group.profiles[i]
                        local io = serie.flow_io_rate
                        local fin = serie.flow_in_rate
                        local ch = serie.change_rate
                        local big_in = serie.flow_big_in_rate

                        if io <= 5 then
                            profile.count = profile.count + 1
                            profile.io = profile.io + io
                            profile.fin = profile.fin + fin
                            profile.ch = profile.ch + ch
                            profile.big_in = profile.big_in + big_in
                            cal.num_level_criteria_count(io, opts.n_io_criteria, profile.n_io)
                            cal.num_level_criteria_count(ch, opts.n_ch_criteria, profile.n_ch)
                        end
                    end
                end
            end
        end

        for i = 1, daycount do
            local profile = group.profiles[i]
            if profile.count > 0 then
                profile.avg_io = simple.numcon(profile.io / profile.count)
                profile.avg_fin = simple.numcon(profile.fin / profile.count)
                profile.avg_ch = simple.numcon(profile.ch / profile.count)
                profile.avg_big_in = simple.numcon(profile.big_in / profile.count)

                cal.array_div_mul(profile.n_io, profile.count, 100, simple.numcon)
                cal.array_div_mul(profile.n_ch, profile.count, 100, simple.numcon)
            end
        end
    end

    return groups
end

----------------------------------------------------------------------------------------------------------------------------------------
function M:go_stock_group_profile(opts)

    local groups = self:profile_stock_group(opts)
    local grouparray = simple.map_to_array(groups)
    local daycount = opts.date_offset_to - opts.date_offset_from + 1
    simple.table_sort(grouparray, { "profiles", daycount, opts.sort_field })

    local profiles = {}
    for i = 1, #grouparray do
        local group = grouparray[i]
        simple.array_append(profiles, group.profiles)
    end

    local n_io_criteria_len = #opts.n_io_criteria
    local n_ch_criteria_len = #opts.n_ch_criteria

    local headers = {
        "date", "count", "io", "ch", "big",
        "io1.2", "io1.5", "io1.7",
        "ch2.5", "ch5", "ch7.5",
        "name"
    }

    local fields = {
        "date", "count", "avg_io", "avg_ch", "avg_big_in",
        { "n_io", n_io_criteria_len - 2 }, { "n_io", n_io_criteria_len - 1 }, { "n_io", n_io_criteria_len },
        { "n_ch", n_ch_criteria_len - 2 }, { "n_ch", n_ch_criteria_len - 1 }, { "n_ch", n_ch_criteria_len },
        "name"
    }


    simple.table_array_print_with_header(profiles,
            1, #profiles,
            fields, headers, 1,
            "\n", "", nil, "name")

    return opts.data, opts.result
end


-----------------------------------------------------------------------------------------------------------------------

return M