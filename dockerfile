#syntax=docker/dockerfile:1

FROM golang:1.24.3-alpine3.19

WORKDIR /Forum

COPY . .

RUN go build -o server .

EXPOSE 8080

RUN useradd -U -u 1000 appuser && \
    chown -R 1000:1000 /Forum

USER 1000
CMD ["./server"]