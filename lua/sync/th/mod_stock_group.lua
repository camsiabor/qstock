-- http://q.10jqka.com.cn/gn/detail/code/301365/
-- http://q.10jqka.com.cn/gn/detail/field/264648/order/desc/page/2/ajax/1/code/304582
-- http://q.10jqka.com.cn/gn/detail/field/3475914/order/desc/page/1/ajax/1/code/301558

local M = {}
M.__index = M
M.TOKEN_PERSIST_GROUP = "ch.stock.group.concept.index"
M.TOKEN_PERSIST_LIST = "ch.stock.group.concept"


local global = require("q.global")
local logger = require("q.logger")
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
    local browser = global[opts.browser]
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
    print("[parse] group count", count)
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
            else
                page = page + 0
            end

            local from, to
            if page == 1 then
                from = 1
                to = 1
            else
                from = 1
                to = page
            end

            for p = from, to do
                local url
                if p == 1 and to == 1 then
                    url = string.format("http://q.10jqka.com.cn/gn/detail/code/%d/", code)
                else
                    url = string.format("http://q.10jqka.com.cn/gn/detail/field/3475914/order/desc/page/%d/ajax/1/code/%d", p, code)
                end

                local reqopt = {}
                reqopt["url"] = url
                reqopt["page"] = p
                reqopt["code"] = group.code
                reqopt["encoding"] = "gbk"
                count = count + 1
                reqopts[count] = reqopt
            end
        end
        n = n + 1
    end

    local err

    local browser = global[opts.browser]
    reqopts, err = browser.Get(reqopts, opts.nice, opts.newsession, opts.concurrent, opts.loglevel)

    if err ~= nil then
        print("[request] fatal", err)
    end

    for i = 1, #reqopts do
        local reqopt = reqopts[i]
        local code = reqopt.code
        local group = groups[code]
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

    if reqopt.page == nil or reqopt.page <= 1 then
        local page_start = '<span class="page_info">'
        local page_end = '</span>'
        local i_page_start = string.find(html, page_start, i_table_end + #table_end + 1)
        if i_page_start == nil then
            group.page = 1
            --logger:error("[parse] list page start not found", group.code, group.name)
        else
            local i_page_end = string.find(html, page_end, i_page_start + #page_start + 1)
            local page_count = string.sub(html, i_page_start + #page_start + 2, i_page_end - 1)
            group.page = page_count + 0
        end
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

function M:list_persist(opts, groups, log)
    local db = opts.db
    if db == nil then
        db = "group"
    end
    local dao = global.daom.Get("main")
    local n = 1
    for code, group in pairs(groups) do
        if n >= opts.request_from and n <= opts.request_to then
            local key = "ch" .. code
            local jsonstr = json.encode(group)
            local _, err = dao.Update(db, self.TOKEN_PERSIST_LIST, key, jsonstr, true, 0, nil)
            if err == nil then
                if log then
                    print("[persist] list", code, group.name, group.page)
                end
            else
                if log then
                    print("[persist] list failure", code, group.name, err)
                end
            end
        end
        n = n + 1
    end
end

------------------------------------------------------------------------------------------------------------------------

function M:list_reload(opts)
    local db = opts.db
    if db == nil then
        db = "group"
    end
    local dao = global.daom.Get("main")
    print("[reload] stock group concept")
    local map, err = dao.Get(db, "", self.TOKEN_PERSIST_LIST, 1, nil)
    if err ~= nil then
        print("[reload] failure", db, self.TOKEN_PERSIST_LIST, err)
        return
    end
    if map == nil or simple.table_count(map) == 0 then
        print("[reload] empty", db, self.TOKEN_PERSIST_LIST)
        return
    end
    local groups = {}
    for key in pairs(map) do
        local groupstr = map[key]
        if #groupstr > 0 then
            local group = json.decode(groupstr)
            local code = string.gsub(key, "ch", "")
            groups[code] = group
        end
    end
    return groups
end

function M:list_reload_non_complete(opts)
    local index = self:group_reload(opts)
    local groups = self:list_reload(opts)
    local todos = { }
    local count = 0

    for code, igroup in pairs(index) do
        local key = code
        local group = groups[key]
        local noncomplete = group == nil
        if noncomplete then
            print("[noncomplete] group nil", code)
        else
            if group.page == nil then
                group.page = 0
            else
                group.page = group.page + 0
            end
            noncomplete = group.page <= 0
            if noncomplete then
                print("[noncomplete] group page zero", code, group.name, group.page)
            else
                noncomplete = simple.table_count(group.list) <= 0
                if noncomplete then
                    print("[noncomplete] group list zero", code, group.name, group.page)
                end
            end
        end

        if not noncomplete then
            local suppose = (group.page - 1) * 10 + 1
            local currcount = simple.table_count(group.list)
            noncomplete = currcount < suppose
            if noncomplete then
                print("[noncomplete] group list not full", code, group.name, suppose, currcount)
            end
        end
        if noncomplete then
            todos[code] = igroup
            count = count + 1
        end
    end
    print("[noncomplete] count ", simple.table_count(todos))
    return todos
end

function M:list_code_group_mapping(groups)
    local mapping = {}
    for groupcode, group in pairs(groups) do
        for code, name in pairs(group.list) do
            local map = mapping[code]
            if map == nil then
                map = {}
                mapping[code] = map
                mapping[name] = map
            end
            map[group.name] = groupcode
        end
    end
    return mapping
end

function M:code_group_mapping(from_cache)
    local mapping
    local mappingstr
    local cachekey = "code.group.mapping"
    local cache = global.cachem.Get(self.TOKEN_PERSIST_LIST)
    if from_cache then
        local mappingstr = cache.Get(false, cachekey)
        if mappingstr ~= nil and #mappingstr > 0 then
            mapping = json.decode(mappingstr)
        end
    end
    if mapping == nil then
        local groups = self:list_reload({})
        mapping = self:list_code_group_mapping(groups)
        mappingstr = json.encode(mapping)
        cache.Set(mappingstr, cachekey)
    end
    return mapping
end

function M:go(opts)
    local groups
    if opts.request then
        opts.browser_original = opts.browser
        opts.browser = "gorilla"
        groups = self:group_request(opts)
        if opts.persist then
            self:group_persist(opts, groups)
        end

        self:list_request(opts, groups)
        if opts.persist then
            self:list_persist(opts, groups, false)
        end

        opts.browser = opts.browser_original
        self:list_request(opts, groups)
        if opts.persist then
            self:list_persist(opts, groups, true)
        end
    else
        if opts.reload_check then
            opts.browser_original = opts.browser
            opts.browser = "gorilla"
            groups = self:group_request(opts)
            if opts.persist then
                self:group_persist(opts, groups)
            end
            local todos = self:list_reload_non_complete(opts)
            self:list_request(opts, todos)
            opts.browser = opts.browser_original
            self:list_request(opts, todos)
            self:list_persist(opts, todos, true)
        end
        groups = self:list_reload(opts)
    end
    return groups
end

------------------------------------------------------------------------------------------

return M