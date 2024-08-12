package constants

import "time"

const (
	DefaultReadWait  = time.Minute * 3
	DefaultWriteWait = time.Second * 10
	DefaultLoginWait = time.Second * 10
	DefaultHearBeat  = time.Minute
)

const (
	// 定义读取消息的默认goroutine池大小
	DefaultMessageReadPool = 5000
	DefaultConnectionPool  = 5000
)
