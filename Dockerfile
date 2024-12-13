FROM golang:1.22-alpine

WORKDIR /app

COPY . .
RUN go mod tidy
RUN go mod download

RUN go build -o weather-server .

EXPOSE 8080

CMD ["./weather-server"]