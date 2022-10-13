FROM golang:1-alpine as builder

RUN apk update
# Add required certificates to be able to call HTTPS endpoints.
RUN apk add --no-cache ca-certificates git tzdata

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . .


# CGO_ENABLED=0 == Don't depend on libc (bigger but more independent binary)
RUN env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main

FROM scratch
WORKDIR /app

# Import the Certificate-Authority certificates for enabling HTTPS.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo/Europe/Oslo /usr/share/zoneinfo/Europe/Oslo

COPY --from=builder /app/main .
COPY --from=builder /app/index.html .

CMD ["./main"]
