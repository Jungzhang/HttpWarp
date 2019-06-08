package main

import (
	"os"
	"github.com/Jungzhang/HttpWarp/server"
	"github.com/Jungzhang/HttpWarp/client"
)

// 如果可执行文件名为server则执行server端进程
// 否则执行client
func main() {
	if os.Args[0] == "server" {
		server.Start()
	} else {
		client.Start()
	}
}
