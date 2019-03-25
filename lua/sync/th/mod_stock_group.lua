-- http://q.10jqka.com.cn/gn/detail/code/301365/
-- http://q.10jqka.com.cn/gn/detail/field/264648/order/desc/page/2/ajax/1/code/304582
-- http://q.10jqka.com.cn/gn/detail/field/3475914/order/desc/page/1/ajax/1/code/301558

--
-- http://q.10jqka.com.cn/thshy/detail/code/881121/
-- http://q.10jqka.com.cn/thshy/detail/field/199112/order/desc/page/2/ajax/1/code/881121

-- 我在去年股票最低估的时候，我选出了以下几支     601258       000727   600157
local M = {}
M.__index = M



M.TOKENS = {
    MAPPING_ENG_TO_CH = {
        concept = "gn",
        industry = "thshy",
        csrc = "zjhhy",
        region = "dy",
    },
    PERSIST_CONCEPT_GROUP = "ch.stock.group.concept.index",
    PERSIST_CONCEPT_LIST = "ch.stock.group.concept",
    PERSIST_INDUSTRY_GROUP = "ch.stock.group.industry.index",
    PERSIST_INDUSTRY_LIST = "ch.stock.group.industry",
    PERSIST_CSRC_GROUP = "ch.stock.group.scrc.index",
    PERSIST_CSRC_LIST = "ch.stock.group.scrc",
    PERSIST_REGION_GROUP = "ch.stock.group.region.index",
    PERSIST_REGION_LIST = "ch.stock.group.region",
}

M.URL_PATTERNS = {
    concept_group = "http://q.10jqka.com.cn/gn/",
    concept_list = "http://q.10jqka.com.cn/gn/detail/field/199112/order/desc/page/%d/ajax/1/code/%s",
    industry_group = "http://q.10jqka.com.cn/thshy/",
    industry_list = "http://q.10jqka.com.cn/thshy/detail/field/199112/order/desc/page/%d/ajax/1/code/%s",
    csrc_group = "http://q.10jqka.com.cn/zjhhy/",
    csrc_list = "http://q.10jqka.com.cn/zjhhy/detail/field/199112/order/desc/page/%d/ajax/1/code/%s",
    region_group = "http://q.10jqka.com.cn/dy/",
    region_list = "http://q.10jqka.com.cn/dy/detail/field/199112/order/desc/page/%d/ajax/1/code/%s",
}


local global = require("q.global")
--local logger = require("q.logger")
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

function M:get_url_pattern(request_type, group_or_list)
    local key = string.format("%s_%s", request_type, group_or_list)
    key = string.lower(key)
    local url = self.URL_PATTERNS[key]
    if url == nil then
        error("[url] pattern not found", key)
    end
    return url
end

-------------------------------------------------------------------------------------------

function M:get_token(request_type, group_or_list)
    local key = string.format("persist_%s_%s", request_type, group_or_list)
    key = string.upper(key)
    local token = self.TOKENS[key]
    if token == nil then
        print("[token] not found", key)
    end
    return token
end

-------------------------------------------------------------------------------------------

function M:group_request(opts)
    local url = self:get_url_pattern(opts.request_type, "group")
    if url == nil then
        return
    end
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
    if html == nil then
        error("[parse] group no html")
        return
    end
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
    local request_type_ch = self.TOKENS.MAPPING_ENG_TO_CH[opts.request_type]
    local pattern = '<a href="http://q.10jqka.com.cn/' .. request_type_ch .. '/detail/code/(%d+)/" target="_blank">(%W+)</a>'
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
    local token = self:get_token(opts.request_type, "group")
    local _, err = dao.Update(db, token, "", jsonstr, true, 0, nil)
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
    local token = self:get_token(opts.request_type, "group")
    local datastr, err = dao.Get(db, token, "", 0, nil)
    if err ~= nil then
        print("[reload] failure", token, "", err)
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
            local token = self:get_token(opts.request_type, "list")
            local _, err = dao.Update(db, token, key, jsonstr, true, 0, nil)
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
    local token = self:get_token(opts.request_type, "list")
    local map, err = dao.Get(db, "", token, 1, nil)
    if err ~= nil then
        print("[reload] failure", db, token, err)
        return
    end
    if map == nil or simple.table_count(map) == 0 then
        print("[reload] empty", db, token)
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

function M:code_group_mapping(request_types, from_cache)
    local all = {}
    for i = 1, #request_types do
        local request_type = request_types[i]
        --print("[rq]", request_type)
        local mapping
        local mappingstr
        local token = self:get_token(request_type, "list")
        --print("[token]", token)
        local cache = global.cachem.Get("stock.group")
        if from_cache then
            local mappingstr = cache.Get(false, token)
            if mappingstr ~= nil and #mappingstr > 0 then
                mapping = json.decode(mappingstr)
            end
        end
        if mapping == nil then
            local groups = self:list_reload({ request_type = request_type })
            if groups ~= nil then
                mapping = self:list_code_group_mapping(groups)
                mappingstr = json.encode(mapping)
                cache.Set(mappingstr, token)
            end
        end

        if mapping ~= nil then
            all[#all + 1] = mapping
        end
    end
    all = simple.table_merge(all)
    return all
end

function M:go(opts)

    local request_type = opts.request_type
    if request_type == nil or request_type == "" then
        print("[request] invalid no request type")
        return
    end

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


function M:goes(opts)
    local all = {}
    local request_types = opts.request_types
    for i = 1, #request_types do
        opts.request_type = request_types[i]
        local groups = M:go(opts)
        all[#all + 1] = groups
    end
    all = simple.table_merge(all)
    return all
end

------------------------------------------------------------------------------------------

return M