-- http://stockpage.10jqka.com.cn/000803/funds/


local M = {}
M.__index = M
M.persist_key = "flow"

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

------------------------------------------------------------------------------------------
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

    for i = 1, count do
        local reqopt = reqopts[i]
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
    
    
    -- html = string.gsub(html, "<!DOCTYPE html>", "")

    local tag_start = '<table class="m_table_3">'
    local tag_end = '</table>'
    local index = string.find(html, tag_start)   
    html = string.sub(html, index, #html)
    index = string.find(html, tag_end)
    html = string.sub(html, 1, index + #tag_end)


    if xml == nil then
        xml = require("common.xml2lua.xml2lua")
        xml_tree_handler = require("common.xml2lua.tree")
    end
   
    local tree = xml_tree_handler:new()
    local parser = xml.parser(tree)
    parser:parse(html)
    
  
    -- /html/body/div[11]/div[6]/table
    -- local div = tree.root.html.body.div[11].div[6]
    -- local htable = div.table
    local htable = tree.root.table
    if htable == nil then
        print("[error] response content invalid "..#html)
        print(url)
        print(html)
        print("")
        return
    end

    --local tbody = htable.tbody
    local tr_count = #htable.tr

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

    local one = {}
    one.flow = {}
    one.code = reqopt["code"]
    data[#data + 1] = one
    
    for i = 3, tr_count do

        local tr = htable.tr[i]
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
        big = simple.numcon((big + 0) / 10000)
        mid = simple.numcon((mid + 0) / 10000)
        tin = simple.numcon((tin + 0) / 10000)
        
        change_rate = simple.percent2num(change_rate)
        big_r = simple.percent2num(big_r)
        mid_r = simple.percent2num(mid_r)
        tin_r = simple.percent2num(tin_r)
        
        local record = {}
        record["date"] = tdate
        record["close"] = kclose
        record.flow = flow
        record.big5 = big5
        record.big = big
        record.mid = mid
        record.tin = tin
        record.big_r = big_r
        record.mid_r = mid_r
        record.tin_r = tin_r
        
        record.change_rate = change_rate
        
    
        one.flows[#one.flows + 1] = record
        
    end -- for tr end
    
end



-------------------------------------------------------------------------------------------

function M:keygen(opts, page)
    local key = string.format("%s.%s.%s.%d", opts.datasrc, opts.field, opts.order, page)
    return key
end



-------------------------------------------------------------------------------------------

function M:persist(opts, data)

    local db = opts.db
    local dao = Q.daom.Get("main")


    local n = #data
    print("[persist] fund data count", n)
    for i = 1, n do
        local one = data[i]
        local group = "ch."..one.code
        pageone[#pageone + 1] = data[i]
        local jsonstr = json.encode(one)
        _, err = dao.Update(db, group, self.persist_key , jsonstr, true, 0, nil)
        if err == nil then
            if opts.debug then
                print("[persist]", group)
            end
        else
            print("[persist] failure", err)
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

    local n = #opts.codes
    for i = 1, n do
        local code = opts.codes[i]
        local group = "ch."..code
        local datastr, err = dao.Get(db, group, self.persist_key, nil)
        if err ~= nil then
            print("[reload] failure", group, self.persist_key, err)
        end
        if datastr == nil or #datastr == 0 then
            print("[reload] empty", datestr, key)
        else
            local one = json.decode(datastr)
            data[#data + 1] = one
        end

    end

end


-------------------------------------------------------------------------------------------
function M:filter(opts, data, result)
    local n = #data
    for i = 1, n do
        local one = data[i]
        result[#result + 1] = one
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




local data = {}
local result = {}

local opts = {}
opts.browser = "wget"
opts.codes = { "603178", "000001", "000009" }

opts.concurrent = 1
opts.newsession = false

opts.dofetch = true

opts.db = "flow"
opts.persist = true

M:go(opts, data, result)

--return M