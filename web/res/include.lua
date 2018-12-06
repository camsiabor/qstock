
local hgetall = function (key)
  local bulk = redis.call('HGETALL', key)
	local result = {}
	local nextkey
	for i, v in ipairs(bulk) do
		if i % 2 == 1 then
			nextkey = v
		else
			result[nextkey] = v
		end
	end
	return result
end

local hmget = function (key, ...)
	if next(arg) == nil then return {} end
	local bulk = redis.call('HMGET', key, unpack(arg))
	local result = {}
	for i, v in ipairs(bulk) do result[ arg[i] ] = v end
	return result
end

local tonum = function(s)
  return s + 0
end

local tonumex = function(s, dv)
    local ok, n = pcall(tonum, s);
    if ok then
        return n
    else
        return dv
    end
end