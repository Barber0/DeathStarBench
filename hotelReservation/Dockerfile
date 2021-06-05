FROM golang:1.15

RUN mkdir -p /app
RUN mkdir -p /mysource

COPY . /mysource
WORKDIR /mysource

ENV GOPROXY="https://goproxy.cn,direct"
ENV GO111MODULE=on

RUN go install -ldflags="-s -w" ./cmd/...