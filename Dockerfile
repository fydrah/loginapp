FROM golang:1.13-alpine AS build
ARG REPO=github.com/fydrah/loginapp

RUN apk add --no-cache git build-base
COPY . /go/src/${REPO}
WORKDIR /go/src/${REPO}
RUN make build-static

FROM scratch
ARG REPO=github.com/fydrah/loginapp
LABEL maintainer="Flavien Hardy <flav.hardy@gmail.com>"

COPY --from=build /go/src/${REPO}/build/loginapp-static /loginapp
COPY --from=build /go/src/${REPO}/web/assets /web/assets
COPY --from=build /go/src/${REPO}/web/templates /web/templates

ENTRYPOINT ["/loginapp"]
CMD [""]
