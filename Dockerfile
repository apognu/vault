#
# Build static binary
#

FROM golang:1.8-alpine as builder
LABEL maintainer "Antoine POPINEAU <antoine.popineau@appscho.com>"

WORKDIR /go/src/github.com/apognu/vault

RUN apk -U add git && go get github.com/Masterminds/glide
COPY . /go/src/github.com/apognu/vault
RUN glide install
RUN CGO_ENABLED=0 GOOS=linux go build \
  -ldflags='-s' \
  -a -installsuffix cgo \
  github.com/apognu/vault

#
# Build runnable lightweight image
#

FROM alpine:latest
LABEL maintainer "Antoine POPINEAU <antoine.popineau@appscho.com>"

RUN apk -U add git
COPY --from=builder /go/src/github.com/apognu/vault/vault /vault
COPY docker-entrypoint.sh /entrypoint.sh

ENTRYPOINT [ "/entrypoint.sh" ]
