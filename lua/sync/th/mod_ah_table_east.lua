-- http://nufm.dfcfw.com/EM_Finance2014NumericApplication/JS.aspx?cb=jQuery11240501735283856235_1554968911732&type=CT&token=4f1862fc3b5e77c150a2b985b12db0fd&sty=FCABHL&js=(%7Bdata%3A%5B(x)%5D%2CrecordsFiltered%3A(tot)%7D)&cmd=C._AHH&st=(AB%2FAH%2FHKD)&sr=-1&p=4&ps=20&_=1554968911745

--[[
序号	H股代码	H股名称	最新价(HKD)	涨跌幅	港股吧	A股代码	A股名称	最新价(RMB)	涨跌幅	A股吧	比价(A/H)	溢价(A/H)
jQuery112409831336260877317_1554968702483({data:["5,00811,新华文轩,6.260,-0.48,601811,1,新华文轩,13.80,-2.82,42.048,5.360,0.33,2.57,0.6718,-1.5747,-0.030,-0.40","5,00317,中船防务,9.390,-4.09,600685,1,中船防务,19.80,-2.51,63.073,8.040,0.31,2.46,0.6861,-1.4628,-0.400,-0.51","5,06869,长飞光纤光缆,21.450,-2.50,601869,1,长飞光纤,45.07,-2.87,144.080,18.365,0.31,2.45,0.6872,-1.4541,-0.550,-1.33","5,01533,庄园牧场,6.740,-0.74,002910,2,庄园牧场,14.06,-3.57,45.273,5.771,0.31,2.44,0.6894,-1.4364,-0.050,-0.52","5,02727,上海电气,3.060,-1.61,601727,1,上海电气,6.39,-3.03,20.554,2.620,0.31,2.44,0.6891,-1.4390,-0.050,-0.20","5,01772,赣锋锂业,14.400,-0.14,002460,2,赣锋锂业,29.85,1.98,96.725,12.329,0.31,2.42,0.6914,-1.4211,-0.020,0.58","5,06198,青岛港,5.600,-1.41,601298,1,青岛港,10.98,2.33,37.615,4.795,0.29,2.29,0.7081,-1.2900,-0.080,0.25","5,06116,拉夏贝尔,4.610,-2.95,603157,1,拉夏贝尔,8.98,-3.02,30.965,3.947,0.29,2.28,0.7100,-1.2751,-0.140,-0.28","5,00719,山东新华制药股份,4.750,-2.86,000756,2,新华制药,9.26,-1.49,31.906,4.067,0.29,2.28,0.7098,-1.2769,-0.140,-0.14","5,03958,东方证券,6.190,-0.64,600958,1,东方证券,11.85,-2.31,41.578,5.300,0.29,2.24,0.7150,-1.2359,-0.040,-0.28","5,00107,四川成渝高速公路,2.660,-0.37,601107,1,四川成渝,4.88,-2.79,17.867,2.277,0.27,2.14,0.7269,-1.1427,-0.010,-0.14","5,01072,东方电气,7.200,1.41,600875,1,东方电气,12.93,10.04,48.362,6.165,0.27,2.10,0.7326,-1.0974,0.100,1.18","5,06196,郑州银行,3.480,-2.52,002936,2,郑州银行,5.86,-1.68,23.375,2.980,0.25,1.97,0.7493,-0.9667,-0.090,-0.10","5,01800,中国交通建设,8.110,-1.93,601800,1,中国交建,13.65,-3.33,54.475,6.944,0.25,1.97,0.7494,-0.9658,-0.160,-0.47","5,01528,红星美凯龙,7.730,-1.65,601828,1,美凯龙,12.96,-3.14,51.922,6.618,0.25,1.96,0.7504,-0.9582,-0.130,-0.42","5,00598,中国外运,3.610,-0.55,601598,1,中国外运,6.02,-3.06,24.248,3.091,0.25,1.95,0.7517,-0.9477,-0.020,-0.19","5,01053,重庆钢铁股份,1.350,-2.88,601005,1,重庆钢铁,2.25,-1.75,9.068,1.156,0.25,1.95,0.7519,-0.9466,-0.040,-0.04","5,06178,光大证券,7.920,-1.61,601788,1,光大证券,13.20,-2.51,53.199,6.781,0.25,1.95,0.7519,-0.9466,-0.130,-0.34","5,00564,郑煤机,4.020,1.77,601717,1,郑煤机,6.62,0.76,27.002,3.442,0.25,1.92,0.7548,-0.9233,0.070,0.05","5,01787,山东黄金,18.760,-0.42,600547,1,山东黄金,30.76,-2.16,126.011,16.062,0.24,1.92,0.7559,-0.9150,-0.080,-0.68"],recordsFiltered:111})


]]--


local M = {}
M.__index = M
M.TOKEN_PERSIST = "ah.table"
M.URL_AH_TABLE_PREFIX = "http://nufm.dfcfw.com/EM_Finance2014NumericApplication/JS.aspx?cb=jQuery11240501735283856235_1554968911732&type=CT&token=4f1862fc3b5e77c150a2b985b12db0fd&sty=FCABHL&js=(%7Bdata%3A%5B(x)%5D%2CrecordsFiltered%3A(tot)%7D)&cmd=C._AHH&st=(AB%2FAH%2FHKD)&sr=-1&p=1&ps=500&_="


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

    local reqopts = {}
    local timestamp = os.time() * 1000
    local url = self.URL_AH_TABLE_PREFIX .. timestamp
    local reqopt = {}
    reqopt["url"] = url
    reqopts[#reqopts + 1] = reqopt

    local err
    local browser = global[opts.browser]
    reqopts, err = browser.Get(reqopts, 0, 0, 1, 1)

    if err ~= nil then
        print("[request] fatal", err)
    end
    reqopt = reqopts[1]
    local data = self:parse_response(opts, reqopt)
    return data

end


-------------------------------------------------------------------------------------------
function M:parse_response(opts, reqopt)
    local url = reqopt["url"]
    local content = reqopt["content"]

    if content == nil then
        print("[error] request failure", url)
        print(reqopt["err"])
        return
    end

    local tag_start = '%({data:%['
    local tag_end = '%],recordsFiltered'

    local istart = string.find(content, tag_start)
    local iend = string.find(content, tag_end, istart)
    content = string.sub(content, istart + #tag_start - 2, iend - 1)
    local data = {}
    local elements = simple.split(content, '","')
    --simple.table_print_all(elements)
    local i = 1
    local step = 18
    local nelements = #elements
    while i <= nelements do
        local one = {}
        one.code_h = elements[i + 1]
        one.name_h = elements[i + 2]
        one.close_h = elements[i + 3] + 0
        one.ch_h = elements[i + 4] + 0
        one.code_a = elements[i + 5]
        one.name_a = elements[i + 7]
        one.close_a = elements[i + 8]
        one.ch_a = elements[i + 9]
        data[#data + 1] = one
        i = i + step
    end
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
            map[one.code_a] = one
            map[one.name_a] = one
            map[one.code_h] = one
            map[one.name_h] = one
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