FROM golang:1-alpine AS build
ARG REPO=github.com/devopy.io/loginapp

RUN apk add --no-cache git build-base
ENV GO111MODULE on
COPY . /go/src/${REPO}
WORKDIR /go/src/${REPO}
RUN make build

FROM scratch
ARG REPO=github.com/devopy.io/loginapp

COPY --from=build /go/src/${REPO}/bin/loginapp /loginapp
COPY --from=build /go/src/${REPO}/assets /assets
COPY --from=build /go/src/${REPO}/templates /templates

ENTRYPOINT ["/loginapp"]
CMD [""]
