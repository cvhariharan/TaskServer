package main

import (
	"net/http"
	"github.com/gorilla/websocket"
	"log"
	"io"
	"fmt"
	"sync"
	"time"
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

	c := new(task.CommandTask)
	id := c.Init("cp /dev/stdin " + uploadInfo.Filename)
	fmt.Println(id)
	c.SetInput(re)
	if utils.InsertTask(conn, id, uploadInfo.Username, c.GetStatus(), "127.0.0.1:9000") {
		processMap[id] = c
	}

	wg.Add(1)
	go func(wg *sync.WaitGroup){
		err := c.Run()
		if err != nil {
			fmt.Println(err)
		}
		wg.Done()
	}(&wg)

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
	id := c.Init("python loop.py")
	fmt.Println(id)
	if utils.InsertTask(conn, id, username, c.GetStatus(), "127.0.0.1:9000") {
		processMap[id] = c
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
			utils.UpdateStatus(conn, "127.0.0.1:9000:"+id, c.GetStatus())
			time.Sleep(1000 * time.Millisecond)
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

	id := taskAction.ID
	action := taskAction.Action
	c := processMap[id]
	switch action {
	case task.KILL_ACTION:
		c.Kill()
	case task.PAUSE_ACTION:
		c.Pause()
	case task.RESUME_ACTION:
		c.Resume()
	}
	fmt.Println(c.GetStatus())
}

func main() {
	Conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		panic(err)
	}
	conn = Conn

	processMap = make(map[string]task.Task)
	// fmt.Println(utils.GetTask(conn, "bmnv3rn20qitifrqfpcg", "127.0.0.1:9000", "Tr"))
	// fmt.Println(utils.GetAllTasks(conn, "TestUser"))
	http.HandleFunc("/uploads", upload)
	http.HandleFunc("/loop", loop)
	http.HandleFunc("/tasks", taskHandler)
	log.Fatal(http.ListenAndServe(":9000", nil))
}