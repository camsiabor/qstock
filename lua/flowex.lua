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

-------------------------------------------------------------------------------------------
function M:request(opts, data, result)

    local url_prefix = "http://stockpage.10jqka.com.cn/"
    local url_suffix = "/funds"

    local count = 1
    local reqopts = {}

    local url = url_prefix..opts.code..url_suffix
    local reqopt = {}
    reqopt["url"] = url
    reqopts[1] = reqopt


    local err
    local browser = Q[opts.browser]
    if browser == nil then
        browser = global.firefox
    end
    if opts.concurrent <= 1 then
        if opts.newsession then
            reqopts, err = browser.GetEx(reqopts, opts.nice)
        else
            reqopts, err = browser.Get(reqopts, opts.nice)
        end
    else
        reqopts, err = browser.GetConcurrent(reqopts, opts.nice, opts.concurrent, opts.newsession)
    end

    if err ~= nil then
        print("[request] fatal", err)
    end

    M:parse_html(opts, data, result, reqopt)

    return result
end


-------------------------------------------------------------------------------------------
function M:parse_html(opts, data, result, reqopt)
    local url = reqopt["url"]
    local html = reqopt["content"]

    print(html)

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


    local htable = tree.root.html.body.table[2]
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

        print(tr[1])

        -- local index = tr.td[1][1]
        -- local code = tr.td[2].a[1]
        -- local name = tr.td[3].a[1]
        -- local change_rate = tr.td[5][1]
        -- local turnover = tr.td[6][1]
        -- local flow_in = tr.td[7][1]
        -- local flow_out = tr.td[8][1]
        -- local flow = tr.td[9][1]
        -- local amount = tr.td[10][1]
        -- local flow_big = tr.td[11][1]

        -- turnover = string.gsub(turnover, "%%", "") + 0
        -- change_rate = string.gsub(change_rate, "%%", "") + 0

        -- flow = simple.str2num(flow)
        -- flow_in = simple.str2num(flow_in)
        -- flow_out = simple.str2num(flow_out)
        -- flow_big = simple.str2num(flow_big)

        -- flow_in = simple.nozero(flow_in)
        -- flow_out = simple.nozero(flow_out)
        -- flow_big = simple.nozero(flow_big)

        -- amount = simple.str2num(amount)

        -- local flow_big_rate = flow_big / amount * 100
        -- local flow_big_rate_compare = flow / flow_big
        -- local flow_big_rate_total = turnover * flow_big_rate / 100

        -- local flow_in_rate = flow_in / amount * 100
        -- local flow_out_rate = flow_out / amount * 100
        -- local flow_io_rate = flow_in / flow_out

        -- local flow_big_in_rate = flow_big / flow_in * 100

        -- --local flow_big_rate_cross = (turnover * amount * flow_big_rate / 100) * flow_io_rate * flow_big_in_rate
        -- local flow_big_rate_cross = flow_io_rate * flow_big_rate_total * flow_big_rate / 100 * flow_big_in_rate
        -- local change_rate_ex = change_rate
        -- if change_rate_ex < 0 then
        --     change_rate_ex = 0.1
        -- end
        -- local flow_big_rate_cross_ex = flow_big_rate_cross / (change_rate_ex + 2.5)

        -- flow_big_rate = simple.numcon(flow_big_rate)
        -- flow_big_rate_compare = simple.numcon(flow_big_rate_compare)
        -- flow_big_rate_total = simple.numcon(flow_big_rate_total)
        -- flow_big_rate_cross = simple.numcon(flow_big_rate_cross)
        -- flow_big_rate_cross_ex = simple.numcon(flow_big_rate_cross_ex)

        -- flow_in_rate = simple.numcon(flow_in_rate)
        -- flow_out_rate = simple.numcon(flow_out_rate)
        -- flow_io_rate = simple.numcon(flow_io_rate)

        -- flow_big_in_rate = simple.numcon(flow_big_in_rate)

        -- local one = {}
        -- one.index = index
        -- one.code = code
        -- one.name = name
        -- one.flow = flow
        -- one.flow_in = flow_in
        -- one.flow_out = flow_out
        -- one.amount = amount
        -- one.turnover = turnover
        -- one.flow_big = flow_big
        -- one.change_rate = change_rate

        -- one.flow_big_rate = flow_big_rate
        -- one.flow_big_rate_total = flow_big_rate_total
        -- one.flow_big_rate_compare = flow_big_rate_compare
        -- one.flow_big_rate_cross = flow_big_rate_cross
        -- one.flow_big_rate_cross_ex = flow_big_rate_cross_ex

        -- one.flow_in_rate = flow_in_rate
        -- one.flow_out_rate = flow_out_rate
        -- one.flow_io_rate = flow_io_rate

        -- one.flow_big_in_rate = flow_big_in_rate

        -- data[#data + 1] = one


    end -- for tr end
end


local data = {}
local result = {}

local opts = {}

opts.browser = "chrome"
opts.concurrent = 2
opts.newsession = false
opts.code = "000001"
M:request(opts, data, result)


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

    local dates = global.calendar.List(0, opts.date_offset, 0, true)
    local datestr = dates[1]
    local db = opts.db
    local dao = global.daom.Get("main")

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