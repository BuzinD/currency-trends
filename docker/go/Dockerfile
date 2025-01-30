FROM golang:1.23.3-alpine AS builder

RUN apk --no-cache add bash git make gcc gettext musl-dev

WORKDIR /usr/local/src

COPY ["data-fetcher/go.mod", "data-fetcher/go.sum", "./"]

RUN go mod download

COPY data-fetcher ./

RUN go build -o ./bin/app cmd/main.go

FROM alpine

COPY --from=builder /usr/local/src/bin/app /app/app

COPY --from=builder /usr/local/src/env /app/env

WORKDIR /app/

CMD ["/app/app"]
