FROM golang:1.10-alpine3.8 as build
RUN apk add --no-cache make git
ADD . .
RUN make deps && make build


FROM alpine:latest
MAINTAINER Néstor Salceda<nestor.salceda@sysdig.com>

RUN apk add --no-cache ca-certificates

COPY  --from=build /go/falco-sns /bin/

CMD ["/bin/falco-sns"]
