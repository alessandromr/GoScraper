package main

import (
	"./pkg/RedisClient"
	"./pkg/Requester"
	"github.com/joho/godotenv"
	// "database/sql"
	"fmt"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iron-io/iron_go3/config"
	"github.com/iron-io/iron_go3/mq"
	"log"
	"os"
	"time"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	// db, err := sql.Open("mysql", "scraper:root@/scraper")
	redis := RedisClient.CreateNewClient()
	settings := &config.Settings{
		Token:     os.Getenv("IRON_TOKEN"),
		ProjectId: os.Getenv("IRON_PROJECT_ID"),
		Host:      os.Getenv("IRON_HOST"),
		Scheme:    os.Getenv("IRON_SCHEME"),
		Port:      8080,
	}
	queue := mq.ConfigNew("scraper", settings)

	var next string = "https://golangcode.com/sleeping-with-go/"
	var cond bool = true
	for cond {
		req := Requester.Requester{
			StartUrl: next,
		}
		//Make Request and Get Links
		external, _ := req.Check()
		print(req)
		registerToQueue(external, redis, queue) //First
		registerToRedis(external, redis)        //After

		next = getNext(queue)
		time.Sleep(4 * time.Second)
	}
}

func registerToRedis(external []string, client *redis.Client) {
	for _, element := range external {
		err := client.SetNX(element, 0, 0).Err()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func registerToQueue(external []string, client *redis.Client, queue mq.Queue) {
	for _, element := range external {
		_, err := client.Get(element).Result()
		if err == redis.Nil {
			_, err := queue.PushMessage(mq.Message{Delay: 0, Body: element})
			if err != nil {
				log.Fatal(err)
			}
		} else if err != nil {
			panic(err)
		}
	}
	return
}

// func registerToMysql() {

// }

func getNext(queue mq.Queue) string {
	message, _ := queue.PeekN(1)
	return message[0].Body
}

func print(req Requester.Requester) {
	//Print Data for dev env
	fmt.Println("Request: ", req.StartUrl)
	fmt.Println("Next request")
}
