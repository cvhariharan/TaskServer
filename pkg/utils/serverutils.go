package utils

import (
	"log"
	"net"
	"github.com/gomodule/redigo/redis"
)

// GetLocalIP returns the non loopback local IP of the host
func GetLocalIP() string {
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        return ""
    }
    for _, address := range addrs {
        // check the address type and if it is not a loopback the display it
        if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
                return ipnet.IP.String()
            }
        }
    }
    return ""
}

// Returns a random server for the task to be forwarded to
// if no server, returns empty string
func GetRandomServer(conn redis.Conn) string {
    result, err := redis.String(conn.Do("SPOP", "heartbeat"))
    if err != nil {
        log.Println(err)
        return ""
    }
    return result
}

// Returns the taskserver for a given task id
// else returns empty string
func GetServer(conn redis.Conn, pid string) string {
	result, err := redis.String(conn.Do("GET", pid))
	if err != nil {
		log.Println(err)
		return ""
	}
	return result
}


func Heartbeat(conn redis.Conn, taskserver string) bool {
	_, err := conn.Do("SADD", "heartbeat", taskserver)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}