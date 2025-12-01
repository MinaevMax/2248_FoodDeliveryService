FROM golang:1.24-alpine AS builder

COPY . /build

WORKDIR /build


COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ./foodservice cmd/server/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /build/foodservice /bin/foodservice

CMD ["/bin/foodservice"]

