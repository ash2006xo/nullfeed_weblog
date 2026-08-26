FROM golang:1.25-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/nullfeed ./cmd/server

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/nullfeed ./nullfeed
COPY web ./web

ENV PORT=8080
EXPOSE 8080

CMD ["./nullfeed"]
