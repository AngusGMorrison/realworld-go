FROM golang:1.21-bullseye AS base

WORKDIR /app

# By copying go.mod and go.sum files first, the dependencies will be
# redownloaded only when these files change.
COPY go.* .
RUN go mod download

FROM base AS build

ARG GO_COMPILER_CACHE

RUN --mount=type=cache,target=$GO_COMPILER_CACHE \
    --mount=type=bind,target=. \
    go build -o /out/ ./cmd/server ./cmd/healthcheck

FROM gcr.io/distroless/base-debian11:nonroot AS optimized

ARG PORT

COPY --from=build --chown=nonroot:nonroot /out/ /

USER nonroot

EXPOSE $PORT

HEALTHCHECK --interval=5s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/healthcheck"]

CMD ["/server"]
