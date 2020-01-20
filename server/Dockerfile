FROM golang:alpine
ENV GO111MODULE=on
WORKDIR /usr/src/app
COPY . .
RUN go build
EXPOSE 8080
ENTRYPOINT ["/usr/src/app/pty"]
