FROM golang:1.21-bullseye AS builder

# The container-side mount point of the data volume.
ARG DATA_DIR
# Path to the Go compiler cache.
ARG GOCACHE

WORKDIR /app

# By copying go.mod and go.sum files first, the dependencies will be redownloaded
# only when these files change.
COPY go.* .
RUN go mod download

COPY . .

# Mount the Go compiler cache to speed up builds.
RUN --mount=type=cache,target=$GOCACHE \
    make build

FROM gcr.io/distroless/base-debian11:nonroot AS distroless

ARG DATA_DIR
ARG PORT

COPY --from=builder --chown=nonroot:nonroot /app/bin/server /app/bin/server
COPY --from=builder --chown=nonroot:nonroot /app/data/* "$DATA_DIR/"

USER nonroot

EXPOSE $PORT

CMD ["/app/bin/server"]
