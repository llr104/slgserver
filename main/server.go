package main

import (
	"fmt"
	"os"
	"slgserver/config"
	"slgserver/server/conn"
	"slgserver/server/run"
)


func getServerAddr() string {
	host := config.File.MustValue("server", "host", "")
	port := config.File.MustValue("server", "port", "8001")
	return host + ":" + port
}

func main() {
	fmt.Println(os.Getwd())
	run.Init()
	needSecret := config.File.MustBool("server", "need_secret", false)
	s := conn.NewServer(getServerAddr(), needSecret)
	s.Router(run.MyRouter)
	s.Start()
}


