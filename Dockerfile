FROM golang:1.20.3-bullseye

ARG goarch
ARG goos
ARG port
ARG volume_mount_path
ARG workdir

WORKDIR $workdir

COPY . .

RUN go mod download

RUN GOOS=$goos GOARCH=$goarch make build

RUN adduser \
  --disabled-password \
  --gecos "" \
  --home $workdir \
  --no-create-home \
  --uid 65532 \
  docker

RUN mkdir $volume_mount_path && chown -R docker:docker $volume_mount_path
RUN chmod 700 $volume_mount_path

RUN chown -R docker:docker /app
RUN chmod -R 700 /app/certs

USER docker:docker

EXPOSE $port

CMD /app/bin/server
