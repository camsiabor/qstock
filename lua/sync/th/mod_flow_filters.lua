

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
        if one == nil then
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

            if prev == nil or curr == nil then
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
function M.groups(fopts)
    local groups = fopts.groups
    local ngroups = #groups
    return function(one, series, code, currindex, opts)
        local one_groups = one.group
        if one_groups == nil then
            return false
        end
        local include = false
        for i = 1, ngroups do
            local group = groups[i]
            if one_groups[group] ~= nil then
                include = true
                break
            end
        end
        return include
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
return M