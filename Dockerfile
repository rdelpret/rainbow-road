FROM golang:1.16.4-buster as builder
RUN mkdir rainbow-road
WORKDIR /rainbow-road
COPY / .
RUN go mod download
RUN go mod verify
RUN make build
FROM debian:buster
COPY --from=builder /rainbow-road/rainbow-road-server .
EXPOSE 9999
CMD ["./rainbow-road-server"]
