FROM golang:1.8-alpine3.6 AS build
ARG REPO=github.com/fydrah/loginapp

RUN apk add --no-cache git build-base
COPY . /go/src/${REPO}
WORKDIR /go/src/${REPO}
RUN make build-static

FROM scratch
ARG REPO=github.com/fydrah/loginapp
LABEL maintainer="Flavien Hardy <flav.hardy@gmail.com>"

COPY --from=build /go/src/${REPO}/bin/loginapp-static /loginapp
COPY --from=build /go/src/${REPO}/assets /assets
COPY --from=build /go/src/${REPO}/templates /templates

ENTRYPOINT ["/loginapp"]
CMD [""]
