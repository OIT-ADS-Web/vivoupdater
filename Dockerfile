FROM golang:1.12.4-alpine

RUN apk update && apk add --no-cache git gcc build-base 

RUN mkdir /app

ADD . /app/

WORKDIR /app/

#RUN glide install
#RUN go build /app/cmd/vivo_indexer/main.go
RUN ./build.sh

RUN adduser -S -D -H -h /app gouser

RUN chmod go+x ./cmd/vivo_indexer/vivo_indexer

USER gouser
