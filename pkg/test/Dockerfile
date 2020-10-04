FROM golang:alpine
WORKDIR /crawler/
COPY . .
CMD ["go", "test", "-v", "./...", "--tags=integration"]