-- http://stockpage.10jqka.com.cn/000803/funds/


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

local url = "http://data.10jqka.com.cn/funds/ggzjl/"
url = "http://stockpage.10jqka.com.cn/603178/funds"
url = "http://data.10jqka.com.cn/funds/ggzjl/"

local h, err = Q.http.Get(url, nil, "gbk")
if h == nil then
    print(h)
    return
else
    print(h)
    return
end

-------------------------------------------------------------------------------------------
function M:request(opts, data, result)

    local url_prefix = "http://stockpage.10jqka.com.cn/"
    local url_suffix = "/funds"

    local count = 1
    local reqopts = {}
    
    
    local codes = opts.codes
    local count = #codes
    
    for i = 1, count do
        local code = codes[i]
        local url = url_prefix..code..url_suffix
        local reqopt = {}
        reqopt["code"] = code
        reqopt["url"] = url
        reqopts[i] = reqopt
    end


    local err
    local browser = Q[opts.browser]
    reqopts, err = browser.Get(reqopts, opts.nice, opts.newsession, opts.concurrent)

    if err ~= nil then
        print("[request] fatal", err)
    end

    reqopt = reqopts[1]
    
    
    for i = 1, count do
        local reqopt = reqopts[1]
        M:parse_html(opts, data, result, reqopt)
    end
    

    return result

end


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
    
    if html ~= nil then
        html = string.gsub(html, "<!DOCTYPE html>", "")
        --print(html)
        --return
    end

    if xml == nil then
        xml = require("common.xml2lua.xml2lua")
        xml_tree_handler = require("common.xml2lua.tree")
    end

    local tree = xml_tree_handler:new()
    local parser = xml.parser(tree)
    parser:parse(html)


    -- /html/body/div[11]/div[6]/table
    local div = tree.root.html.body.div[11].div[6]
    local htable = div.table
    if htable == nil then
        print("[error] response content invalid "..#html)
        print(url)
        print(html)
        print("")
        return
    end

    local tbody = htable.tbody

    local tr_count = #tbody.tr


    --[[
        1. 日期	
        2. 收盘价	
        3. 涨跌幅	
        4. 资金净流入	
        5. 5日主力净额	
        6. 大单净额
        7. 大单净占比
        8. 中单净额
        9. 中单净占比
        10. 小单净额
        11 小单净占比
    ]]--

    local code = reqopt["code"]
    for i = 3, tr_count do

        local tr = tbody.tr[i]
        local v = tr.td[1][1]
        
        
        local tdate = tr.td[1][1]
        local kclose = tr.td[2][1]
        local change_rate = tr.td[3][1]
        local flow = tr.td[4][1]
        local big5 = tr.td[5][1]
        local big = tr.td[6][1]
        local big_r = tr.td[7][1]
        local mid = tr.td[8][1]
        local mid_r = tr.td[9][1]
        local tin = tr.td[10][1]
        local tin_r = tr.td[11][1]
        
        kclose = kclose + 0
        
        flow = simple.numcon((flow + 0) / 10000)
        big5 = simple.numcon((big5 + 0) / 10000)
        bin = simple.numcon((big + 0) / 10000)
        mid = simple.numcon((mid + 0) / 10000)
        tin = simple.numcon((tin + 0) / 10000)
        
        change_rate = simple.percent2num(change_rate)
        big_r = simple.percent2num(big_r)
        mid_r = simple.percent2num(mid_r)
        tin_r = simple.percent2num(tin_r)
        
        print(tdate.."\t"..kclose.."\t"..change_rate)
        
        local one = {}
        
        
        data[#data + 1] = one


    end -- for tr end
end



local data = {}
local result = {}

local opts = {}
opts.browser = "chrome"
opts.codes = { "603178" }

opts.concurrent = 1
opts.newsession = false

M:request(opts, data, result)

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
    print("[persist] data count", n)
    for i = 1, n do
        pageone[#pageone + 1] = data[i]
        if (i % 50 == 0) or (i == n) then
            local jsonstr = json.encode(pageone)
            --print(jsonstr)
            local key = self:keygen(opts, page)
            _, err = dao.Update(db, datestr, key, jsonstr, true, 0, nil)
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

    local dates = Q.calendar.List(0, opts.date_offset, 0, true)
    local datestr = dates[1]
    local db = opts.db
    local dao = Q.daom.Get("main")

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
function M:print_data(data)
    local n = #data
    local count = 1
    local print_head = "i\tcode\tname\tch\tturn\tio\tin\tbig_in\tbig_r\tbig_t\tbig_c\tcross\tcross2\tbig"
    for i = 1, n do
        local one = data[i]
        if count % 10 == 1 then
            print("")
            print(print_head)
        end
        print(one.index.."\t"..one.code.."\t"..one.name.."\t"..one.change_rate.."\t"..one.turnover.."\t"..one.flow_io_rate.."\t"..one.flow_in_rate.."\t"..one.flow_big_in_rate.."\t"..one.flow_big_rate.."\t"..one.flow_big_rate_total.."\t"..one.flow_big_rate_compare.."\t"..one.flow_big_rate_cross.."\t"..one.flow_big_rate_cross_ex.."\t"..one.flow_big)
        count = count + 1
    end
end


------------------------------------------------------------------------------------------

function M:go(opts)
    local data = {}
    local result = {}
    if opts.dofetch then
        self:request(opts, data, result)
        self:persist(opts, data)
    else
        self:reload(opts, data, result)
    end

    if opts.filter == nil then
        self:filter(opts, data, result)
    else
        opts.filter(opts, data, result)
    end

    simple.table_sort(result, opts.sort_field)

    self:print_data(result)
    return data, result
end

--return M