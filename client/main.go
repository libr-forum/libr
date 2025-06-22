package main

import (
	"libr/handler"
	"libr/keycache"
)

func main() {
	keycache.InitKeys()
	handler.RunInputLoop()
}
