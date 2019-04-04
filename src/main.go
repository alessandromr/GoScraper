package main

import (
	// "./pkg/RedisClient"
	"./pkg/Requester"
	"fmt"
	"sort"
	"time"
)

func main() {
	start := time.Now()
	// redis := RedisClient.CreateNewClient()
	req := Requester.Requester{
		StartUrl: "https://stackoverflow.com/questions/23330781/sort-go-map-values-by-keys",
	}

	//Make Request and Get Links
	external, internal := req.Check()

	//Sort Array
	sort.Strings(external)
	sort.Strings(internal)

	//Print Data for dev env
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
