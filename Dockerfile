# FROM golang:alpine AS builder
# ENV DIR=/go/code/basic-server
# WORKDIR ${DIR}
# COPY . .
# RUN CGO_ENABLED=0 GOOS=linux go build -o main
# RUN go build -o main
# RUN ls -la
# CMD ["./main"]
# EXPOSE 8085

FROM golang:1.12.5
WORKDIR /usr/go/app
COPY . /usr/go/app
RUN CGO_ENABLED=0 GOOS=linux go build -o main
ENTRYPOINT [ "./main" ]