# Build stage
FROM golang:1-alpine as builder

RUN apk add git

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

CMD [ "./go-auth" ]