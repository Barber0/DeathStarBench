#ifndef SOCIAL_NETWORK_INFLUX_CLIENT_H
#define SOCIAL_NETWORK_INFLUX_CLIENT_H

#include <iostream>
#include <memory>
#include <InfluxDBFactory.h>
#include <chrono>
#include <map>
#include <stdlib.h>
#include <sstream>
#include "logger.h"
#include "utils.h"
#include <mutex>

#define InfluxDns "InfluxDns"
#define InfluxPort "InfluxPort"
#define InfluxDBName "InfluxDBName"
#define InfluxAuth "InfluxAuth"
#define INFLUX_SVC_STAT "InfluxServiceStat"
#define InfluxBatchSize "InfluxBatchSize"

#define INFLUX_CONN_STR_DEFAULT "http://autosys:12345678@influxdb.autosys:8086/?db=autosys"
#define INFLUX_SVC_STAT_DEFAULT "service_metric"
#define AUTOSYS_SVC "AutosysService"
#define AUTOSYS_POD "AutosysPod"

#define LabelServiceName "service"
#define LabelMethod "method"
#define LabelSrcService "src_service"
#define LabelSrcPod "src_pod"
#define LabelSrcMethod "src_method"
#define LabelFailed "failed"
#define LabelDbOpType "op"
#define LabelEpoch "epoch"

#define LabelPodName "pod"
#define LabelPodIP "ip"
#define LabelSrcIP "src_ip"

#define PerfLatency "latency"

#define UnKnown "unknown"

#define time_now std::chrono::system_clock::now()
#define unix_ms_duration(dur) std::chrono::duration_cast<std::chrono::microseconds>(dur)

#define INFLUX_CLIENT_PTR std::shared_ptr<InfluxClient>
#define INFLUX_CLIENT_VAR _influxCli

#define NEW_INFLUX_CLIENT INFLUX_CLIENT_PTR INFLUX_CLIENT_VAR = std::make_shared<InfluxClient>();
#define CLOSE_INFLUX_CLIENT INFLUX_CLIENT_VAR->Close();

#define INFLUX_CLIENT_PLACEHOLDER INFLUX_CLIENT_PTR INFLUX_CLIENT_VAR
#define ANNOUNCE_INFLUX_CLIENT INFLUX_CLIENT_PLACEHOLDER;

#define INJECT_INFLUX_CLIENT(name) INFLUX_CLIENT_VAR((name))
#define INJECT_INFLUX_CLIENT_DEFAULT INFLUX_CLIENT_VAR(INFLUX_CLIENT_VAR)

#define START_SPAN_WITH_CARRIER(name, carrier) \
    std::shared_ptr<InfluxSpan> span_##name = INFLUX_CLIENT_VAR->StartSpan(#name, (carrier));

#define START_SPAN_WITH_CARRIER_AND_DOWNSTREAM(name, carrier) \
    START_SPAN_WITH_CARRIER(name, carrier)                    \
    const std::map<std::string, std::string> next_carrier_##name = span_##name->NextCarrier();

#define START_SPAN(name) START_SPAN_WITH_CARRIER_AND_DOWNSTREAM(name, carrier)

namespace social_network
{
    std::string getEnv(std::string key, std::string defaultValue)
    {
        char *tmpStr;
        if (!(tmpStr = getenv(key.c_str())))
        {
            return defaultValue;
        }
        std::string outStr(tmpStr);
        return outStr;
    }

    int getIntEnv(std::string key, int defaultValue)
    {
        int outVal;
        char *tmpStr;
        if (!(tmpStr = getenv(key.c_str())))
        {
            return defaultValue;
        }
        sscanf(tmpStr, "%d", &outVal);
        return outVal;
    }

    class InfluxClient;
    class InfluxSpan
    {
    private:
        std::chrono::system_clock::time_point startTime;
        std::chrono::system_clock::time_point endTime;
        InfluxClient *influxCli;
        std::string method;

        std::string src_service;
        std::string src_pod;
        std::string src_method;
        std::string epoch;

        bool failed = false;

    public:
        InfluxSpan(
            InfluxClient *cli,
            std::string method,
            const std::map<std::string, std::string> &carrier);

        void Finish();

        const std::map<std::string, std::string> NextCarrier();
    };

    class InfluxClient
    {

    private:
        std::unique_ptr<influxdb::InfluxDB> influxCli;
        influxdb::Point copyPoint(influxdb::Point p) { return p; }
        std::mutex mu;

    public:
        std::string service;
        std::string pod;
        std::string service_metric;

        InfluxClient();

        void Close()
        {
            influxCli->flushBuffer();
        }

        void Write(influxdb::Point p)
        {
            try
            {
                const std::lock_guard<std::mutex> lock(mu);
                influxCli->write(copyPoint(p));
            }
            catch (const std::exception &e)
            {
                std::cerr << "write influx Point failed: " << e.what() << '\n';
            }
        }

        std::shared_ptr<InfluxSpan> StartSpan(
            std::string method,
            const std::map<std::string, std::string> &carrier);
    };

    InfluxSpan::InfluxSpan(
        InfluxClient *cli,
        std::string method,
        const std::map<std::string, std::string> &carrier) : influxCli(cli), method(method)
    {
        try
        {
            src_service = carrier.at(LabelSrcService);
            src_pod = carrier.at(LabelSrcPod);
            src_method = carrier.at(LabelSrcMethod);
            epoch = carrier.at(LabelEpoch);
        }
        catch (std::out_of_range e)
        {
            failed = true;
            LOG(error) << "build span(" << influxCli->service << "/" << influxCli->pod << "/" << method << ") failed: " << e.what();
        }
    }

    void InfluxSpan::Finish()
    {
        if (failed)
            return;
        endTime = time_now;
        long long latency = unix_ms_duration(endTime - startTime).count();

        influxdb::Point influxPoint = influxdb::Point{influxCli->service_metric}

                                          .addTag(LabelSrcService, src_service)
                                          .addTag(LabelSrcPod, src_pod)
                                          .addTag(LabelSrcMethod, src_method)
                                          .addTag(LabelEpoch, epoch)

                                          .addTag(LabelServiceName, influxCli->service)
                                          .addTag(LabelPodName, influxCli->pod)
                                          .addTag(LabelMethod, method)
                                          .addField(PerfLatency, latency);
        influxCli->Write(influxPoint);
    }

    const std::map<std::string, std::string> InfluxSpan::NextCarrier()
    {
        std::map<std::string, std::string> nextCarrier;
        if (failed)
            return nextCarrier;

        nextCarrier.insert({LabelSrcService, influxCli->service});
        nextCarrier.insert({LabelSrcPod, influxCli->pod});
        nextCarrier.insert({LabelSrcMethod, method});
        nextCarrier.insert({LabelEpoch, epoch});
        return nextCarrier;
    }

    InfluxClient::InfluxClient()
    {
        auto influxDns = getEnv(InfluxDns, "influxdb.autosys.svc.cluster.local");
        auto influxPort = getEnv(InfluxPort, "8086");
        auto influxDbName = getEnv(InfluxDBName, "ingestion");
        auto influxAuth = getEnv(InfluxAuth, "autosys:00000000");
        std::string connStr = std::string("http://") + influxAuth + std::string("@") + influxDns + std::string("/?db=") + influxDbName;

        this->influxCli = influxdb::InfluxDBFactory::Get(connStr);

        auto batchSize = getIntEnv(InfluxBatchSize, 10000);
        influxCli->batchOf(batchSize);

        this->service_metric = getEnv(INFLUX_SVC_STAT, INFLUX_SVC_STAT_DEFAULT);
        this->service = getEnv(AUTOSYS_SVC, UnKnown);
        this->pod = getEnv(AUTOSYS_POD, "0");
    }

    std::shared_ptr<InfluxSpan> InfluxClient::StartSpan(
        std::string method,
        const std::map<std::string, std::string> &carrier)
    {
        auto span = std::make_shared<InfluxSpan>(
            this,
            method,
            carrier);
        return span;
    }
}

#endif