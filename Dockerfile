FROM golang:alpine AS builder

ADD ./main.go /build/main.go
ADD ./go.mod /build/go.mod
ADD ./go.sum /build/go.sum

RUN cd /build && CGO_ENABLED=0 go build -v -a -ldflags '-extldflags "-static"' -o server .

FROM centurylink/ca-certs

COPY --from=builder /build/server /server

ENTRYPOINT [ "/server" ]