

local simple = require("common.simple")

local M = {}

-----------------------------------------------------------------------------------------------------------
function M.io(fopts)

    simple.def(fopts, "io_lower", 1)
    simple.def(fopts, "io_upper", 100)

    simple.def(fopts, "big_in_lower", 10)
    simple.def(fopts, "big_in_upper", 100)

    simple.def(fopts, "turnover", 0.5)

    simple.def(fopts, "ch_lower", -1.5)
    simple.def(fopts, "ch_upper", 6)


    simple.def(fopts, "date_offset", 0)


    local date_offset = fopts.date_offset
    local msg = "[filter] [io] %f <= io <= %f, %f <= ch <= %f, date_offset = %d"
    print(string.format(msg, fopts.io_lower, fopts.io_upper, fopts.ch_lower, fopts.ch_upper, fopts.date_offset))
    return function(one, series, code, currindex, opts)

        --[[
        if code == "600340" then
            for i = 1, #series do
                local o = series[i]
                print("?", i, o.change_rate)
            end
        end
        ]]--
        --print(one.name, "-----------------------------", #series, "curindex", currindex)


        if date_offset ~= 0 then
            if series == nil then
                return false
            end
            one = series[currindex + date_offset]
        end
        if one == nil or one.empty then
            return false
        end

        return one.flow_io_rate >= fopts.io_lower and one.flow_io_rate <= fopts.io_upper
                and one.flow_big_in_rate >= fopts.big_in_lower and one.flow_big_in_rate <= fopts.big_in_upper
                and one.turnover >= fopts.turnover
                and one.change_rate >= fopts.ch_lower and one.change_rate <= fopts.ch_upper

    end
end

-----------------------------------------------------------------------------------------------------------
function M.io_increase(fopts)


    simple.def(fopts, "in_lower", 50)
    simple.def(fopts, "in_upper", 100)
    simple.def(fopts, "in_swing", 3)

    simple.def(fopts, "ch_lower", -1.5)
    simple.def(fopts, "ch_upper", 6)

    return function(one, series, code, currindex, opts)
        if currindex == 0 then
            return true
        end
        local include = true

        for i = 1, currindex - 1 do
            local prev = series[i]
            local curr = series[i + 1]

            if curr == nil or curr.empty then
                return false
            end
            if prev == nil or prev.empty then
                return false
            end

            local ch = curr.change_rate
            local flow_in = curr.flow_in_rate

            if ch < fopts.ch_lower or ch > fopts.ch_upper then
                include = false
                break
            end

            if flow_in < fopts.in_lower or flow_in > fopts.in_upper then
                include = false
                break
            end

            if curr.flow_in_rate < (prev.flow_in_rate - fopts.in_swing) then
                include = false
                break
            end

        end

        return include
    end
end

-----------------------------------------------------------------------------------------------------------
---
---
---
function M.io_any(fopts)

    simple.def(fopts, "io_lower", 1)
    simple.def(fopts, "io_upper", 100)

    simple.def(fopts, "big_in_lower", 10)
    simple.def(fopts, "big_in_upper", 100)

    simple.def(fopts, "turnover", 0.5)

    simple.def(fopts, "ch_lower", -1.5)
    simple.def(fopts, "ch_upper", 6)

    simple.def(fopts, "date_offset", 0)

    simple.def(fopts, "ch_avg_lower", 0)
    simple.def(fopts, "ch_avg_upper", 10)

    local cal_avg = fopts.ch_avg_lower < fopts.ch_avg_upper

    return function(one, series, code, currindex, opts)
        if series == nil then
            series = { one }
        end
        local include = false
        for i = 1, currindex do
            local one = series[i]
            if one ~= nil then
                local io = one.flow_io_rate
                local big_in = one.flow_big_in_rate
                include = io >= fopts.io_lower and io <= fopts.io_upper
                        and big_in >= fopts.big_in_lower and big_in <= fopts.big_in_upper
                        and one.turnover >= fopts.turnover
                        and one.change_rate >= fopts.ch_lower and one.change_rate <= fopts.ch_upper
                if include then
                    break
                end
            end
        end

        if include and cal_avg then
            local sum = 0
            local count = 0
            for i = 1, currindex do
                local one = series[i]
                if one ~= nil then
                    count = count + 1
                    sum = sum +  one.change_rate
                end
            end
            if count == 0 then
                include = false
            else
                local avg = sum / count
                include = avg >= fopts.ch_avg_lower and avg <= fopts.ch_avg_upper
            end
        end

        return include
    end
end

-----------------------------------------------------------------------------------------------------------
function M.io_any_simple(fopts)

    simple.def(fopts, "io_lower", 1)
    simple.def(fopts, "io_upper", 100)

    simple.def(fopts, "date_offset_to", 0)
    simple.def(fopts, "date_offset_from", -100)

    simple.def(fopts, "count", 1)
    simple.def(fopts, "tag", false)

    local tag = fopts.tag
    local count = fopts.count
    local io_lower = fopts.io_lower
    local io_upper = fopts.io_upper
    local date_offset_to = fopts.date_offset_to
    local date_offset_from = fopts.date_offset_from
    return function(one, series, code, currindex, opts)
        if series == nil then
            return true
        end
        local limit = #series
        local to = currindex + date_offset_to
        local from = currindex + date_offset_from
        if to > limit then
            to = limit
        end
        if from < 1 then
            from = 1
        end
        if series == nil then
            return false
        end
        local num = 0
        for i = from, to do
            local one = series[i]
            if one ~= nil and not one.empty then
                local io = one.flow_io_rate
                if  io >= io_lower and io <= io_upper then
                    one.star = "*"
                    num = num + 1
                end
            end
        end
        if tag then
            return true
        end
        return num >= count
    end
end


-----------------------------------------------------------------------------------------------------------
function M.io_all(fopts)

    simple.def(fopts, "io_lower", 1)
    simple.def(fopts, "io_upper", 100)

    simple.def(fopts, "big_in_lower", 10)
    simple.def(fopts, "big_in_upper", 100)

    simple.def(fopts, "turnover", 0.5)

    simple.def(fopts, "ch_lower", -1.5)
    simple.def(fopts, "ch_upper", 6)

    simple.def(fopts, "date_offset", 0)

    simple.def(fopts, "ch_avg_lower", 0)
    simple.def(fopts, "ch_avg_upper", 0)
    
    simple.def(fopts, "date_offset", 0)

    local cal_avg = fopts.ch_avg_lower < fopts.ch_avg_upper

    return function(one, series, code, currindex, opts)
        local include = true
        for i = 1, currindex do
            local one = series[i]
            if one ~= nil then
                local io = one.flow_io_rate
                local big_in = one.flow_big_in_rate
                include = io >= fopts.io_lower and io <= fopts.io_upper
                        and big_in >= fopts.big_in_lower and big_in <= fopts.big_in_upper
                        and one.turnover >= fopts.turnover
                        and one.change_rate >= fopts.ch_lower and one.change_rate <= fopts.ch_upper
                if not include then
                    break
                end
            end
        end

        if include and cal_avg then
            local sum = 0
            local count = 0
            for i = 1, currindex do
                local one = series[i]
                if one ~= nil then
                    count = count + 1
                    sum = sum +  one.change_rate
                end
            end
            if count == 0 then
                include = false
            else
                local avg = sum / count
                include = avg >= fopts.ch_avg_lower and avg <= fopts.ch_avg_upper
            end
        end

        return include
    end
end

-----------------------------------------------------------------------------------------------------------
function M.codes(fopts)

    local codes = fopts.codes
    local codes_map = simple.array_to_map(codes)

    return function(one, series, code, currindex, opts)
        return codes_map[code] ~= nil
    end
end
-----------------------------------------------------------------------------------------------------------
function M.names(fopts)

    local names = fopts.names
    local names_map = simple.array_to_map(names)

    return function(one, series, code, currindex, opts)
        local name = one.name
        return names_map[name] ~= nil
    end
end


-----------------------------------------------------------------------------------------------------------
function M.names_contain(fopts)
    local names = fopts.names
    local n = #names
    return function(one, series, code, currindex, opts)
        local name = one.name
        for i = 1, n do
            if string.find(name, names[n]) ~= nil then
                return true
            end
        end
        return false
    end
end


-----------------------------------------------------------------------------------------------------------
function M.groups(fopts)
    local groups = fopts.groups
    local ngroups = #groups
    return function(one, series, code, currindex, opts)
        local one_groups = one.group
        if one_groups == nil then
            return false
        end
        for i = 1, ngroups do
            local group = groups[i]
            if one_groups[group] ~= nil then
                return true
            end
        end
        return false
    end
end
-----------------------------------------------------------------------------------------------------------
function M.no3(fopts)
    return function(one, series, code, currindex, opts)
        if code:sub(1, 1) == "3" then
            return false
        end
        return true
    end
end
-----------------------------------------------------------------------------------------------------------

function M.st(fopts)
    return function(one, series, code, currindex, opts)
        local index = string.find(one.name, "ST")
        return index ~= nil
    end
end



-----------------------------------------------------------------------------------------------------------

function M.avg_diff(fopts)
    simple.def(fopts, "field", "change_rate")
    simple.def(fopts, "short_cycle", 5)
    simple.def(fopts, "long_cycle", 10)
    simple.def(fopts, "diff_lower", 0)
    simple.def(fopts, "diff_upper", 100)
    simple.def(fopts, "deduce", "")
    simple.def(fopts, "set", "custom")

    simple.def(fopts, "per", 0)

    local per = fopts.per

    local deduce = fopts.deduce
    local dodeduce = deduce ~= nil and #deduce > 0
    return function(one, series, code, currindex, opts)
        if series == nil then
            return true
        end
        local from = currindex - fopts.long_cycle + 1
        local to = currindex
        if currindex < 0 then
            currindex = 1
        end
        local set = fopts.set
        local field = fopts.field
        local short_sum = 0
        local short_count = 0
        local short_cycle = fopts.short_cycle
        local long_sum = 0
        local long_count = 0

        local v_prev = 0
        local deduce_init = false
        local deduce_value = 100
        for i = to, from, -1 do
            local serie = series[i]
            if serie ~= nil and not serie.empty then

                local v = serie[field]
                if dodeduce then
                    if deduce_init then
                        deduce_value = deduce_value / (100 + v_prev) * 100
                    else
                        deduce_init = true
                    end
                    v_prev = v
                    v = deduce_value
                    serie[deduce] = deduce_value
                end

                if short_count < short_cycle then
                    short_sum = short_sum + v
                    short_count = short_count + 1
                end
                long_sum = long_sum + v
                long_count = long_count + 1
            end
        end
        if short_count == 0 then
            return true
        end
        --print(one.name, one.code, short_sum, long_sum, short_count, long_count)
        local short_avg = short_sum / short_count
        local long_avg = long_sum / long_count

        local rate
        if per <= 0 then
            rate = (short_avg - long_avg) / long_avg * 100
        else
            rate = short_avg / long_avg * per
        end


        --print(one.name, short_avg, long_avg, short_count, long_count, from, to)
        --print(one.name, one.code, short_avg, long_avg)
        local include = rate >= fopts.diff_lower and rate <= fopts.diff_upper
        if include and set ~= nil then
            rate = simple.numcon(rate)
            one[set] = rate
        end
        return include
    end
end

-----------------------------------------------------------------------------------------------------------
function M.ratio(fopts)
    local set = fopts.set
    local field1 = fopts.field1
    local field2 = fopts.field2
    local ratio_lower = fopts.ratio_lower
    local ratio_upper = fopts.ratio_upper

    simple.def(fopts, "date_offset", 0)
    simple.def(fopts, "absolute", false)
    local absolute = fopts.absolute
    local date_offset = fopts.date_offset
    return function(one, series, code, currindex, opts)
        if date_offset ~= 0 then
            if series == nil then
                return false
            end
            one = series[currindex + date_offset]
        end
        if one == nil or one.empty then
            return false
        end
        local v1 = one[field1]
        local v2 = one[field2]
        if v2 == nil then
            return false
        end
        local ratio
        if v2 == 0 then
            v2 = 0.01
        end
        ratio = v1 / v2
        if ratio < 0 and absolute then
            --print("abs", ratio)
            ratio = -ratio
        end
        if set ~= nil then
            one[set] = simple.numcon(ratio)
        end
        --print(one.name, ratio)
        return ratio >= ratio_lower and ratio <= ratio_upper
    end
end
-----------------------------------------------------------------------------------------------------------

function M.field(fopts)
    local field = fopts.field
    local lower = fopts.lower
    local upper = fopts.upper
    simple.def(fopts, "date_offset", 0)
    local date_offset = fopts.date_offset
    return function(one, series, code, currindex, opts)
        if date_offset ~= 0 then
            if series == nil then
                return true
            end
            one = series[currindex + date_offset]
        end
        if one == nil or one.empty then
            return true
        end
        local v = one[field]
        return lower <= v and v <= upper
    end
end

return M