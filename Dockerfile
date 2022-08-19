FROM golang:1.19

WORKDIR /go/src/twowaysql
COPY . .

RUN go install -v ./...
CMD ["go", "test", "-v", "./..."]