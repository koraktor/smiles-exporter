ARG ARCH="amd64"
ARG OS="linux"
FROM golang:alpine3.20 AS builder

WORKDIR /src
COPY ./. ./
RUN go build -o /build/${OS}-${ARCH}/smiles_exporter

ARG ARCH="amd64"
ARG OS="linux"
FROM golang:alpine3.20

RUN apk upgrade --no-cache

LABEL authors="koraktor"

COPY --from=builder /build/${OS}-${ARCH}/smiles_exporter /bin/smiles_exporter

EXPOSE 9776
ENTRYPOINT [ "/bin/sh", "-c", "/bin/smiles_exporter --username=${USERNAME} --password=${PASSWORD} --listen-address=${LISTEN_ADDRESS:-:9776} --log-level=${LOG_LEVEL:-warn} ${*}", "entrypoint" ]
