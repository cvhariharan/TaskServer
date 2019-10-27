FROM golang
COPY ./pkg/task /go/src/github.com/cvhariharan/TaskServer/pkg/task
COPY ./pkg/utils /go/src/github.com/cvhariharan/TaskServer/pkg/utils

WORKDIR /go/src/app
COPY . .

RUN go get github.com/sacOO7/gowebsocket
RUN go get github.com/gomodule/redigo/redis
RUN go get github.com/rs/xid
RUN go install -v ./...
EXPOSE 8000

CMD ["app"]