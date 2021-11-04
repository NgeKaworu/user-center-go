FROM golang:latest

EXPOSE 80

RUN mkdir -p /home/go/app/bin

WORKDIR /home/go/app

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/app

CMD ["bin/app", '-l=$ADDR', "i=$DBINIT", "m=$MONGO", "db=$DB"," k=$K"]
