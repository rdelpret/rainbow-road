FROM golang:1.16.4-buster
RUN mkdir rainbow-road
WORKDIR /rainbow-road
COPY / .
RUN go mod download
RUN go mod verify
RUN cd server && go install
EXPOSE 9999
CMD ["server"]
