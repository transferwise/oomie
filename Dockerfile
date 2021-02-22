FROM golang:1.15 as builder
WORKDIR $GOPATH/src/github.com/kgtw/oomie
COPY . $GOPATH/src/github.com/kgtw/oomie
RUN make build

FROM gcr.io/distroless/static-debian10
COPY --from=builder /go/src/github.com/kgtw/oomie/bin/oomie /oomie
ENTRYPOINT ["/oomie"]
