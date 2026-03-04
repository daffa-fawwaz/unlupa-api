FROM golang:1.24-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates tzdata
ENV GOTOOLCHAIN=auto
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -o /app/hifzhun-api ./

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/hifzhun-api /app/hifzhun-api
RUN mkdir -p /app/data
COPY --from=builder /src/data/surah.json /app/data/surah.json
ENV APP_PORT=3000
EXPOSE 3000
CMD ["/app/hifzhun-api"]
