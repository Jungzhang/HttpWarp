package main

import (
	"log"
	"flag"
	"os"
	"strconv"
	"encoding/base64"
	"net/http"
	"fmt"
	"net"
	"sync"
	"github.com/gorilla/websocket"
)

var (
	port    int
	path    string
)

var applicationConnMap sync.Map

func initAll() {

	// init command line args
	flag.IntVar(&port, "p", 80, "local proxy port")
	flag.StringVar(&path, "u", "/data/put", "url path of post data")

	// init log
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func handleConn(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if errMsg := recover(); errMsg != nil {
			log.Println("Recovered in handleConn, errMsg is ", errMsg)
		}
	}()

	ws := websocket.Upgrader{}
	wsSrv, err := ws.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade websocket failed:", err)
		return
	}
	defer wsSrv.Close()

	done := make(chan string)
	// 处理websocket客户端发过来的数据
	go handlerWsConn(done, wsSrv)

	for ; ; {
		select {
		case doneMsg := <-done:
			// 关闭与应用服务端的连接
			conn, ok := applicationConnMap.Load(wsSrv)
			if ok {
				conn.(net.Conn).Close()
				applicationConnMap.Delete(wsSrv)
			}
			log.Println("[info] ", doneMsg)
		}
	}
}

// 处理websocket客户端发过来的数据
func handlerWsConn(done chan string, wsCliConn *websocket.Conn) {
	defer func() {
		if errMsg := recover(); errMsg != nil {
			log.Println("Recovered in handleConn, errMsg is ", errMsg)
		}
	}()
	defer func() {
		done <- fmt.Sprintf("Wsclient %s be closed", wsCliConn.RemoteAddr().String())
	}()

	for ; ; {
		// 从ws客户端中读取数据
		wsData := make(map[string]string)
		err := wsCliConn.ReadJSON(wsData)
		if err != nil {
			log.Println("[error] read WSClient failed:", err.Error())
			return
		}

		// 解析应用的数据
		appData, err := base64.StdEncoding.DecodeString(wsData["data"])
		if err != nil {
			log.Println("decode application client data failed:", err.Error())
		}

		// 将数据发送给应用服务端
		if appSrvConn, ok := applicationConnMap.Load(wsCliConn); ok {
			// 如果已经存在连接则直接发送
			if ok {
				_, err := appSrvConn.(net.Conn).Write(appData)
				// 发送成功则返回, 发送失败则重试
				if err == nil {
					continue
				}
			}
		}

		// 建立连接
		appSrvConn, err := connectAndSend(wsData["app_srv_ip"], wsData["app_srv_port"], wsCliConn)
		if err != nil {
			log.Println("[error] connectAndSend:", err.Error())
			return
		}
		go processAppSrvWrite(done, wsCliConn, appSrvConn)

		// send data to application server
		_, err = appSrvConn.Write([]byte(appData))
		if err != nil {
			log.Println("[error] send data to application server failed: ", err.Error())
		}
	}
}

// connect and send data to application server
func connectAndSend(ip, port string, wsCliConn *websocket.Conn) (net.Conn, error) {

	// 幂等操作, 如果之前连接存在过则无脑关闭一次
	if srvConn, ok := applicationConnMap.Load(wsCliConn); ok {
		srvConn.(net.Conn).Close()
	}

	// 连接应用的服务端
	appSrvConn, err := net.Dial("tcp", ip+":"+port)
	if err != nil {
		return nil, fmt.Errorf("connect application server failed, %s", err.Error())
	}

	// 保存连接进列表
	applicationConnMap.Store(wsCliConn, appSrvConn)

	return appSrvConn, nil
}

// 处理应用服务端发来的数据
func processAppSrvWrite(done chan string, wsCliConn *websocket.Conn, appSrvConn net.Conn) {
	defer func() {
		if errMsg := recover(); errMsg != nil {
			log.Println("Recovered in handleConn, errMsg is ", errMsg)
		}
	}()
	defer appSrvConn.Close()

	// 获取该ws client对应的应用服务端连接
	for ; ; {
		// 从应用服务端读取数据
		appSrvData := make([]byte, 0)
		_, err := appSrvConn.Read(appSrvData)
		if err != nil {
			log.Println("[error] read application server failed:", err.Error())
			return
		}
		resp := base64.StdEncoding.EncodeToString(appSrvData)
		// 发送给ws client
		if err := wsCliConn.WriteJSON(resp); err != nil {
			log.Println("[error] send application server data to WsClient failed:", err.Error())
			done <- fmt.Sprintf("Wsclient %s be closed", wsCliConn.RemoteAddr().String())
			return
		}
	}
}

func main() {

	initAll()
	flag.Parse()
	if port == 0 {
		flag.Usage()
		os.Exit(1)
	}

	http.HandleFunc(path, handleConn)
	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		log.Fatalf("[fatalf] start http server failed, " + err.Error())
	}
}
