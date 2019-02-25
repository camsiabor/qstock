function multi_down(A)


    local choice = A.choice
    local prev = A.prev + 0
    local prev_plus = prev + 2
    if choice then
        prev_plus = prev + 1
    end

    local cache_code = Q.cachem.Get("stock.code");
    local cache_snapshot = Q.cachem.Get("stock.snapshot");
    local cache_khistory = Q.cachem.Get("stock.khistory");
    local codes = cache_code.Get(false, A.market);
    local dates = Q.calendar.List(prev_plus - 1, A.date_offset, 0, true)
    local rate_up = A.rate_up + 0
    local rate_down = A.rate_down + 0

    local from = 3
    if choice then
        from = 2
    end



    local all = 0
    local hit = {}
    local anti = {}
    local missing = {}
    for i = 1, #codes do
        local code = codes[i];
        Qrace(code)
        local ks = cache_khistory.ListSubVal(true, { code }, dates)

        if #ks >= prev_plus then

            local k1 = ks[2]
            local open1 = k1["open"] + 0
            local close1 = k1["close"] + 0

            local here = true
            if not choice then
                if rate_up > 0 then
                    local ch_rate = (close1 - open1) / open1 * 100
                    here = ch_rate > rate_up
                else
                    here = close1 > open1
                end
            end

            if here then
                for n = from, prev_plus do
                    local k = ks[n]
                    local kopen = k["open"] + 0
                    local kclose = k["close"] + 0
                    if kclose >= kopen then
                        here = false
                        break
                    end

                    if rate_down ~= 0 then
                        local krate = (kclose - kopen) / kopen * 100
                        if krate > rate_down then
                            here = false
                            break
                        end
                    end
                end
            end

            if here then
                all = all + 1

                local k0 = ks[1]
                local open0 = k0["open"] + 0
                local close0 = k0["close"]
                if close0 == nil then
                    close0 = k0["now"]
                end
                close0 = close0 + 0

                local dohit
                if choice then
                    if rate_up > 0 then
                        local ch_rate = (close0 - open0) / open0 * 100
                        dohit = ch_rate > rate_up
                    else
                        dohit = close0 > open0
                    end
                else
                    dohit = open0 > close1 or close0 > close1
                end

                if dohit then
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
            hits = hit,
            antis = anti
        }
    end


    if A.mode == "anti" then
        return { codes = anti }
    end

    return { codes = hit }

end