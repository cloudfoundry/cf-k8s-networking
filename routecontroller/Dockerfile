FROM golang:1.13 AS build

COPY ./ /go/src/routecontroller/
WORKDIR /go/src/routecontroller/
RUN go install

FROM cloudfoundry/run:tiny
COPY --from=build /go/bin/routecontroller /routecontroller/
WORKDIR /routecontroller
ENTRYPOINT ["/routecontroller/routecontroller"]
