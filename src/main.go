package main

import (
	"./pkg/RedisClient"
	"./pkg/Requester"
	"database/sql"
	"fmt"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iron-io/iron_go3/config"
	"github.com/iron-io/iron_go3/mq"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open("mysql", "scraper:root@/scraper")
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	redis := RedisClient.CreateNewClient()
	settings := &config.Settings{
		Token:     os.Getenv("IRON_TOKEN"),
		ProjectId: os.Getenv("IRON_PROJECT_ID"),
		Host:      os.Getenv("IRON_HOST"),
		Scheme:    os.Getenv("IRON_SCHEME"),
		Port:      8080,
	}
	queue := mq.ConfigNew("scraper", settings)
	queue.Clear()
	// redis.FlushDB()

	var next string = "https://reddit.com"
	var cond bool = true
	for cond {
		req := Requester.Requester{
			StartUrl: next,
		}
		//Make Request and Get Links
		external, internal := req.Check()
		print(req)
		registerToRedis(external, redis)        //One
		registerToQueue(external, redis, queue) //Two
		registerToMysql(external, db)           //Three

		next = getNext(req.StartUrl, redis, queue)
		if next == "" {
			if len(internal) > 1 {
				next = internal[0]
			} else {
				cond = false
			}
		}
		time.Sleep(4 * time.Second)
	}
}

func registerToRedis(external []string, client *redis.Client) {
	for _, element := range external {
		err := client.SetNX(element, "0", 0).Err()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func registerToQueue(external []string, client *redis.Client, queue mq.Queue) {
	for _, element := range external {
		val, err := client.Get(element).Result()
		if err != nil {
			panic(err)
		} else if err == redis.Nil {
			continue
		}
		if val == "0" {
			_, err := queue.PushMessage(mq.Message{Delay: 0, Body: element})
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return
}

func registerToMysql(external []string, db *sql.DB) {
	stmtIns, err := db.Prepare("INSERT INTO sites (url) VALUES( ? )")
	if err != nil {
		panic(err.Error())
	}
	defer stmtIns.Close()

	for _, element := range external {
		_, err = stmtIns.Exec(element)
		if err != nil {
			log.Fatal(err)
		}
	}

}

func getNext(current string, client *redis.Client, queue mq.Queue) string {
	err := client.Set(current, "1", 0).Err()
	if err != nil {
		log.Fatal(err)
	}
	message, _ := queue.Pop()
	if message.Body == "" {
		return ""
	}
	return message.Body
}

func print(req Requester.Requester) {
	//Print Data for dev env
	fmt.Println("Request: ", req.StartUrl)
	fmt.Println("Next request")
}
