FROM golang:1-alpine AS build

RUN apk update && apk add make git gcc musl-dev

ADD . /go/src/github.com/devopyio/loginapp

WORKDIR /go/src/github.com/devopyio/loginapp

ENV GO111MODULE on
RUN make build

FROM alpine:latest

RUN adduser app -u 1001 -g 1001 -s /bin/false -D rrac

COPY --from=build /go/src/github.com/devopyio/loginapp/loginapp /usr/bin/loginapp
COPY --from=build /go/src/github.com/devopyio/loginapp/assets /assets
COPY --from=build /go/src/github.com/devopyio/loginapp/templates /templates
RUN chown app /usr/bin/loginapp && chown -R app /assets && chown -R app /templates

USER app
ENTRYPOINT ["/usr/bin/loginapp"]
