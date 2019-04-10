-- http://vis-free.10jqka.com.cn/hk/cache/ahld_yjl_desc_b.html

local M = {}
M.__index = M
M.TOKEN_PERSIST = "ah.table"
M.URL_AH_TABLE = "http://vis-free.10jqka.com.cn/hk/cache/ahld_yjl_desc_b.html"

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

    local url_pattern = self.URL_AH_TABLE

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
    local data = self:parse_html(opts, reqopt)
    return data

end

-------------------------------------------------------------------------------------------
function M:parse_html(opts, reqopt)
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
        one.name = tds[2][1]
        one.code_a = tds[3][1]
        one.ch_a = tds[4][1] + 0
        one.close_a = tds[5][1] + 0
        one.pe_a = tds[6][1]
        one.pb_a = tds[7][1]
        one.vol_a = tds[8][1]
        one.code_h = tds[9][1]
        one.ch_h = tds[10][1] + 0
        one.close_h = tds[11][1] + 0
        one.pe_h = tds[12][1]
        one.pb_h = tds[13][1]
        one.vol_h = tds[14][1]

        if one.pe_a == "亏损" then
            one.pe_a = -1
        end

        if one.pe_h == "亏损" then
            one.pe_h = -1
        end

        one.ch_a = simple.numcon(one.ch_a)
        one.ch_h = simple.numcon(one.ch_h)

        one.close_a = simple.numcon(one.close_a)
        one.close_h = simple.numcon(one.close_h)

        one.vol_a = cal.str2num(one.vol_a)
        one.vol_h = cal.str2num(one.vol_h)


        data[#data + 1] = one
    end -- for tr end
    return data
end

-------------------------------------------------------------------------------------------

function M:persist(opts, data)

    local db = opts.db
    if db == nil then
        db = "group"
    end

    local dates = global.calendar.List(0, 0, 0, false)
    local datestr = dates[1]
    local dao = global.daom.Get("main")
    local n = #data
    local jsonstr = json.encode(data)
    local _, err = dao.Update(db, self.TOKEN_PERSIST, datestr, jsonstr, true, 0, nil)
    if err == nil then
        print("[persist] success", db, self.TOKEN_PERSIST, datestr, n)
    else
        print("[persist] failure", db, self.TOKEN_PERSIST, datestr, err)
    end
end

-------------------------------------------------------------------------------------------

function M:reload(opts, datestr)

    if opts.db == nil then
        opts.db = "group"
    end

    if datestr == nil or #datestr == 0 then
        local dates = global.calendar.List(0, 0, 0, false)
        datestr = dates[1]
    end

    local data = {}
    local db = opts.db
    local dao = global.daom.Get("main")
    local datastr, err = dao.Get(db, self.TOKEN_PERSIST, datestr, 0, nil)
    if err ~= nil then
        print("[reload] failure", db, self.TOKEN_PERSIST, datestr, err)
    end
    if datastr == nil or #datastr == 0 then
        print("[reload] empty", db, self.TOKEN_PERSIST, datestr)
    else
        data = json.decode(datastr)
    end
    print("[reload]", db, self.TOKEN_PERSIST, datestr, #data)

    if opts.reload_as_map ~= nil and opts.reload_as_map then
       local map = {}
        for i = 1, #data do
            local one = data[i]
            map[one.name] = one
            map[one.code_a] = one
            map[one.code_h] = one
        end
        data = map
    end
    return data
end

function M:reloads(opts)

    local dates = global.calendar.List(opts.date_offset_from, opts.date_offset, opts.date_offset_to, false)
    local data = {}
    for i = 1, #dates do
        local datestr = dates[i]
        local one = self:reload(opts, datestr)
        data[#data + 1] = one
    end
    return data
end


------------------------------------------------------------------------------------------

function M:go(opts)
    if opts.request then
        local data = self:request(opts)
        self:persist(opts, data)
    end
    local data = self:reload(opts, opts.reload_as_map)
    return data
end

return M