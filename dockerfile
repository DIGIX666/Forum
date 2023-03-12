#syntax=docker/dockerfile:1

FROM golang:1.20.2

WORKDIR /Forum

COPY . .

RUN go build -o server .

EXPOSE 8080

CMD ["./server"]