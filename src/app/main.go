package main

import (
	"Pixiv/src/server"
)

//Bump Version 0.0.5
func main() {
	s := server.New()
	s.InitServer()
	s.Start()
}
