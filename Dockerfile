FROM golang:1.19

WORKDIR /go/src/twowaysql
COPY go.* .
RUN go mod download
COPY . .

RUN go install -v ./...
CMD ["go", "test", "-v", "./..."]