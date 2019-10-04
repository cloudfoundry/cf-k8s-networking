FROM golang:1.13 AS build

COPY ./ /go/src/cfroutesync/
WORKDIR /go/src/cfroutesync/
RUN go install

FROM debian:stretch-slim
COPY --from=build /go/bin/cfroutesync /cfroutesync/
WORKDIR /cfroutesync
ENTRYPOINT ["/cfroutesync/cfroutesync"]
EXPOSE 8080
