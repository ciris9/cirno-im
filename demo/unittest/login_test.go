package unittest

import (
	"testing"
	"time"

	"cirno-im"
	"cirno-im/demo/dialer"
	"cirno-im/websocket"
	"github.com/stretchr/testify/assert"
)

func login(account string) (cim.Client, error) {
	cli := websocket.NewClient(account, "unittest", nil, websocket.ClientOptions{})

	cli.SetDialer(&dialer.ClientDialer{})
	err := cli.Connect("ws://localhost:8000")
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func Test_login(t *testing.T) {
	cli, err := login("test1")
	assert.Nil(t, err)
	time.Sleep(time.Second * 2)
	cli.Close()
}
