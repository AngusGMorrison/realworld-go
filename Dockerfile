FROM golang:1.21-bullseye AS base

# Path to the Go compiler cache.
ARG GOCACHE

WORKDIR /app

# By copying go.mod and go.sum files first, the dependencies will be redownloaded
# only when these files change.
COPY go.* .
RUN go mod download

COPY . .

FROM base AS build
# Mount the Go compiler cache to speed up builds.
RUN --mount=type=cache,target=$GOCACHE \
    make build

FROM gcr.io/distroless/base-debian11:nonroot AS distroless

ARG PORT

COPY --from=build --chown=nonroot:nonroot /app/bin /app/bin

USER nonroot

EXPOSE $PORT

HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/bin/healthcheck"]

CMD ["/app/bin/server"]
