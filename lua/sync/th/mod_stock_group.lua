-- http://q.10jqka.com.cn/gn/detail/code/301365/
-- http://q.10jqka.com.cn/gn/detail/field/264648/order/desc/page/2/ajax/1/code/304582
-- http://q.10jqka.com.cn/gn/detail/field/3475914/order/desc/page/1/ajax/1/code/301558

local M = {}
M.__index = M
M.TOKEN_PERSIST_GROUP = "ch.stock.group.concept.index"
M.TOKEN_PERSIST_LIST = "ch.stock.group.concept"


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
    local groups = self:group_parse(opts, html)
    return groups
end

-----------------------------------------------------------------------------------------------------------------

function M:group_parse(opts, html)
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

    local count = 0
    for code, name in iterator do
        local group = { code = code, name = name, list = { } }
        groups[code] = group
        count = count + 1
    end
    return groups, count
end

-----------------------------------------------------------------------------------------------------------------

function M:group_persist(opts, groups)
    local db = opts.db
    if db == nil then
        db = "group"
    end
    local dao = global.daom.Get("main")

    local jsonstr = json.encode(groups)
    local _, err = dao.Update(db, self.TOKEN_PERSIST_GROUP, "", jsonstr, true, 0, nil)
    if err == nil then
        print("[persist] group")
    else
        print("[persist] group failure", err)
    end
end

-----------------------------------------------------------------------------------------------------------------

function M:group_reload(opts)
    local db = opts.db
    if db == nil then
        db = "group"
    end
    print("[reload] group")
    local dao = global.daom.Get("main")
    local datastr, err = dao.Get(db, self.TOKEN_PERSIST_GROUP, "", 0, nil)
    if err ~= nil then
        print("[reload] failure", self.TOKEN_PERSIST_GROUP, "", err)
    end
    local groups = json.decode(datastr)
    return groups
end

-----------------------------------------------------------------------------------------------------------------

function M:list_request(opts, groups)


    local count = 0
    local reqopts = {}

    local n = 1
    simple.def(opts, "request_from", 1)
    simple.def(opts, "request_to", 1000)
    for code, group in pairs(groups) do
        if n >= opts.request_from and n <= opts.request_to then

            local page = group.page
            if page == nil or page == 0 then
                page = 1
            end
            local from, to
            if page == 1 then
                from = 1
                to = 1
            else
                from = 2
                to = page

            end

            for p = from, to do

                local url
                if p == 1 then
                    url = string.format("http://q.10jqka.com.cn/gn/detail/code/%d/", code)
                else
                    url = string.format("http://q.10jqka.com.cn/gn/detail/field/3475914/order/desc/page/%d/ajax/1/code/%d", p, code)
                end

                local reqopt = {}
                reqopt["url"] = url
                reqopt["page"] = page
                reqopt["group"] = group
                reqopt["encoding"] = "gbk"
                count = count + 1
                reqopts[count] = reqopt
            end
        end
        n = n + 1
    end

    local err

    local browser = simple.get(global, "gorilla", global[opts.browser])
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


--[[
    1 序号
    2 代码
    3 名称
    4 现价
    5 涨跌幅(%)
    6 涨跌(%)
    7 涨速(%)
    8 换手(%)
    9 量比
    10 振幅(%)
    11 成交额
    12 流通股
    13 流通市值
    14 市盈率
]]--
function M:list_parse(opts, reqopt, group, data)

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

    if reqopt.page == nil or reqopt.page <= 1 then
        local page_start = '<span class="page_info">'
        local page_end = '</span>'
        local i_page_start = string.find(html, page_start, i_table_end + #table_end + 1)
        local i_page_end = string.find(html, page_end, i_page_start + #page_start + 1)
        local page_count = string.sub(html, i_page_start + #page_start + 2, i_page_end - 1)
        group.page = page_count
    end

    if self.htmlparser == nil then
        self.htmlparser = require("common.htmlparser.htmlparser")
    end

    local root = self.htmlparser.parse(html_table)
    local tbody = root:select("tbody")[1]
    local tr_count = #tbody.nodes
    for i = 1, tr_count do

        local tr = tbody.nodes[i]
        local tds = tr.nodes
        local code = tds[2].nodes[1]:getcontent()
        local name = tds[3].nodes[1]:getcontent()

        group.list[code] = name

    end -- for tr end

    return group
end

------------------------------------------------------------------------------------------------------------------------

function M:list_persist(opts, groups)
    local db = opts.db
    if db == nil then
        db = "group"
    end
    local dao = global.daom.Get("main")

    local n = 1
    for code, group in pairs(groups) do
        if n >= opts.request_from and n <= opts.request_to then
            local jsonstr = json.encode(group)
            local _, err = dao.Update(db, self.TOKEN_PERSIST_LIST, code, jsonstr, true, 0, nil)
            if err == nil then
                print("[persist] list", code)
            else
                print("[persist] list failure", err)
            end
        end
        n = n + 1
    end
end

------------------------------------------------------------------------------------------------------------------------

function M:list_reload(opts)

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


function M:go(opts)
    local data = opts.data
    local result = opts.result
    if data == nil then
        data = {}
    end
    if result == nil then
        result = {}
    end

    local groups, ngroups
    if opts.request then
        opts.browser_original = opts.browser
        opts.browser = "gorilla"
        groups, ngroups = self:group_request(opts)
        if opts.persist then
            self:group_persist(opts, groups)
        end
        self:list_request(opts, groups)
        opts.browser = opts.browser_original
        self:list_request(opts, groups)
        if opts.persist then
            self:list_persist(opts, groups)
        end
    else

    end

    return data, result
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

return M