package main

import (
	"os"
	"github.com/Jungzhang/HttpWarp/servers"
	"github.com/Jungzhang/HttpWarp/clients"
	"runtime"
	"strings"
)

// 如果可执行文件名为server则执行server端进程
// 否则执行client
func main() {

	split := "/"
	if runtime.GOOS == "windows" {
		split = "\\"
	}
	path := strings.Split(os.Args[0], split)

	if path[len(path)-1] == "server" || path[len(path)-1] == "./server" {
		servers.Start()
	} else {
		clients.Start()
	}
}
