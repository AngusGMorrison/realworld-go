FROM golang:1.20.3-bullseye

ARG volume_mount_path
ARG port

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o ./server ./cmd/server

RUN adduser \
  --disabled-password \
  --gecos "" \
  --home /app \
  --no-create-home \
  --uid 65532 \
  docker

RUN mkdir $volume_mount_path && chown -R docker:docker $volume_mount_path
RUN chmod 700 $volume_mount_path

USER docker:docker

EXPOSE $port

ENTRYPOINT ["bash", "/app/entrypoint.sh"]
