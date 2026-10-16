FROM golang:1.27-alpine AS build

WORKDIR /src

COPY go.mod ./
COPY . .
RUN mkdir -p /out && CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/api ./cmd/api

FROM alpine:3.22

COPY --from=build /out/api /api

EXPOSE 8080
ENV PORT=8080

USER 65534:65534
ENTRYPOINT ["/api"]