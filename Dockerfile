FROM golang:1.16-buster AS builder
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ./
RUN go env -w GOFLAGS=-mod=mod
RUN go build -v -o server
CMD ["/app/server"]


# FROM scratch
# # RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
# #     ca-certificates && \
# #     rm -rf /var/lib/apt/lists/*
# COPY --from=builder /app/server /app/server
# CMD ["/app/server"]