package main

import (
	"Pixiv/src/server"
)

//Bump Version 0.0.6
func main() {
	s := server.New()
	s.InitServer()
	s.Start()
}
