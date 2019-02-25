function multi_up(A)

    local prev = A.prev + 0
    local cache_code = Q.cachem.Get("stock.code");
    local cache_snapshot = Q.cachem.Get("stock.snapshot");
    local cache_khistory = Q.cachem.Get("stock.khistory");
    local codes = cache_code.Get(false, A.market);
    local dates = Q.calendar.List(prev, A.date_offset, 0, true)

    local pb_low = A.pb_low + 0
    local pb_high = A.pb_high + 0
    local ch_low = A.ch_low + 0
    local ch_high = A.ch_high + 0


    local all = 0
    local hit = {}
    local anti = {}
    local missing = {}
    for i = 1, #codes do
        local code = codes[i];
        local ks = cache_khistory.ListSubVal(true, { code }, dates)
        local prev_plus_1 = prev + 1

        if #ks >= prev_plus_1 then

            local here = true

            for n = 2, prev_plus_1 do
                local k = ks[n]
                local ch = k["change_rate"] + 0
                if ch < ch_low then
                    here = false
                    break
                end
            end

            if here then
                all = all + 1
                local k0 = ks[1]
                local k1 = ks[2]
                local ch0 = k0["change_rate"] + 0;
                local open0 = k0["open"] + 0
                local close0 = k0["close"] + 0
                --local high0 = k0["high"] + 0
                local close1 = k1["close"] + 0
                if open0 > close1 or close0 > close1 then
                    hit[#hit + 1] = code
                else
                    anti[#anti + 1] = code
                end
            end
        else
            missing[#missing + 1] = code
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
            dates = dates,
        }
    end


    if A.mode == "anti" then
        return { codes = anti }
    end

    return { codes = hit }

end