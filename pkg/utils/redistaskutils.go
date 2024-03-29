package utils

import (
	"log"
	"github.com/gomodule/redigo/redis"
)

type TaskInfo struct {
	ID string
	Owner string
	Status string
	TaskServer string
	Command string
}

// Insert the task into a hashmap with key as taskserver:id and also add 
// taskserver:id to a set corresponding to each username
func InsertTask(conn redis.Conn, t TaskInfo) bool {
	taskId := t.TaskServer+":"+t.ID
	_, err := conn.Do("HMSET", redis.Args{taskId}.AddFlat(t)...)
	if err != nil {
		return false
	}
	_, err = conn.Do("SADD", t.Owner, taskId)
	if err != nil {
		return false
	}
	_, err = conn.Do("SET", t.ID, t.TaskServer)
	if err != nil {
		return false
	}
	return true
}

// Checks if the task owner is the user and if yes, returns the TaskInfo
// else returns nil
func GetTask(conn redis.Conn, id, taskserver, username string) TaskInfo {
	var t TaskInfo
	taskId := taskserver+":"+id
	exists, err := redis.Bool(conn.Do("SISMEMBER", username, taskId))
	if err != nil {
		panic(err)
	}
	if exists {
		value, err := redis.Values(conn.Do("HGETALL", taskId))
		if err != nil {
			panic(err)
		}
		err = redis.ScanStruct(value, &t)
		if err != nil {
			panic(err)
		}
	}
	return t
}

func UpdateStatus(conn redis.Conn, taskId, status string) bool {
	_, err := conn.Do("HSET", taskId, "Status", status)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

// Returns true if the user is the task owner else false
func IsOwner(conn redis.Conn, taskId, username string) bool {
	exists, err := redis.Bool(conn.Do("SISMEMBER", username, taskId))
	if err != nil {
		panic(err)
	}
	return exists
}

// Returns an array of taskserver:id for a given user
func GetAllTasks(conn redis.Conn, username string) []string {
	results, err := redis.Strings(conn.Do("SMEMBERS", username))
	if err != nil {
		panic(err)
	}
	return results
}





