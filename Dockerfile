FROM golang:1.21-bullseye AS builder

# The container-side mount point of the data volume.
ARG DATA_DIR

WORKDIR /app

COPY . .

RUN go mod download
RUN make build

FROM gcr.io/distroless/base-debian11:nonroot AS distroless

ARG DATA_DIR
ARG PORT

COPY --from=builder --chown=nonroot:nonroot /app/bin/server /app/bin/server
COPY --from=builder --chown=nonroot:nonroot /app/data/* "$DATA_DIR/"


USER nonroot

EXPOSE $PORT

CMD ["/app/bin/server"]
