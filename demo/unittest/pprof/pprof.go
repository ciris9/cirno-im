package main

import (
	_ "net/http/pprof"
)

import (
	"cirno-im/logger"
	"net/http"
)

func main() {

	logger.Println(http.ListenAndServe(":6060", nil))
}
