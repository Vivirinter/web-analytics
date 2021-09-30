FROM golang:latest as builder
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go build -o main

FROM busybox:glibc
WORKDIR /bin/app
COPY --from=builder /app/main .
EXPOSE 6379
CMD ["/bin/app/main"]
