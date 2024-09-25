FROM golang:1.22-alpine AS builder

WORKDIR /usr/local/src

COPY ["go.mod", "./"]
COPY . .

RUN go build -o ./api ./main.go

FROM alpine AS runner

COPY --from=builder /usr/local/src/api /

CMD ["ash", "-c", "/api"]

EXPOSE 8080