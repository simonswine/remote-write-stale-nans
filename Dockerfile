FROM golang:1.15.6

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:3.12
RUN apk --no-cache add ca-certificates
COPY --from=0 /app/app /usr/local/bin/remote-write-stale-nans
USER nobody
CMD ["remote-write-stale-nans"]
