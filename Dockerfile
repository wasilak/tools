FROM quay.io/wasilak/golang:1.23 AS builder

ADD ./app /app
WORKDIR /app
RUN mkdir -p ./dist
RUN go build -o ./dist/tools

FROM scratch

COPY --from=builder /app/dist/tools /tools

ENTRYPOINT ["/tools", "--listen=0.0.0.0:3000"]
