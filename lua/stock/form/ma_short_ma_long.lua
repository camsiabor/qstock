local ma_short_count = 5
local ma_long_count = 20

local ma_rate_upper = 1
local ma_rate_lower = -1

local cache_code = global.cachem.Get("stock.code");
local cache_snapshot = global.cachem.Get("stock.snapshot");
local cache_khistory = global.cachem.Get("stock.khistory");
local codes = cache_code.Get(false, "sz.sh");

local dates = global.calendar.List(ma_long_count - 1, 0, 0, true)

local all = 0
local hit = {}
local anti = {}
local kcount = #dates
for i = 1, #codes do
    local code = codes[i];
    --local snapshot = cache_snapshot.Get(true, code);
    local ks = cache_khistory.ListSubVal(true, { code }, dates)

    if #ks >= kcount then

        local sum = 0
        local ma_long = 0
        local ma_short = 0
        for n = 1, ma_long_count do
            local kclose = ks[n]["close"]
            if kclose == nil then
                kclose = ks[n]["now"]
            end
            sum = sum + kclose
            if n <= ma_short_count then
                ma_short = sum / ma_short_count
            end
        end
        ma_long = sum / ma_long_count
        local ma_rate = (ma_short - ma_long) / ma_long * 100
        if ma_rate >= ma_rate_lower and ma_rate <= ma_rate_upper then
            hit[#hit + 1] = code
        end

    end
end

return { codes = hit }