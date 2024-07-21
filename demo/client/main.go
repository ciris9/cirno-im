package main

import "context"

func main() {
	if err := run(context.Background(), &StartOptions{
		address: "ws://127.0.0.1:8080",
		user:    "dajiang",
	}); err != nil {
		panic(err)
	}
}
