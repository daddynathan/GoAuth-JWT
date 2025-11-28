FROM golang:1.25.1-alpine AS builder
RUN apk update && apk add --no-cache git gcc musl-dev
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

RUN go build -o app ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/app .
COPY --from=builder /app/docs ./docs 
COPY --from=builder /app/app.env ./app.env 

CMD ["./app"]