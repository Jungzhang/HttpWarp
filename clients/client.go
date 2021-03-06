package clients

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
)

// 初始化命令行参数以及log
func initAll() {

	// init command line args
	flag.StringVar(&domain, "d", "", "domain name (must input)")
	flag.IntVar(&appSrvPort, "p", 0, "backend application port (must input)")

	flag.StringVar(&appSrvIp, "i", "127.0.0.1", "backend application ip")
	flag.IntVar(&localPort, "l", 10086, "local proxy port")
	flag.StringVar(&path, "u", "/data/put", "url of post data")

	// init log
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// 连接处理
func handleAppCliConn(c net.Conn) {
	defer func() {
		if errMsg := recover(); errMsg != nil {
			log.Println("Recovered in handleAppCliConn, errMsg is ", errMsg)
		}
	}()
	defer c.Close()

	// 建立websocket连接
	wsUri := fmt.Sprintf("ws://%s%s", domain, path)
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
			log.Println("Recovered in handleAppCliConn, errMsg is ", errMsg)
		}
	}()

	//input := bufio.NewScanner(c)
	for ; ; {
		// 读取客户端数据
		//appData := input.Bytes()
		appData := make([]byte, 4096)
		n, err := c.Read(appData)
		if err != nil {
			log.Println("[error] read data from appClient failed, ", err.Error())
			done <- fmt.Sprintf("app client %s be closed", wsCli.RemoteAddr().String())
			return
		}

		// 构造发送到websocket服务端的数据
		payload := map[string][]byte{
			"data":         appData[0:n],
			"app_srv_ip":   []byte(appSrvIp),
			"app_srv_port": []byte(strconv.Itoa(appSrvPort)),
		}

		// 向websocket服务端发送数据
		err = wsCli.WriteJSON(payload)
		if err != nil {
			log.Println("[error] send data to WSServer failed, ", err.Error())
			done <- fmt.Sprintf("ws server %s be closed", wsCli.RemoteAddr().String())
			return
		}
	}
}

func processWsSrv(done chan string, c net.Conn, wsCli *websocket.Conn) {
	defer func() {
		if errMsg := recover(); errMsg != nil {
			log.Println("Recovered in handleAppCliConn, errMsg is ", errMsg)
		}
	}()

	for ; ; {
		// 从websocket服务端读取数据
		wsData := make(map[string]string)
		err := wsCli.ReadJSON(&wsData)
		if err != nil {
			log.Println("[error] read data from WSServer failed, ", err.Error())
			done <- fmt.Sprintf("ws server %s be closed", wsCli.RemoteAddr().String())
			return
		}
		// 解析应用的真实数据
		ret, err := base64.StdEncoding.DecodeString(wsData["data"])
		if err != nil {
			log.Printf("[error] decode application server data failed, %s\n", err.Error())
			continue
		}
		// 发送给应用客户端
		_, err = c.Write(ret)
		if err != nil {
			log.Println("[error] write data to application client failed, ", err.Error())
			if err == io.ErrUnexpectedEOF || err == io.EOF || err == io.ErrClosedPipe {
				done <- fmt.Sprintf("app client %s be closed", wsCli.RemoteAddr().String())
				return
			}
		}
	}
}

func Start() {

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
		go handleAppCliConn(c)
	}
}
