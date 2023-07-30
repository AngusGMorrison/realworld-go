FROM golang:1.20.3-bullseye AS builder

# The container-side mount point of the data volume.
ARG DATA_DIR

WORKDIR /app

COPY . .

# Create the data directory if it doesn't exist. It may already exist locally
# if the application has previously been run with Air.
# RUN mkdir -p $DATA_DIR

RUN go mod download
RUN make build

FROM gcr.io/distroless/base-debian11:nonroot AS production

# The container-side mount point of
ARG DATA_DIR
ARG PORT

COPY --from=builder --chown=nonroot:nonroot /app/bin/server /app/bin/server


USER nonroot

EXPOSE $PORT

CMD ["/app/bin/server"]
