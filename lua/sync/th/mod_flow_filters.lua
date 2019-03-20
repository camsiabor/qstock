

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

    return function(one, series, code, opts)
        local io = one.flow_io_rate
        local big_in = one.flow_big_in_rate
        return
            io >= fopts.io_lower and io <= fopts.io_upper
            and big_in >= fopts.big_in_lower and big_in <= fopts.big_in_upper
            and one.turnover >= fopts.turnover
            and one.change_rate >= fopts.ch_lower and one.change_rate <= fopts.ch_upper
    end
end

return M