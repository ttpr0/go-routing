FROM golang:1.19

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-routing

EXPOSE 5002

# Run
CMD ["/go-routing"]
