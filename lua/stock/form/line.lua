function conserve(A)

    local conserve_count = A.conserve_count
    local conserve_interval = A.conserve_interval
    local conserve_expect_ch = A.conserve_expect_ch

    local ch_lower = A.ch_lower
    local ch_upper = A.ch_upper
    local ch_overall_rate = A.ch_overall_rate
    local ch_overall_lower = ch_lower * ch_overall_rate
    local ch_overall_upper = ch_upper * ch_overall_rate

    local turnover_lower = A.turnover_lower
    local turnover_upper = A.turnover_upper
    local turnover_ch_lower = A.turnover_ch_lower
    local turnover_ch_upper = A.turnover_ch_upper

    local cache_code = Q.cachem.Get("stock.code");
    local cache_snapshot = Q.cachem.Get("stock.snapshot");
    local cache_khistory = Q.cachem.Get("stock.khistory");
    local codes = cache_code.Get(false, A.market);

    local date_offset = A.date_offset


    local dates_count = conserve_count + conserve_interval + 1
    local dates = Q.calendar.List(conserve_count + conserve_interval, date_offset, 0, true)

    local to = conserve_interval + 1
    local from = to + conserve_count - 1

    local k_count
    if conserve_interval <= 0 then
        k_count = conserve_count
    else
        k_count = conserve_count + conserve_interval + 1
    end


    local all = 0
    local hit = {}
    local anti = {}

    for i = 1, #codes do
        local code = codes[i];
        local ks = cache_khistory.ListSubVal(true, { code }, dates)

        if #ks >= k_count then

            local here = true

            local k_from = ks[from]
            local k_to = ks[to]

            local k_from_open = k_from["open"] + 0
            local k_to_close = k_to["close"]
            if k_to_close == nil then
                k_to_close = k_to["now"]
            end
            k_to_close = k_to_close + 0

            local k_ch_overall = ( k_to_close - k_from_open ) / k_from_open * 100
            local here = k_ch_overall >= ch_overall_lower and k_ch_overall <= ch_overall_upper


            if here then

                for n = from, to, -1 do
                    local k = ks[n]
                    local ch = k["change_rate"] + 0
                    local turnover = k["turnover"] + 0

                    here = ch >= ch_lower and ch <= ch_upper
                    if here then
                        here = turnover >= turnover_lower and turnover <= turnover_upper
                    end

                    if here and n > 1 and  ch_ch_lower ~= 0 then
                        local knext = ks[n - 1]
                        local turnovern = knext["turnover"] + 0
                        local turnover_ch = (turnovern - turnover) / turnover * 100
                        here = turnover_ch >= turnover_ch_lower and turnover_ch <= turnover_ch_upper
                    end
                    if not here then
                        break
                    end
                end
            end

            if here then
                all = all + 1
                if conserve_interval == 0 then
                    hit[#hit + 1] = code
                else
                    here = false
                    local conserve_close = ks[to]["close"] + 0
                    for n = 1, conserve_interval do
                        local k = ks[n]
                        local kclose = k["close"] + 0
                        local expect_ch = (kclose - conserve_close) / kclose * 100
                        here =  expect_ch >= conserve_expect_ch
                        if here then
                            break
                        end
                    end

                    if here then
                        hit[#hit + 1] = code
                    else
                        anti[#anti + 1] = code
                    end
                end
            end
        end
    end

    if A.mode == "debug" then
        local hitp
        if all == 0 then
            hitp = 0
        else
            hitp = #hit / all * 100
        end
        return {
            all = all,
            hit = #hit,
            anti = #anti,
            hitp = hitp,
            dates = dates
        }
    end


    if A.mode == "anti" then
        return { codes = anti }
    end

    return { codes = hit }

end

local opt = {}
opt["mode"] = Q.mode
opt["market"] = "sz.sh"
opt["date_count"] = 0
opt["date_offset"] = -1
opt["conserve_count"] = 5
opt["conserve_interval"] = 0
opt["conserve_expect_ch"] = 2
opt["ch_lower"] = -2.5
opt["ch_upper"] = 2.5
opt["ch_overall_rate"] = 1.5
opt["turnover_lower"] = 0.5
opt["turnover_upper"] = 5
opt["turnover_ch_lower"] = -50
opt["turnover_ch_upper"] = 50
local r = conserve(opt)
return r