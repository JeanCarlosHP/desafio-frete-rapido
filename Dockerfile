FROM golang:1.24.1-alpine AS build

WORKDIR /app

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN apk add --no-cache ca-certificates

COPY . .

RUN go build -ldflags="-s -w" -o app ./cmd/main.go

# FROM build AS test-stage
# RUN go test -v ./...

FROM scratch AS final

ENV GO_ENV=production

WORKDIR /app

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY --from=build ./app/app ./app

EXPOSE 8080

USER 1001

ENTRYPOINT ["./app"]