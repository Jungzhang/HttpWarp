package main

import (
	"net"
	"log"
	"io"
	"flag"
	"os"
	"strconv"
	"encoding/base64"
	"fmt"
	"github.com/gorilla/websocket"
)

var (
	appSrvIp   string
	appSrvPort int
	localPort  int
	domain     string
	path       string
	timeout    int
)

// 初始化命令行参数以及log
func initAll() {

	// init command line args
	flag.StringVar(&appSrvIp, "i", "0.0.0.0", "backend application ip")
	flag.IntVar(&appSrvPort, "P", 0, "backend application port (must input)")
	flag.IntVar(&localPort, "p", 10086, "local proxy port")
	flag.StringVar(&domain, "d", "", "domain name (must input)")
	flag.StringVar(&path, "u", "/data/put", "url path of post data")
	flag.IntVar(&timeout, "t", 10, "connect domain timeout. units are seconds")

	// init log
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// 连接处理
func handleConn(c net.Conn) {
	defer func() {
		if errMsg := recover(); errMsg != nil {
			log.Println("Recovered in handleConn, errMsg is ", errMsg)
		}
	}()
	defer c.Close()

	// 建立websocket连接
	wsUri := fmt.Sprintf("ws://%s:%s%s", domain, appSrvIp, path)
	wsCli, _, err := websocket.DefaultDialer.Dial(wsUri, nil)
	if err != nil {
		log.Println("[error] dial websocket err:", err)
		return
	}
	defer wsCli.Close()

	done := make(chan string)
	// 处理对应app客户端发过来的数据
	go processAppCli(done, c, wsCli)
	// 处理对应ws服务端发过来的数据
	go processWsSrv(done, c, wsCli)

	// 等待 read handler 或者 write handler 将其置为done
	select {
	case doneMsg := <-done:
		log.Println("[info] ", doneMsg)
		return
	}
}

// 处理app客户端发来的数据
func processAppCli(done chan string, c net.Conn, wsCli *websocket.Conn) {
	defer func() {
		if errMsg := recover(); errMsg != nil {
			log.Println("Recovered in handleConn, errMsg is ", errMsg)
		}
	}()
	defer func() {
		done <- fmt.Sprintf("app client %s be closed", c.RemoteAddr().String())
	}()

	for ; ; {
		// 读取客户端数据
		appData := make([]byte, 0)
		_, err := c.Read(appData)
		if err != nil {
			if err != io.EOF {
				log.Println("[error] read client filed, err:", err.Error())
			} else {
				log.Println("[info] client is closed: ", err.Error())
			}
			return
		}

		// 构造发送到websocket服务端的数据
		appDataEncoded := base64.StdEncoding.EncodeToString(appData)
		payload := map[string]string{
			"app_data":     appDataEncoded,
			"app_srv_ip":   appSrvIp,
			"app_srv_port": strconv.Itoa(appSrvPort),
		}

		// 向websocket服务端发送数据
		err = wsCli.WriteJSON(payload)
		if err != nil {
			log.Println("[error] send data to WSServer failed, ", err.Error())
			return
		}
	}
}

func processWsSrv(done chan string, c net.Conn, wsCli *websocket.Conn) {
	defer func() {
		if errMsg := recover(); errMsg != nil {
			log.Println("Recovered in handleConn, errMsg is ", errMsg)
		}
	}()
	defer func() {
		done <- fmt.Sprintf("app client %s be closed", c.RemoteAddr().String())
	}()

	for ; ; {
		// 从websocket服务端读取数据
		wsData := make(map[string]string)
		err := wsCli.ReadJSON(&wsData)
		if err != nil {
			log.Println("[error] read data from WSServer failed, ", err.Error())
			return
		}
		// 解析应用的真实数据
		ret, err := base64.StdEncoding.DecodeString(wsData["data"])
		if err != nil {
			log.Printf("[error] decode application server data failed, %s\n", err.Error())
			return
		}
		// 发送给应用客户端
		_, err = c.Write(ret)
		if err != nil {
			log.Println("[error] write data to application client failed, ", err.Error())
		}
	}
}

func main() {

	initAll()
	flag.Parse()
	if domain == "" || appSrvPort == 0 {
		flag.Usage()
		os.Exit(1)
	}

	// 启动本地tcp端口
	l, err := net.Listen("tcp", ":"+strconv.Itoa(localPort))
	if err != nil {
		log.Println("listen error:", err)
		return
	}

	for {
		c, err := l.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		}
		go handleConn(c)
	}
}
