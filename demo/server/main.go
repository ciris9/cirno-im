package main

import "context"

func main() {
	if err := RunServerStart(context.Background(), &ServerStartOptions{
		id:     "1",
		listen: ":8080",
	}, ""); err != nil {
		panic(err)
	}
}
