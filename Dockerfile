FROM alpine:3.4

ARG VERSION=master
ARG IMPORT_PATH=github.com/alkar/odyn

ADD . /go/src/${IMPORT_PATH}

RUN apk add --no-cache \
        -X http://dl-cdn.alpinelinux.org/alpine/edge/community \
        ca-certificates \
        gcc \
        musl-dev \
        git \
        go \
        go-tools \
        glide \
  && export GOPATH=/go \
  && export PATH=${GOPATH}/bin:${PATH} \
  && cd ${GOPATH}/src/${IMPORT_PATH} \
  && glide i \
  && CGO_ENABLED=0 go test -v -cover $(glide nv) \
  && CGO_ENABLED=0 go build -v -o odyn -ldflags '-s -X "main.appVersion=${VERSION}" -extldflags "-static"' ${IMPORT_PATH}/cli \
  && mv odyn / \
  && apk del --no-cache gcc musl-dev git go go-tools glide \
  && rm -rf $GOPATH ~/.glide

ENTRYPOINT ["/odyn"]
