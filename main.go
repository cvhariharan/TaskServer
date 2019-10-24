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
	var resp utils.Response
	jwt := r.Header.Get("Authorization")
	if jwt == "" {
		resp.Err = "Authorization header must be set with a valid token"
		json, _ := json.Marshal(resp)
		w.Write(json)
		return
	}
	endpoint := "/uploads"
	taskserver := "ws://"+utils.GetRandomServer(conn)+endpoint

	username := utils.ValidateJWT(jwt)
	if username != "" {
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
	} else {
		resp.Err = "Invalid jwt token"
		json, _ := json.Marshal(resp)
		w.Write(json)
		return
	}
	
}

func tasksGateway(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var resp utils.Response
	endpoint := "/tasks"
	jwt := r.Header.Get("Authorization")
	if jwt == "" {
		resp.Err = "Authorization header must be set with a valid token"
		json, _ := json.Marshal(resp)
		w.Write(json)
		return
	}

	username := utils.ValidateJWT(jwt)
	if username != "" {
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
	} else {
		resp.Err = "Invalid jwt token"
		json, _ := json.Marshal(resp)
		w.Write(json)
		return
	}
}

func loop(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var resp utils.Response
	endpoint := "/loop"
	taskserver := "ws://"+utils.GetRandomServer(conn)+endpoint
	jwt := r.Header.Get("Authorization")
	if jwt == "" {
		resp.Err = "Authorization header must be set with a valid token"
		json, _ := json.Marshal(resp)
		w.Write(json)
		return
	}

	username := utils.ValidateJWT(jwt)
	if username != "" {
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
	} else {
		resp.Err = "Invalid jwt token"
		json, _ := json.Marshal(resp)
		w.Write(json)
		return
	}

}
	

func getToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var resp utils.Response
	switch r.Method {
	case "POST":
		if err := r.ParseForm(); err != nil {
            resp.Err = "Form Parsing error"
			json, _ := json.Marshal(resp)
			w.Write(json)
            return
		}
		username := r.FormValue("username")
		if username == "" {
			resp.Err = "username required"
			json, _ := json.Marshal(resp)
			w.Write(json)
		}
		jwt := utils.GenerateJWT(username)
		resp.Response = jwt
		json, _ := json.Marshal(resp)
		w.Write(json)
	default:
		resp.Err = "Only post method supported"
		json, _ := json.Marshal(resp)
		w.Write(json)
	}
}

func main() {
	Conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		panic(err)
	}
	conn = Conn
	// jwt := utils.GenerateJWT("TestUser")
	// fmt.Println(jwt)
	// fmt.Println(utils.ValidateJWT(jwt))
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/tasks", tasksGateway)
	http.HandleFunc("/loop", loop)
	http.HandleFunc("/auth/token", getToken)
	log.Fatal(http.ListenAndServe(":8080", nil))
}