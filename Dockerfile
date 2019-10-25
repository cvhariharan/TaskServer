FROM golang
COPY ./pkg/task /go/src/github.com/cvhariharan/Atlan-Task/pkg/task
COPY ./pkg/utils /go/src/github.com/cvhariharan/Atlan-Task/pkg/utils

WORKDIR /go/src/app
COPY . .

RUN go get github.com/sacOO7/gowebsocket
RUN go get github.com/gomodule/redigo/redis
RUN go get github.com/rs/xid
RUN go install -v ./...
EXPOSE 8080

CMD ["app"]