FROM golang:alpine AS builder
WORKDIR /work/
COPY . .
RUN GOOS=linux GOARCH=amd64 go build -mod vendor -ldflags="-w -s" -o ./bin/crawler ./cmd/service/main.go


FROM golang:alpine
WORKDIR /project/
COPY --from=builder /work/bin/crawler .
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
CMD ["./crawler"]