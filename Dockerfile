FROM quay.io/wasilak/golang:1.23 AS builder

ADD ./app /app
WORKDIR /app
RUN mkdir -p ./dist
RUN CGO_ENABLED=0  go build -o ./dist/tools

FROM scratch

COPY --from=builder /app/dist/tools .

ENV USER=root

ENTRYPOINT ["/tools", "--listen=0.0.0.0:3000"]
