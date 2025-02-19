local _M = {}

local function _StrIsEmpty(s)
    return s == nil or s == ''
end

function _M.Follow()
    local ngx = ngx
    local GenericObjectPool = require "GenericObjectPool";
    GenericObjectPool:setTimeout(120000)
    local SocialGraphServiceClient = require"social_network_SocialGraphService".SocialGraphServiceClient
    local jwt = require "resty.jwt"

    local req_id = tonumber(string.sub(ngx.var.request_id, 0, 15), 16)

    require "influx/influx"
    local span, carrier, err = Span:new("Follow", ngx.req)
    if (not span or not carrier) then
        ngx.log(ngx.ERR, "start span failed: ", err)
        carrier = {}
    end

    ngx.req.read_body()
    local post = ngx.req.get_post_args()

    local client = GenericObjectPool:connection(SocialGraphServiceClient,
        "social-graph-service.social-network.svc.cluster.local", 9090)

    -- -- new start -- 
    -- if (_StrIsEmpty(ngx.var.cookie_login_token)) then
    --   ngx.status = ngx.HTTP_UNAUTHORIZED
    --   ngx.exit(ngx.HTTP_OK)
    -- end

    -- local login_obj = jwt:verify(ngx.shared.config:get("secret"), ngx.var.cookie_login_token)
    -- if not login_obj["verified"] then
    --   ngx.status = ngx.HTTP_UNAUTHORIZED
    --   ngx.say(login_obj.reason);
    --   ngx.exit(ngx.HTTP_OK)
    -- end
    -- -- get user id/name from login obj
    -- local user_id = tonumber(login_obj["payload"]["user_id"])
    -- local username = login_obj["payload"]["username"]

    -- -- new end --

    local status
    local err
    if (not _StrIsEmpty(post.user_id) and not _StrIsEmpty(post.followee_id)) then
        status, err = pcall(client.Follow, client, req_id, tonumber(post.user_id), tonumber(post.followee_id), carrier)
    elseif (not _StrIsEmpty(post.user_name) and not _StrIsEmpty(post.followee_name)) then
        status, err = pcall(client.FollowWithUsername, client, req_id, post.user_name, post.followee_name, carrier)
    else
        ngx.status = ngx.HTTP_BAD_REQUEST
        ngx.say("Incomplete arguments")
        ngx.log(ngx.ERR, "Incomplete arguments")
        ngx.exit(ngx.HTTP_BAD_REQUEST)
    end

    if not status then
        ngx.status = ngx.HTTP_INTERNAL_SERVER_ERROR
        ngx.say("Follow Failed: " .. err.message)
        ngx.log(ngx.ERR, "Follow Failed: " .. err.message)
        ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
    end

    GenericObjectPool:returnConnection(client)

    if span then
        span:finish()
    end
    ngx.redirect("../../contact.html")
    -- ngx.header.content_type = "application/json; charset=utf-8"
    -- ngx.say(cjson.encode(home_timeline) )
    ngx.exit(ngx.HTTP_OK)

end

return _M
