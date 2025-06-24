package main

import (
	"github.com/devlup-labs/Libr/core/client/handler"
	"github.com/devlup-labs/Libr/core/client/keycache"
	util "github.com/devlup-labs/Libr/core/client/utils"
)

func main() {
	keycache.InitKeys()
	util.InitMockDB()
	handler.RunInputLoop()
}
