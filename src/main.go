package main

import (
	"./pkg/DbManager"
	"./pkg/DomManager"
	"./pkg/HttpRequester"
	"./pkg/LinkManager"
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
	loadEnv()
	next := "https://www.reddit.com/r/EpicSeven/comments/adbngw/compilation_of_useful_links/"
	var dbManager = DbManager.DbManager{
		RedisClient: getRedisClient(),
		MySqlClient: getMysqlClient(),
		QueueClient: getQueueClient(),
	}
	// getQueueClient().Clear()
	// getRedisClient().FlushDB()

	cond := true
	for cond {
		/*Preparation*/
		log.Println("Current Queue: ", dbManager.GetQueueSize())
		if dbManager.GetQueueSize() > 1 {
			next = dbManager.GetNextURLFromQueue(next) //Right Now "next" is the current url and we pass it to GetNext to mark it as used
		}
		printStart(next)
		requester := HttpRequester.HttpRequester{
			URL:       next,
			UserAgent: "Url Scraper AC",
			Timeout:   30,
		}

		/*Actual Check*/
		document := requester.MakeCheck()

		/*Get Link from response document*/
		dom := DomManager.DomManager{
			Document: document,
		}
		urls := dom.GetURLFromDocument()
		extMap, intMap := LinkManager.DivideURLByType(urls, LinkManager.GetDomainFromUrl(next))
		/*Convert Link From Map To Slice (Array)*/
		external := mapToSlice(extMap)
		internal := mapToSlice(intMap)
		_ = internal //Remove Declared but not used error

		/*Get a cleaned map of URL*/
		dbManager.RegisterURLToMysql(LinkManager.GetCleanedURLMap(extMap))

		/*Save to redis and save in Queue*/
		dbManager.RegisterURLToRedis(external)
		dbManager.RegisterURLToQueue(external)

		/*Get Url For Next Request*/
		if dbManager.GetQueueSize() > 1 {
			next = dbManager.GetNextURLFromQueue(next)
		} else {
			cond = false
			log.Fatal("No More URL")
		}

		/*Wait 4 seconds before next request*/
		printComplete()
		time.Sleep(4 * time.Second)
	}
}

func printStart(url string) {
	fmt.Println("Request Started For: ", url)
}
func printComplete() {
	fmt.Println("Complete.... Waiting 4 seconds for next one.")
}

func loadEnv() {
	checkErr(godotenv.Load())
}

func getRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}
func getMysqlClient() *sql.DB {
	db, err := sql.Open("mysql", "scraper:root@/scraper")
	checkErr(err)
	checkErr(db.Ping())
	return db
}
func getQueueClient() mq.Queue {
	settings := &config.Settings{
		Token:     os.Getenv("IRON_TOKEN"),
		ProjectId: os.Getenv("IRON_PROJECT_ID"),
		Host:      os.Getenv("IRON_HOST"),
		Scheme:    os.Getenv("IRON_SCHEME"),
		Port:      8080,
	}
	return mq.ConfigNew("scraper", settings)
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
