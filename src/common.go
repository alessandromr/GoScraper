package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"log"
)

func loadEnv() {
	checkErr(godotenv.Load())
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func mapToSlice(m map[string]int) []string {
	var slice []string
	for k := range m {
		slice = append(slice, k)
	}
	return slice
}
