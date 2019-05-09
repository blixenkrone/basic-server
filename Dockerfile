FROM golang:1.12.5 AS builder
ENV DIR=/go/code/basic-server
WORKDIR ${DIR}
COPY . .
# RUN CGO_ENABLED=0 GOOS=linux go build -o main
RUN go mod verify
RUN go build -o main

# FROM alpine:latest
# RUN apk --no-cache add ca-certificates
# WORKDIR /root/
# COPY --from=builder ${DIR} .
RUN ls -la
CMD ["./main"]
EXPOSE 8085
# COPY --from=nginx:latest /etc/nginx/nginx.conf /nginx.conf