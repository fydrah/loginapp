FROM golang:1.21-alpine AS build
ARG REPO=github.com/fydrah/loginapp

RUN apk add --no-cache git build-base
COPY . /go/src/${REPO}
WORKDIR /go/src/${REPO}
RUN make

FROM scratch
ARG REPO=github.com/fydrah/loginapp
LABEL maintainer="Flavien Hardy <flav.hardy@gmail.com>"

COPY --from=build /go/src/${REPO}/build/loginapp /loginapp

ENTRYPOINT ["/loginapp"]
CMD [""]
