package main

import (
	// "./pkg/RedisClient"
	"./pkg/Requester"
	"fmt"
	"time"
)

func main() {
	start := time.Now()
	// redis := RedisClient.CreateNewClient()
	req := Requester.Requester{
		StartUrl: "https://github.com/angular/angular",
	}
	external, internal := req.Check()
	fmt.Println("Duration:", time.Since(start))

	fmt.Println("External")
	for _, element := range external {
		fmt.Println(element)
	}
	fmt.Println()
	fmt.Println("Internal")
	for _, element := range internal {
		fmt.Println(element)
	}
}
