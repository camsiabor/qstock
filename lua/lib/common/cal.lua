local cal = {}


function cal.str2num(str, keep)
    if keep == nil then
        keep = 5
    end
    local n = string.find(str, "亿")
    if n == nil then
        str = string.gsub(str, "万", "") + 0
        str = str / 10000
    else
        str = string.gsub(str, "亿", "") + 0
    end
    return string.sub(str, 1, keep) + 0
end

function cal.num_level_criteria_init(criteria, container)
    for i = 1, #criteria do
        container[i] = 0
    end
    container[#container + 1] = 0
    return container
end

function cal.num_level_criteria_count(num, criteria, container)

    local bottom = criteria[1]
    if num < bottom then
        container[1] = container[1] + 1
        return 1
    end

    local top = criteria[#criteria]
    if num > top then
        local t = #container
        container[t] = container[t] + 1
        return t
    end

    local ncriteria = #criteria - 1
    for i = 1, ncriteria do
        local lower = criteria[i]
        local upper = criteria[i + 1]
        if lower <= num and num < upper then
            container[i + 1] = container[i + 1] + 1
            return i
        end
    end

    return 0
end


return cal