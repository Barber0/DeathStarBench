Span = {
    influx_cli = nil,
    start_time = -1,
    end_time = -1
}

function Span:unix_ms()
    local socket = require "socket"
    return socket.gettime() * 1000
end

function Span:new(method, req)
    local o = {}
    setmetatable(o, self)
    self.__index = self

    local influxServiceStat = ngx.shared.config:get("InfluxServiceStat")
    local influxDns = ngx.shared.config:get("InfluxDns")
    local influxPortStr = ngx.shared.config:get("InfluxPort")
    local influxDBName = ngx.shared.config:get("InfluxDBName")
    local influxAuth = ngx.shared.config:get("InfluxAuth")
    local autosysService = ngx.shared.config:get("AutosysService")
    local autosysPod = ngx.shared.config:get("AutosysPod")

    local influxPort = tonumber(influxPortStr)

    local i = require "resty.influx.object"
    local influx, err = i:new({
        host = influxDns,
        port = influxPort,
        db = influxDBName,
        auth = influxAuth
    })
    if (not influx) then
        return nil, nil, err
    end

    local httpHeaders = req.get_headers()

    influx:set_measurement(influxServiceStat)

    influx:add_tag("service", autosysService)
    influx:add_tag("pod", autosysPod)
    influx:add_tag("method", method)

    influx:add_tag("src_service", "client")
    influx:add_tag("src_pod", "0")
    influx:add_tag("src_method", "client")

    local epoch = httpHeaders["epoch"] or "-1"
    influx:add_tag("epoch", epoch)

    self.influx_cli = influx
    self.start_time = self.unix_ms()

    local carrier = {
        src_service = autosysService,
        src_pod = autosysPod,
        src_method = method,
        epoch = epoch
    }

    return o, carrier, nil
end

function Span:finish()
    self.end_time = self.unix_ms()
    self.influx_cli:add_field("latency", self.end_time - self.start_time)
    self.influx_cli:buffer()
    local ok, err = self.influx_cli:flush()
    if (not ok) then
        ngx.log(ngx.ERR, "upload metric failed: ", err)
    end
end
