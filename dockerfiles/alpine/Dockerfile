FROM golang:1.8-alpine3.6 AS build

RUN apk add --no-cache git build-base
ADD . /build/
WORKDIR /build
RUN go get -d ./...
RUN make

FROM alpine:3.6
MAINTAINER Flavien Hardy <flav.hardy@gmail.com>

COPY --from=build /build/loginapp /

ENTRYPOINT ["/loginapp"]
CMD [""]
