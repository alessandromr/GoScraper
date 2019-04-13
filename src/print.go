package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
)

func printStart(url string) {
	fmt.Println("Starting request for: ", url)
}
func printComplete() {
	fmt.Printf("Complete.... Waiting %s seconds for next one.\n", os.Getenv("WAITING_TIME"))
	fmt.Println("-------------------------------------------")
	fmt.Println("")
}
