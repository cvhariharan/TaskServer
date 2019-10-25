package main

import (
	"net/http"
	"github.com/gorilla/websocket"
	"log"
	"io"
	"fmt"
	"sync"
	"time"
	"os"
	"encoding/json"
	"github.com/cvhariharan/Atlan-Task/pkg/task"
	"github.com/gomodule/redigo/redis"
	"github.com/cvhariharan/Atlan-Task/pkg/utils"
)


var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const (
	KILL_ACTION = task.KILL_ACTION
	PAUSE_ACTION = task.PAUSE_ACTION
	RESUME_ACTION = task.RESUME_ACTION
)

var conn redis.Conn
var processMap map[string]task.Task

func upload(w http.ResponseWriter, r *http.Request) {
	var wg sync.WaitGroup 
	var uploadInfo utils.UploadInfo
	re, wr := io.Pipe()
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	
	_, fi, err := ws.ReadMessage()
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(fi, &uploadInfo)
	if err != nil {
		log.Println(err)
	}

	// Create a copy process and feed the file inputs into stdin
	c := new(task.CommandTask)
	command := "cp /dev/stdin " + uploadInfo.Filename
	id := c.Init(command)
	fmt.Println(id)
	c.SetInput(re)
	
	// Update the info to redis and store a reference to the task object
	taskInfo := utils.TaskInfo{id, uploadInfo.Username, c.GetStatus(), os.Getenv("TASK_SERVER"), command}
	if utils.InsertTask(conn, taskInfo) {
		processMap[id] = c
		t := utils.Response{id, ""}
		ws.WriteJSON(t)
	}

	// Waitgroup used to prevent the handler from closing before the task 
	// completes
	wg.Add(1)
	go func(wg *sync.WaitGroup){
		err := c.Run()
		if err != nil {
			fmt.Println(err)
		}
		wg.Done()
	}(&wg)

	// Read the chunks from the websocket connection and write to a 
	// pipe connected to stdin
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		wr.Write(message)
	}
	wg.Wait()
}

func loop(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	_, message, err := ws.ReadMessage()
	if err != nil {
		log.Println(err)
	}
	username := string(message)

	c := new(task.CommandTask)
	command := "python3.6 loop.py"
	id := c.Init(command)
	taskInfo := utils.TaskInfo{id, username, c.GetStatus(), os.Getenv("TASK_SERVER"), command}
	if utils.InsertTask(conn, taskInfo) {
		processMap[id] = c
		fmt.Println(id)
		t := utils.Response{id, ""}
		ws.WriteJSON(t)
	}
	go func() {
		err := c.Run()
		if err != nil {
			fmt.Println(err)
		}
	}()

	// Update the status of the task every second
	go func() {
		for {
			time.Sleep(1000 * time.Millisecond)
			utils.UpdateStatus(conn, os.Getenv("TASK_SERVER")+":"+id, c.GetStatus())
		}
	}()
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	var taskAction utils.TaskAction
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	_, message, err := ws.ReadMessage()
	if err != nil {
		log.Println(err)
	}
	err = json.Unmarshal(message, &taskAction)
	if err != nil {
		log.Println(err)
	}

	// Get the task from the map and apply the action
	id := taskAction.ID
	action := taskAction.Action
	c := processMap[id]
	if c != nil {
		switch action {
		case task.KILL_ACTION:
			c.Kill()
		case task.PAUSE_ACTION:
			c.Pause()
		case task.RESUME_ACTION:
			c.Resume()
		}
		t := utils.TaskAction{id, c.GetStatus()}
		ws.WriteJSON(t)
		fmt.Println(c.GetStatus())
	} else {
		t := utils.Response{"", "Task evicted"}
		ws.WriteJSON(t)
	}
}

func main() {
	port := ":9000"
	server := utils.GetLocalIP()+port
	if server == "" {
		server = "127.0.0.1:9000"
	}
	os.Setenv("TASK_SERVER", server)
	dialOps := redis.DialKeepAlive(10*60000 * time.Millisecond)
	Conn, err := redis.Dial("tcp", os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"), dialOps)
	if err != nil {
		panic(err)
	}
	conn = Conn

	processMap = make(map[string]task.Task)

	// Heartbeat every second
	go func() {
		for {
			utils.Heartbeat(conn, os.Getenv("TASK_SERVER"))
			time.Sleep(1000 * time.Millisecond)
		}
	}()
	

	http.HandleFunc("/uploads", upload)
	http.HandleFunc("/loop", loop)
	http.HandleFunc("/tasks", taskHandler)
	log.Fatal(http.ListenAndServe(port, nil))
}