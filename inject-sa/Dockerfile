FROM golang:1.13 AS build

RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN go build -o webhook-server cmd/main.go cmd/controller.go

FROM cloudfoundry/run:tiny
COPY --from=build /build/webhook-server /app/
WORKDIR /app
ENTRYPOINT ["./webhook-server"]
