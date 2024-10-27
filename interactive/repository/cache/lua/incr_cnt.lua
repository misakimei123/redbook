local key = KEYS[1]
local cntKey = ARGS[1]
local delta = ARGS[2]
local exist = redis.call('EXISTS', key)
if exist == 0 then
    return 0
end

redis.call('HINCRBY', key, cntKey, delta)
return 1