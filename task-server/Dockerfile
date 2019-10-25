FROM ubuntu:18.04

# install golang
# RUN add-apt-repository ppa:ubuntu-lxc/lxd-stable

RUN apt-get update
RUN apt-get upgrade -y
RUN apt-get install golang -y
# RUN apt-get install wget -y
# RUN wget https://storage.googleapis.com/golang/go1.11.2.linux-amd64.tar.gz
# RUN tar -xvf go1.11.2.linux-amd64.tar.gz
# RUN mv go /usr/local

# set path for golang
ENV GOPATH=$HOME/go
# ENV PATH=$PATH:/usr/local/go/bin:$GOPATH/bin

RUN mkdir $HOME/go

RUN apt-get install git -y

COPY ./pkg/task $HOME/go/src/github.com/cvhariharan/Atlan-Task/pkg/task
COPY ./pkg/utils $HOME/go/src/github.com/cvhariharan/Atlan-Task/pkg/utils

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