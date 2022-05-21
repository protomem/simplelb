FROM golang:alpine AS builder
RUN apk add --no-cache make
WORKDIR /github.com/protomem/simplelb
COPY . .
RUN CGO_ENABLED=0 GOOS=linux make build

FROM alpine:latest
RUN  apk --no-cache add ca-certificates
WORKDIR /root
COPY --from=builder /github.com/protomem/simplelb/build/lb .
ENTRYPOINT [ "/root/lb" ]

