package main

import (
	"fmt"
	"net/http"
	"log"
	"io"
	"encoding/json"
	"github.com/sacOO7/gowebsocket"
	"github.com/gomodule/redigo/redis"
	"github.com/cvhariharan/Atlan-Task/pkg/utils"
	"github.com/cvhariharan/Atlan-Task/pkg/task"
)

var conn redis.Conn
var socket gowebsocket.Socket

const (
	KILL_ACTION = task.KILL_ACTION
	PAUSE_ACTION = task.PAUSE_ACTION
	RESUME_ACTION = task.RESUME_ACTION
)

func upload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)

	file, fh, err := r.FormFile("image")
	if err != nil {
		fmt.Println(err)
	}
	socket = gowebsocket.New("ws://localhost:9000/uploads")

	socket.OnConnectError = func(err error, socket gowebsocket.Socket) {
		log.Fatal("Received connect error - ", err)
	}

	socket.OnConnected = func(socket gowebsocket.Socket) {
		log.Println("Connected to server")
	}

	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		log.Println("Received message - " + message)
	}

	socket.OnPingReceived = func(data string, socket gowebsocket.Socket) {
		log.Println("Received ping - " + data)
	}

	socket.OnPongReceived = func(data string, socket gowebsocket.Socket) {
		log.Println("Received pong - " + data)
	}

	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		log.Println("Disconnected from server ")
	}

	socket.Connect()

	u := utils.UploadInfo{fh.Filename, "TestUser"}
	msg, _ := json.Marshal(u)
	socket.SendBinary(msg)
	
	b := make([]byte, 4096)
	for {
		_, err := file.Read(b)
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		socket.SendBinary(b)
	}
	socket.SendText("done")
	w.Write([]byte("Created upload task"))
}

func tasksGateway(w http.ResponseWriter, r *http.Request) {
	socket = gowebsocket.New("ws://localhost:9000/tasks")
	socket.Connect()
	username := "TestUser"
	pid := r.URL.Query().Get("pid")
	action := r.URL.Query().Get("action")

	if pid == "" || action == "" {
		w.Write([]byte("Query params must have pid and action"))
		return
	}

	taskserver := utils.GetServer(conn, pid)
	taskId := taskserver+":"+pid
	if taskserver != ""{
		t := utils.TaskAction{ID: pid}
		if utils.IsOwner(conn, taskId, username) {
			switch action {
			case task.KILL_ACTION:
				t.Action = KILL_ACTION
			case task.PAUSE_ACTION:
				t.Action = PAUSE_ACTION
			case task.RESUME_ACTION:
				t.Action = RESUME_ACTION
			default:
				w.Write([]byte("Action not supported"))
				return
			}
		}
		msg, _ := json.Marshal(t)
		socket.SendBinary(msg)
	} else {
		w.Write([]byte("Check pid, pid doesn't exist"))
	}
}

func loop(w http.ResponseWriter, r *http.Request) {
	username := "TestUser"
	socket = gowebsocket.New("ws://localhost:9000/loop")
	socket.Connect()
	socket.SendText(username)
}

func main() {
	Conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		panic(err)
	}
	conn = Conn
	// socket = gowebsocket.New("ws://localhost:9000/loop")
	// socket.Connect()
	// msg, _ := json.Marshal(utils.TaskAction{"784eh", task.KILL_ACTION})
	// socket.SendText("TestUser")
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/tasks", tasksGateway)
	http.HandleFunc("/loop", loop)
	log.Fatal(http.ListenAndServe(":8080", nil))
}