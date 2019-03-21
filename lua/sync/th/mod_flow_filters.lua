

local simple = require("common.simple")

local M = {}

function M.io(fopts)

    simple.def(fopts, "io_lower", 1)
    simple.def(fopts, "io_upper", 100)

    simple.def(fopts, "big_in_lower", 10)
    simple.def(fopts, "big_in_upper", 100)

    simple.def(fopts, "turnover", 0.5)

    simple.def(fopts, "ch_lower", -1.5)
    simple.def(fopts, "ch_upper", 6)


    --simple.table_print_all(fopts)

    return function(one, series, code, currindex, opts)
        local io = one.flow_io_rate
        local big_in = one.flow_big_in_rate
        return
            io >= fopts.io_lower and io <= fopts.io_upper
            and big_in >= fopts.big_in_lower and big_in <= fopts.big_in_upper
            and one.turnover >= fopts.turnover
            and one.change_rate >= fopts.ch_lower and one.change_rate <= fopts.ch_upper
    end
end

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

function M.codes(fopts)

    local codes = fopts.codes
    local codes_map = simple.array_to_map(codes)

    return function(one, series, code, currindex, opts)
        return codes_map[code] ~= nil
    end
end

function M.names(fopts)

    local names = fopts.codes
    local names_map = simple.array_to_map(names)

    return function(one, series, code, currindex, opts)
        return names_map[code] ~= nil
    end
end

return M