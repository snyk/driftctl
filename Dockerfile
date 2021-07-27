FROM golang:1.16 AS builder

ARG OS="linux"
ARG ARCH="amd64"

WORKDIR /go/src/app
COPY go.mod go.sum Makefile ./
RUN make deps
COPY . .
RUN make release

FROM alpine:3.13

ARG OS="linux"
ARG ARCH="amd64"

WORKDIR /app
COPY --from=builder /go/src/app/bin/driftctl_${OS}_${ARCH}/driftctl /bin/driftctl
RUN chmod +x /bin/driftctl
ENTRYPOINT ["/bin/driftctl"]
