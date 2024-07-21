ARG ARCH="amd64"
ARG OS="linux"
FROM golang:alpine3.20 as builder

WORKDIR /src
COPY ./. ./
RUN go build -o /build/${OS}-${ARCH}/smiles_exporter

ARG ARCH="amd64"
ARG OS="linux"
FROM golang:alpine3.20

LABEL authors="koraktor"

ENV USERNAME username
ENV PASSWORD password

COPY --from=builder /build/${OS}-${ARCH}/smiles_exporter /bin/smiles_exporter

EXPOSE 9776
ENTRYPOINT /bin/smiles_exporter --username=${USERNAME} --password=${PASSWORD}
