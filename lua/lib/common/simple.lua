

local simple = {}

function simple.str2num(str, keep)
    if keep == nil then
        keep = 5
    end
    local n = string.find(str, "亿" )
    if n == nil then
        str = string.gsub(str, "万", "") + 0
        str = str / 10000
    else
        str = string.gsub(str, "亿", "") + 0
    end
    return string.sub(str, 1, keep) + 0
end

function simple.numcon(num)
    return string.sub(num.."", 1, 5) + 0
end

function simple.nozero(num)
    if num <= 0 then
        num = 0.0001
    end
    return num
end

function simple.table_sort(t, field)
    local n = #t
    for i = 1, n do
        for j = 1, n - i do
            local a = t[j]
            local b = t[j + 1]
            if a[field] < b[field] then
                t[j] = b
                t[j + 1] = a
            end
        end
    end
end

return simple
