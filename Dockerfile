FROM golang:1.16 AS builder

ARG OS="linux"
ARG ARCH="amd64"

WORKDIR /go/src/app
COPY go.mod go.sum Makefile ./
RUN make deps
COPY . .
RUN OS_ARCH=${OS}/${ARCH} make release

FROM alpine:3.13

ARG OS="linux"
ARG ARCH="amd64"

WORKDIR /app
COPY --from=builder /go/src/app/bin/driftctl_${OS}_${ARCH} /bin/driftctl
ENTRYPOINT ["/bin/driftctl"]
