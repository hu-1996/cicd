# Build the manager binary
FROM golang:1.23.4 AS builder
ARG TARGETOS
ARG TARGETARCH
ARG MODULE

WORKDIR /workspace
COPY . .
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o server ${MODULE}/main.go

FROM alpine:3.20.0
WORKDIR /
COPY --from=builder /workspace/server .

ENTRYPOINT ["/server"]
