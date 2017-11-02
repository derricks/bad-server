FROM golang

RUN  apt-get update && apt-get install -y curl

ADD . /go/src/bad-server

CMD go run /go/src/bad-server/main.go
 
