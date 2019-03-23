-- http://q.10jqka.com.cn/gn/detail/code/301365/
-- http://q.10jqka.com.cn/gn/detail/field/264648/order/desc/page/2/ajax/1/code/304582
-- http://q.10jqka.com.cn/gn/detail/field/3475914/order/desc/page/1/ajax/1/code/301558

local M = {}
M.__index = M

local xml, xml_tree_handler
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

function M:group_request(opts)
    local url = "http://q.10jqka.com.cn/gn"

    local reqopts = {}
    local reqopt = {}
    reqopt["url"] = url
    reqopt["encoding"] = "gbk"
    reqopts[1] = reqopt

    local err
    local browser = global["gorilla"]
    if browser == nil then
        browser = global[opts.browser]
    end
    reqopts, err = browser.Get(reqopts, opts.nice, opts.newsession, opts.concurrent, opts.loglevel)
    if err ~= nil then
        print("[request] fatal", err)
    end

    reqopt = reqopts[1]
    local html = reqopt["content"]
    local groups = self:group_parse(html)
    return groups
end

function M:group_parse(html)
    local tag_start = '<div class="category boxShadow m_links">'
    local tag_end = '<div class="cate_toggle_wrap">'
    local index = string.find(html, tag_start)
    if index == nil then
        print("[group] [request] failure")
        print(html)
        return
    end
    html = string.sub(html, index)
    index = string.find(html, tag_end)
    html = string.sub(html, 1, index - 1)


    local groups = {}
    local pattern = '<a href="http://q.10jqka.com.cn/gn/detail/code/(%d+)/" target="_blank">(%W+)</a>'
    local iterator = string.gmatch(html, pattern)
    for code, name in iterator do
        local group = { code = code, name = name, list = { } }
        groups[code] = group
    end
    return groups
end


function M:list_request(opts, groups)

    local url_pattern = "http://q.10jqka.com.cn/gn/detail/field/3475914/order/desc/page/%d/ajax/1/code/%d"
    local count = 0
    local reqopts = {}

    for code, group in pairs(groups) do
        local page = group.page
        if page == nil or page == 0 then
            page = 1
        end
        local from, to, browser
        if page == 1 then
            from = 1
            to = 1
        else
            from = 2
            to = page
        end

        for p = from, to do
            local url = string.format(url_pattern, p, code)
            local reqopt = {}
            reqopt["url"] = url
            reqopt["page"] = page
            reqopt["group"] = group
            reqopt["encoding"] = "gbk"
            count = count + 1
            reqopts[count] = reqopt

        end
    end

    local err
    local browser = global["gorilla"]
    if browser == nil then
        browser = global[opts.browser]
    end
    reqopts, err = browser.Get(reqopts, opts.nice, opts.newsession, opts.concurrent, opts.loglevel)

    if err ~= nil then
        print("[request] fatal", err)
    end

    for i = 1, #reqopts do
        local reqopt = reqopts[i]
        local group = reqopt.group
        self:list_parse(opts, reqopt, group)
    end

    return groups
end

function M:list_parse(opts, reqopt, group)

    local html = reqopt["content"]
    if html == nil then
        local err = reqopt["err"]
        print("[request] [list] fail", reqopt["url"])
        print(err)
        print(reqopt)
        return
    end

    local table_start = '<table class="m%-table m%-pager%-table">'
    local table_end = '</table>'

    local i_table_start = string.find(html, table_start)
    if i_table_start == nil then
        print("[request] [list] failure", reqopt["url"])
        print(html)
        return
    end
    local html_table = string.sub(html, i_table_start)
    local i_table_end = string.find(html_table, table_end)
    html_table = string.sub(html_table, 1, i_table_end + #table_end)


    print(html_table)

    local pattern_td = '<td>(%W+)</td>'
    local iterater = string.gmatch(html_table, pattern_td)
    for one in iterater do
        print(one)
    end

    local page_start = '<span class="page_info">'
    local page_end = '</span>'
    local i_page_start = string.find(html, page_start, i_table_end + #table_end + 1)
    local i_page_end = string.find(html, page_end, i_page_start + #page_start + 1)
    local page_count = string.sub(html, i_page_start + #page_start + 2, i_page_end - 1)

    return groups
end


function M:go(opts)
    local data = opts.data
    local result = opts.result
    if data == nil then
        data = {}
    end
    if result == nil then
        result = {}
    end
    opts.browser_original = opts.browser
    opts.browser = "gorilla"
    local groups = self:group_request(opts)
    groups = self:list_request(opts, groups, 1)



    return data, result
end


local opts = {}
opts.debug = false
opts.loglevel = 0
opts.browser = "firefox"
local groups = M:group_request(opts)
--simple.table_print_all(groups)
M:list_request(opts, groups)


-------------------------------------------------------------------------------------------
function M:parse_html(opts, data, result, reqopt)
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

    local dates = global.calendar.List(0, 0, 0, true)

    local db = opts.db
    local datestr = dates[1]
    local dao = global.daom.Get("main")

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

function M:reload(opts, data, result)

    if opts.date_offset == nil then
        opts.date_offset = 0
    end

    local dates = global.calendar.List(0, opts.date_offset, 0, true)
    local datestr = dates[1]
    local db = opts.db
    local dao = global.daom.Get("main")

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
                data[#data + 1] = fragment[i]
            end
        end

    end

end


-------------------------------------------------------------------------------------------
function M:filter(opts, data, result)

    local n = #data
    for i = 1, n do
        local one = data[i]
        local critical = one.change_rate >= opts.ch_lower and one.change_rate <= opts.ch_upper
        if critical then
            critical = one.flow_big_rate_compare >= opts.big_c_lower and one.flow_big_rate_compare <= opts.big_c_upper
            if critical then
                result[#result + 1] = one
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

--return M