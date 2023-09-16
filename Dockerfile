FROM golang:1.21-bullseye AS base

WORKDIR /app

# By copying go.mod and go.sum files first, the dependencies will be
# redownloaded only when these files change.
COPY go.* .
RUN go mod download

FROM base AS test

ARG GO_COMPILER_CACHE

# Mount the project directory and Go compiler cache to avoid copying
# and speed up builds.
# Uses read-write access to the project directory to generate
# coverage reports.
RUN --mount=type=cache,target=$GO_COMPILER_CACHE \
    --mount=type=bind,target=.,rw \
    go test -race -v -coverprofile=coverage.txt -covermode=atomic ./...

FROM base AS build

ARG GO_COMPILER_CACHE

RUN --mount=type=cache,target=$GO_COMPILER_CACHE \
    --mount=type=bind,target=. \
    CGO_ENABLED=1 GOFLAGS=-buildvcs=false go build -o /out/ ./cmd/server ./cmd/healthcheck

FROM gcr.io/distroless/base-debian11:nonroot AS optimized

ARG PORT

COPY --from=build --chown=nonroot:nonroot /out/ /

USER nonroot

EXPOSE $PORT

HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/healthcheck"]

CMD ["/server"]
