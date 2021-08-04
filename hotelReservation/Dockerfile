FROM golang:1.15

RUN mkdir -p /app
RUN mkdir -p /mysource

COPY . /mysource
WORKDIR /mysource
RUN go install -ldflags="-s -w" ./cmd/...