local key = KEYS[1]
local cnt = tonumber(ARGV[1])

local cur = redis.call("GET", key)
if not cur then
  return -2
end

cur = tonumber(cur)
if cur < cnt then
  return -1
end

local remain = redis.call("DECRBY", key, cnt)
return remain
