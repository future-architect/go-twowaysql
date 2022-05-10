FROM golang:1.18

WORKDIR /go/src/twowaysql
COPY . .

RUN go install -v ./...
RUN wget -O /usr/bin/wait-for-it https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh \
 && chmod +x /usr/bin/wait-for-it