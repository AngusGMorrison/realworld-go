FROM golang:1.20.3-bullseye

ARG volume_mount_path
ARG port

WORKDIR /app

COPY . .

RUN go mod download

RUN GOOS=linux GOARCH=arm64 make build

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

CMD /app/bin/server
