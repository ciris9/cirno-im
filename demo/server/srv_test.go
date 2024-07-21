package serv

import (
	"context"
	"testing"
)

func TestServer(t *testing.T) {
	if err := RunServerStart(context.Background(), &ServerStartOptions{
		id:     "1",
		listen: ":8080",
	}, ""); err != nil {
		panic(err)
	}
}
