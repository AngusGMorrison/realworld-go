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

COPY --from=build --chown=nonroot:nonroot /app/bin/server /app/bin/server

USER nonroot

EXPOSE $PORT

CMD ["/app/bin/server"]
