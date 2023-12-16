FROM golang:alpine AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOPROXY="https://goproxy.cn,direct" \
    GO111MODULE=on

WORKDIR /build

COPY . .

RUN go mod download
RUN go build -ldflags="-s -w" -o /app/server ./cmd/server/main.go


FROM alpine

ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=builder /app/server /app/
COPY --from=builder /build/conf/application.yaml /app/conf/

EXPOSE 8080

CMD ["./server"]