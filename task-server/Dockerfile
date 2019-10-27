FROM ubuntu:18.04

RUN apt-get update
RUN apt-get upgrade -y
RUN apt-get install golang -y

ENV GOPATH=$HOME/go

RUN mkdir $HOME/go

RUN apt-get install git -y

COPY ./pkg/task $HOME/go/src/github.com/cvhariharan/TaskServer/pkg/task
COPY ./pkg/utils $HOME/go/src/github.com/cvhariharan/TaskServer/pkg/utils

RUN go get github.com/sacOO7/gowebsocket
RUN go get github.com/gomodule/redigo/redis
RUN go get github.com/rs/xid
RUN go get github.com/gorilla/websocket

RUN apt install python3.6 -y

WORKDIR $HOME/go/src/app
COPY ./task-server .

RUN go build -o main
EXPOSE 9000

CMD ["./main"]