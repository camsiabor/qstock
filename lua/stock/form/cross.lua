--[[

    market
        market

    date_offset
        offset of day

    observe_count
        count of observe

    cross_count
        count of cross

    c_change_rate
        current change rate

    v_swing_rate_lower
    v_swing_rate_higher
        vertical swing rate : (high0 - low0) / close0

    h_swing_rate_lower
    h_swing_rate_higher
        horizontal swing rate : (close0 - close1) / close1

    turnover_lower
        turnover lower

]]--

function form_cross(A)

    local market = A.market
    local date_offset = A.date_offset
    local c_change_rate = A.c_change_rate
    local observe_count = A.observe_count

    local cross_count = A.cross_count
    local turnover_lower = A.turnover_lower
    local v_swing_rate_lower = A.v_swing_rate_lower
    local v_swing_rate_higher = A.v_swing_rate_higher
    local h_swing_rate_lower = A.h_swing_rate_lower
    local h_swing_rate_higher = A.h_swing_rate_higher


    local cache_code = global.cachem.Get("stock.code");

    local cache_snapshot = global.cachem.Get("stock.snapshot");
    local cache_khistory = global.cachem.Get("stock.khistory");
    local codes = cache_code.Get(false, A.market);
    local dates = global.calendar.List(cross_count, date_offset, 0, true)
    --Qr(dates)
    local date_count = cross_count + 1
    local cross_to = 1
    local cross_from = cross_count

    local all = 0
    local hit = {}
    local anti = {}
    for i = 1, #codes do

        local code = codes[i];
        local ks = cache_khistory.ListSubVal(true, { code }, dates)
        if #ks >= date_count then

            local here = false

            for n = cross_from, cross_to, -1 do

                Qrace({ cross_from, cross_to, n, #ks})

                local k0 = ks[n]
                local k1 = ks[n + 1]

                -- turnover --
                local turnover0 = k0["turnover"] + 0
                here = turnover0 >= turnover_lower
                if not here then
                    break
                end


                -- horizontal swing rate --
                local close0 = k0["close"]
                if close0 == nil then
                    close0 = k0["now"]
                end
                close0 = close0 + 0
                local close1 = k1["close"] + 0
                local h_change = (close0 - close1) / close1 * 100
                here = h_change >= h_swing_rate_lower and h_change <= h_swing_rate_higher
                if not here then
                    break
                end


                -- vertical swing rate --
                local low0 = k0["low"] + 0
                local high0 = k0["high"] + 0
                local v_change = (high0 - low0) / close1 * 100
                if v_change < 0 then
                    v_change = -v_change
                end
                here = v_change >= v_swing_rate_lower and v_change <= v_swing_rate_higher
                if not here then
                    break
                end

            end


            if here then

                if observe_count <= 0 then
                    all = all + 1
                    hit[#hit + 1] = code
                else
                    -- TODO
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
opt["mode"] = global.mode
opt["market"] = "sz.sh"
opt["date_offset"] = -1
opt["cross_count"] = 2
opt["observe_count"] = 0
opt["turnover_lower"] = 0.3
opt["v_swing_rate_lower"] = 2
opt["v_swing_rate_higher"] = 10
opt["h_swing_rate_lower"] = -1
opt["h_swing_rate_higher"] = 1

local r = form_cross(opt)
return r