FROM golang:alpine AS build

WORKDIR /go/src

RUN apk --no-cache add make git gcc libtool musl-dev ca-certificates dumb-init 

ARG ENTRYPOINT

COPY . .

RUN go mod download

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o app ./cmd/${ENTRYPOINT}

FROM alpine:latest

WORKDIR /app

COPY --from=build /go/src/app .

CMD ["./app"]

