FROM golang:1.10 AS build

ARG REF
ARG REV
ENV DEP_VERSION=v0.4.1 \
    GOOS=linux \
    GOARCH=amd64 \
    CGO_ENABLED=0 \
    GODEBUG=netdns=cgo
RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/$DEP_VERSION/dep-linux-amd64 \
    && chmod +x /usr/local/bin/dep \
    && mkdir -p /go/src/github.com/lumoslabs/nginx_request_exporter

WORKDIR /go/src/github.com/lumoslabs/nginx_request_exporter
COPY Gopkg.* ./

RUN dep ensure -v -vendor-only

COPY . ./
RUN go build -v -a \
      -ldflags "-s -X main.REFERENCE=$REF -X main.REVISION=$REV" \
      -installsuffix cgo -o /nginx_request_exporter . \
    && chmod -v +x /nginx_request_exporter

FROM scratch
EXPOSE 9147 9514/udp
ENTRYPOINT ["/nginx_request_exporter"]
COPY --from=build /nginx_request_exporter /
