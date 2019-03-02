FROM golang:1-alpine AS build
ARG REPO=github.com/devopy.io/loginapp

RUN apk add --no-cache git build-base
ENV GO111MODULE on
COPY . /go/src/${REPO}
WORKDIR /go/src/${REPO}
RUN make build

FROM alpine:latest

#ARG REPO=github.com/devopy.io/loginapp
RUN adduser app -u 1001 -g 1001 -s /bin/false -D rrac

COPY --from=build loginapp /loginapp
COPY --from=build assets /assets
COPY --from=build templates /templates
RUN chown app /loginapp && chown -R app /assets && chown -R app /templates

ENTRYPOINT ["/loginapp"]
CMD [""]
