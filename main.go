package main

import (
	"net/http"
	"log"
	"io"
	"strings"
	"fmt"
	"os"
	"encoding/json"
	"github.com/sacOO7/gowebsocket"
	"github.com/gomodule/redigo/redis"
	"github.com/cvhariharan/Atlan-Task/pkg/utils"
	"github.com/cvhariharan/Atlan-Task/pkg/task"
)

var conn redis.Conn
// var socket gowebsocket.Socket

// Redefining for ease
const (
	KILL_ACTION = task.KILL_ACTION
	PAUSE_ACTION = task.PAUSE_ACTION
	RESUME_ACTION = task.RESUME_ACTION
)

// Utility to handle error responses
func handleError(w http.ResponseWriter, err string) {
	var resp utils.Response
	resp.Err = err
	json, _ := json.Marshal(resp)
	w.Write(json)
}

// Returns a single task as json response
func taskInfo(w http.ResponseWriter, pid, taskserver, username string) {
	t := utils.GetTask(conn, pid, taskserver, username)
	if t.ID == "" {
		handleError(w, "Unauthorized")
		return
	}
	resp, _ := json.Marshal(t)
	w.Write(resp)
}

// Returns all the tasks of a user as a json array
func allTaskInfo(w http.ResponseWriter, username string) {
	var t []utils.TaskInfo
	ids := utils.GetAllTasks(conn, username)
	for _, id := range ids {
		pid := strings.Split(id, ":")[2]
		taskserver := strings.Split(id, ":")[0] + ":" + strings.Split(id, ":")[1]
		fmt.Println(taskserver)
		t = append(t, utils.GetTask(conn, pid, taskserver, username))
	}
	resp, _ := json.Marshal(t)
	w.Write(resp)
}

// To demonstrate an upload process
// POST method
func upload(w http.ResponseWriter, r *http.Request) {
	exit := false
	w.Header().Set("Content-Type", "application/json")
	jwt := r.Header.Get("Authorization")
	if jwt == "" {
		handleError(w, "Authorization header must be set with a valid token")
		return
	}

	// endpoint of the taskserver. Same nomenclature throughout
	endpoint := "/uploads"

	// Get a random taskserver using redis which keeps track of all taskservers
	taskserver := "ws://"+utils.GetRandomServer(conn)+endpoint

	// Authorization header must be set with jwt to get username
	username := utils.ValidateJWT(jwt)
	if username == "" {
		handleError(w, "Invalid JWT")
		return
	}

	r.ParseMultipartForm(32 << 20)

	file, fh, err := r.FormFile("image")
	if err != nil {
		handleError(w, "File not uploaded")
		return
	}
	socket := gowebsocket.New(taskserver)

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
		exit = true
	}

	// Maybe choose another taskserver
	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		log.Println("Disconnected from server ")
	}
	socket.Connect()

	// Send info about the file to be uploaded
	u := utils.UploadInfo{fh.Filename, username}
	msg, _ := json.Marshal(u)
	socket.SendBinary(msg)
	
	// Send the actual file contents as chunks
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
	// Just to signify the end. Not required but kept for testing
	socket.SendText("done")

	for !exit {

	}

}

// Endpoint to allow sending control commands to taskserver
// GET method
func tasksGateway(w http.ResponseWriter, r *http.Request) {
	exit := false
	w.Header().Set("Content-Type", "application/json")
	endpoint := "/tasks"
	jwt := r.Header.Get("Authorization")
	if jwt == "" {
		handleError(w, "Authorization header must be set with a valid token")
		return
	}

	username := utils.ValidateJWT(jwt)
	if username == "" {
		handleError(w, "Invalid JWT")
		return
	}

	pid := r.URL.Query().Get("pid")
	action := r.URL.Query().Get("action")

	server := utils.GetServer(conn, pid)
	taskId := server+":"+pid
	taskserver := "ws://"+server+endpoint

	// If pid not set, get all tasks of the user
	if pid == "" {
		allTaskInfo(w, username)
		return
	}

	socket := gowebsocket.New(taskserver)
	socket.Connect()

	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		w.Write([]byte(message))
		exit = true
	}

	// If no action, info has to be returned. Only send
	// the pid to the corresponding taskserver
	if action == "" {
		taskInfo(w, pid, server, username)
		return
	}

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
				handleError(w, "Action not supported")
				return
			}
		}

		msg, _ := json.Marshal(t)
		socket.SendBinary(msg)
	} else {
		handleError(w, "Given pid doesn't exist")
	}
	// To prevent the handler from exiting before getting the response
	// from the task server
	for !exit {

	}
}

// An example of a long running task
// tells the taskserver to run an infinite loop
// GET method
func loop(w http.ResponseWriter, r *http.Request) {
	exit := false
	w.Header().Set("Content-Type", "application/json")
	endpoint := "/loop"
	taskserver := "ws://"+utils.GetRandomServer(conn)+endpoint
	jwt := r.Header.Get("Authorization")
	if jwt == "" {
		handleError(w, "Authorization header must be set with a valid token")
		return
	}

	username := utils.ValidateJWT(jwt)
	if username == "" {
		handleError(w, "Invalid JWT")
		return
	}

	socket := gowebsocket.New(taskserver)
	socket.Connect()
	// Maybe choose another taskserver
	socket.OnConnectError = func(err error, socket gowebsocket.Socket) {
		log.Fatal("Received connect error - ", err)
		handleError(w, err.Error())
	}

	socket.OnConnected = func(socket gowebsocket.Socket) {
		log.Println("Connected to server")
	}

	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		w.Write([]byte(message))
		exit = true
	}

	// Maybe choose another taskserver
	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		log.Println("Disconnected from server ")
	}
	socket.SendText(username)

	// To prevent the handler from exiting before getting the response
	// from the task server
	for !exit {

	}
}
	
// Returns a jwt for a username
// Just an example and no auth performed
// POST method
func getToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var resp utils.Response
	switch r.Method {
	case "POST":
		if err := r.ParseForm(); err != nil {
            handleError(w, "Form Parsing error")
            return
		}

		username := r.FormValue("username")
		if username == "" {
			handleError(w, "Username required")
			return
		}

		jwt := utils.GenerateJWT(username)
		resp.Response = jwt
		json, _ := json.Marshal(resp)
		w.Write(json)

	default:
		handleError(w, "Only post method supported")
	}
}

func main() {
	Conn, err := redis.Dial("tcp", os.Getenv("REDIS_HOST"))
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