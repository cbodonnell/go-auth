# Build stage
FROM golang:1-alpine as builder

RUN apk update && apk add openssl
RUN mkdir /etc/ssl/go-auth
RUN openssl req -x509 -newkey rsa:4096 -keyout /etc/ssl/go-auth/key.pem -out /etc/ssl/go-auth/cert.pem -days 365 -nodes -subj '/CN=*'

RUN mkdir /app
WORKDIR /app

COPY . .

RUN go get -d -v ./...
RUN go build

# Production stage
FROM alpine:latest as prod

RUN mkdir /app
WORKDIR /app

COPY --from=builder /app/go-auth ./
COPY --from=builder /app/templates/* ./templates/
COPY --from=builder /etc/ssl/go-auth/* ./certs/

CMD [ "./go-auth" ]