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
	w.Header().Set("Content-Type", "application/json")
	username := "TestUser"
	endpoint := "/uploads"
	taskserver := "ws://"+utils.GetRandomServer(conn)+endpoint
	r.ParseMultipartForm(32 << 20)

	file, fh, err := r.FormFile("image")
	if err != nil {
		panic(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	socket = gowebsocket.New(taskserver)

	// Maybe choose another taskserver
	socket.OnConnectError = func(err error, socket gowebsocket.Socket) {
		log.Fatal("Received connect error - ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	socket.OnConnected = func(socket gowebsocket.Socket) {
		log.Println("Connected to server")
	}

	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		w.Write([]byte(message))
	}

	// Maybe choose another taskserver
	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		log.Println("Disconnected from server ")
	}

	socket.Connect()
	u := utils.UploadInfo{fh.Filename, username}
	msg, _ := json.Marshal(u)
	socket.SendBinary(msg)
	
	b := make([]byte, 4096)
	for {
		_, err := file.Read(b)
		if err != nil {
			if err == io.EOF {
				break
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		socket.SendBinary(b)
	}
	socket.SendText("done")
	w.Write([]byte("Created upload task"))
}

func tasksGateway(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var resp utils.Response
	endpoint := "/tasks"
	username := "TestUser"
	pid := r.URL.Query().Get("pid")
	action := r.URL.Query().Get("action")

	if pid == "" || action == "" {
		resp.Err = "Query params must have pid and action"
		json, _ := json.Marshal(resp)
		w.Write(json)
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
				resp.Err = "Action not supported"
				json, _ := json.Marshal(resp)
				w.Write(json)
				return
			}
		}
		taskserver = "ws://"+taskserver+endpoint
		fmt.Println(taskserver)
		socket = gowebsocket.New(taskserver)
		socket.Connect()

		socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
			w.Write([]byte(message))
		}

		msg, _ := json.Marshal(t)
		socket.SendBinary(msg)
	} else {
		resp.Err = "pid doesn't exist"
		json, _ := json.Marshal(resp)
		w.Write(json)
	}
	
}

func loop(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	endpoint := "/loop"
	taskserver := "ws://"+utils.GetRandomServer(conn)+endpoint
	username := "TestUser"
	socket = gowebsocket.New(taskserver)

	// Maybe choose another taskserver
	socket.OnConnectError = func(err error, socket gowebsocket.Socket) {
		log.Fatal("Received connect error - ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		w.Write([]byte(message))
	}

	// Maybe choose another taskserver
	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		log.Println("Disconnected from server ")
	}

	socket.Connect()
	socket.SendText(username)
}

func main() {
	Conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		panic(err)
	}
	conn = Conn
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/tasks", tasksGateway)
	http.HandleFunc("/loop", loop)
	log.Fatal(http.ListenAndServe(":8080", nil))
}