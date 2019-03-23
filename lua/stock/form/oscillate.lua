function form_oscillate(A)

    local choice = A.choice
    local up_count = A.up_count
    local down_count = A.down_count

    local down_from = up_count + 2
    local down_to = down_count + up_count + 1
    local up_from = 2
    local up_to = up_count + 1


    local date_count = down_count + up_count + 1

    local cache_code = global.cachem.Get("stock.code");
    local cache_snapshot = global.cachem.Get("stock.snapshot");
    local cache_khistory = global.cachem.Get("stock.khistory");
    local codes = cache_code.Get(false, A.market);
    local dates = global.calendar.List(date_count, A.date_offset, 0, true)


    local up_red = A.up_red
    local up_rate = A.up_rate
    local down_rate = A.down_rate
    local k_count = #dates - 1

    if up_red == nil then
        up_red = false
    end


    local all = 0
    local hit = {}
    local anti = {}
    local missing = {}
    for i = 1, #codes do
        local code = codes[i];
        local ks = cache_khistory.ListSubVal(true, { code }, dates)


        if #ks >= k_count then
            local here = true


            for n = up_from, up_to do
                local k = ks[n]
                local ch = k["change_rate"] + 0
                if ch < up_rate then
                    here = false
                    break
                end
                if up_red then
                    local kopen = k["open"] + 0
                    local kclose = k["close"]
                    if kclose == nil then
                        kclose = k["now"]
                    end
                    kclose = kclose + 0
                    if kclose < kopen then
                        here = false
                        break
                    end
                end
            end


            if here then
                for n = down_from, down_to do
                    local k = ks[n]
                    local ch = k["change_rate"] + 0
                    if ch > down_rate then
                        here = false
                        break
                    end
                end
            end

            if here then

                all = all + 1

                local k0 = ks[1]
                local k1 = ks[2]
                local open0 = k0["open"] + 0
                local close1 = k1["close"]
                if close1 == nil then
                    close1 = k1["now"]
                end
                close1 = close1 + 0
                local ch0 = k0["change_rate"] + 0

                if ch0 > 0 or open0 > close1 then
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
            dates = dates--,
            --hits = hit,
            --antis = anti
        }
    end


    if A.mode == "anti" then
        return { codes = anti }
    end

    return { codes = hit }

end

--[[
local opt = {}
opt["mode"] = global.mode
opt["market"] = "sz.sh"
opt["up_count"] = 2
opt["down_count"] = 3
opt["up_rate"] = 0.1
opt["down_rate"] = -0.1
opt["date_offset"] = 0
local r = form_oscillate(opt)
return r
]]--