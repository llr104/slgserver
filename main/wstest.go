package main

import (
	"fmt"
	"github.com/forgoer/openssl"
	"github.com/goinggo/mapstructure"
	"github.com/gorilla/websocket"
	"slgserver/net"
	proto2 "slgserver/server/loginserver/proto"
	"slgserver/util"
	"time"
)

var origin = "httpserver://127.0.0.1:8002/"
var secretKey = []byte("")
var session = ""

func main() {

	var dialer *websocket.Dialer
	//通过Dialer连接websocket服务器
	conn, _, err := dialer.Dial("ws://127.0.0.1:8001", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	go do(conn)
	go timeWriter(conn)

	time.Sleep(10*time.Second)
}

func do(conn *websocket.Conn)  {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("%s\n", err)
		}
	}()

	for {
		_, message, _ := conn.ReadMessage()
		msg := &net.RspBody{}
		if len(secretKey) == 0 {
			message, _ = util.UnZip(message)
			if err := util.Unmarshal(message, msg); err == nil{
				if msg.Name == "handshake"{
					h := &net.Handshake{}
					mapstructure.Decode(msg.Msg, h)

					secretKey = []byte(h.Key)
				}
				fmt.Println(msg.Name)
			}
		}else{
			message, _ = util.UnZip(message)
			data, err := util.AesCBCDecrypt(message, secretKey, secretKey, openssl.ZEROS_PADDING)
			if err == nil {
				if err := util.Unmarshal(data, msg); err == nil {
					fmt.Printf("received: %s, code:%d, %v\n", msg.Name, msg.Code, msg.Msg)
					if msg.Name == "login" {
						lr := &proto2.LoginRsp{}
						mapstructure.Decode(msg.Msg, lr)
						session = lr.Session
					}
				}else{
					secretKey = []byte("")
				}
			}else{
				secretKey = []byte("")
			}
		}
	}
}

func login(conn *websocket.Conn)  {
	l := &proto2.LoginReq{Ip: "127.0.0.1", Username: "test", Password: "123456"}
	send(conn, "login", l)
}

func reLogin(conn *websocket.Conn, session string)  {
	l := &proto2.ReLoginReq{Session: session}
	fmt.Println(session)
	send(conn, "reLogin", l)
}

func logout(conn *websocket.Conn)  {
	l := &proto2.LogoutReq{UId: 5}
	send(conn, "logout", l)
}

func send(conn *websocket.Conn, name string, dd interface{})  {
	msg := &net.ReqBody{Name: name, Msg: dd}

	if len(secretKey) == 0 {

	}else{
		if data, err := util.Marshal(msg); err == nil {
			data, _ := util.AesCBCEncrypt(data, secretKey, secretKey, openssl.ZEROS_PADDING)
			data, _ = util.Zip(data)

			conn.WriteMessage(websocket.BinaryMessage, data)
		}
	}
}

func timeWriter(conn *websocket.Conn) {

	time.Sleep(time.Second * 1)
	login(conn)

	time.Sleep(time.Second * 1)
	reLogin(conn, session)

	time.Sleep(20 * time.Second)

	//time.Sleep(time.Second * 1)
	//reLogin(conn, "123")
	//
	//time.Sleep(time.Second * 1)
	//logout(conn)
}
