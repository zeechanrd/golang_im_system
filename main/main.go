package main

import "golang_im_system/structFunc"

func main() {
	socketServer := structFunc.NewServer("127.0.0.1", 8088)
	socketServer.Start()
}
