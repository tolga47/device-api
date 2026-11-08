FROM golang:1.27-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /device-api .


FROM alpine:3.22

RUN adduser -D app && mkdir /data && chown app /data
USER app

COPY --from=build /device-api /device-api

ENV DB_PATH=/data/devices.db
ENV GIN_MODE=release
VOLUME /data
EXPOSE 8080

ENTRYPOINT ["/device-api"]
