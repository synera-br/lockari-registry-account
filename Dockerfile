FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go mod tidy

# RUN go build -o /manager cmd/main.go
RUN time CGO_ENABLED=0 GOOS=linux go build -v -ldflags='-s -w' -a -o /manager cmd/main.go


FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /manager .

USER 1000:1000
ENTRYPOINT ["/manager"]

EXPOSE 8080
CMD ["/manager"]