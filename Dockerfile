FROM golang:1.17 as builder
WORKDIR $GOPATH/src/github.com/transferwise/oomie
COPY . $GOPATH/src/github.com/transferwise/oomie
RUN make build

FROM gcr.io/distroless/static-debian10
COPY --from=builder /go/src/github.com/transferwise/oomie/bin/oomie /oomie
ENTRYPOINT ["/oomie"]
