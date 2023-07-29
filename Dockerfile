FROM golang:1.20.3-bullseye

ARG certs_dir
ARG goarch
ARG goos
ARG port
ARG user
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
  $user

RUN mkdir -p $volume_mount_path
RUN chown -R $user:$user $volume_mount_path
RUN chmod 700 $volume_mount_path

RUN chown -R $user:$user $workdir
RUN chmod -R 700 $certs_dir

USER $user:$user

EXPOSE $port

CMD /app/bin/server
