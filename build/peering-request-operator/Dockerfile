FROM golang:1.13 as builder
ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go
COPY . /go/src/github.com/liqoTech/liqo
WORKDIR /go/src/github.com/liqoTech/liqo
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ./cmd/peering-request-operator/
RUN cp peering-request-operator /usr/bin/peering-request-operator

FROM scratch
COPY --from=builder /usr/bin/peering-request-operator /usr/bin/peering-request-operator
ENTRYPOINT [ "/usr/bin/peering-request-operator" ]
