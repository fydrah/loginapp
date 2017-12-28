FROM golang:alpine3.6 AS build

ADD . /app/
WORKDIR /app
RUN apk update && \
    apk add git build-base && \
    go get -d && \
    go build -o login-app


FROM alpine:3.6
MAINTAINER Flavien Hardy <flav.hardy@gmail.com>

COPY --from=build /app/login-app /

ENTRYPOINT ["/login-app"]
CMD ["/app/config.yaml"]
