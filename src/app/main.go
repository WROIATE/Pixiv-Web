package main

import (
	"Pixiv/src/server"
)

func main() {
	s := server.New()
	s.InitServer()
	s.Start()
}
