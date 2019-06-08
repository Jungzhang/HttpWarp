package main

import (
	"os"
	"github.com/Jungzhang/HttpWarp/servers"
	"github.com/Jungzhang/HttpWarp/clients"
)

// 如果可执行文件名为server则执行server端进程
// 否则执行client
func main() {
	if os.Args[0] == "server" {
		servers.Start()
	} else {
		clients.Start()
	}
}
