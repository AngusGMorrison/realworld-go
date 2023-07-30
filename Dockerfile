FROM golang:1.20.3-bullseye AS builder

ARG DATA_DIR

WORKDIR /app

COPY . .

RUN mkdir -p $DATA_DIR

RUN go mod download
RUN make build

FROM gcr.io/distroless/base-debian11:nonroot AS production

ARG CERTS_DIR
ARG DATA_DIR
ARG PORT

COPY --from=builder --chmod=700 --chown=nonroot:nonroot /app/bin/server /app/bin/server
COPY --from=builder --chmod=700 --chown=nonroot:nonroot $CERTS_DIR $CERTS_DIR
COPY --from=builder --chmod=700 --chown=nonroot:nonroot $DATA_DIR $DATA_DIR

EXPOSE $PORT

CMD ["/app/bin/server"]
