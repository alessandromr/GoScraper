package main

import (
	"./pkg/DbManager"
	"./pkg/DomManager"
	"./pkg/HttpRequester"
	"./pkg/LinkManager"
	"./pkg/URLsQueue"
	"database/sql"
	"fmt"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iron-io/iron_go3/config"
	"github.com/iron-io/iron_go3/mq"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	loadEnv()
	next := os.Getenv("START_URL")
	var recent = URLsQueue.Queue{Qlen: 80} //Must be below 100

	//Prepare tge dbManager clients
	var dbManager = DbManager.DbManager{
		RedisClient: getRedisClient(),
		MySqlClient: getMysqlClient(),
		QueueClient: getQueueClient(),
	}
	// getQueueClient().Clear()
	// getRedisClient().FlushDB()

	cond := true
	app, err := strconv.ParseInt(os.Getenv("WAITING_TIME"), 10, 64)
	checkErr(err)
	wait := time.Duration(app)
	if dbManager.GetQueueSize() > 1 {
		next = dbManager.GetNextURLFromQueue(next, recent) //Right Now "next" is the current url and we pass it to GetNext to mark it as used
	}
	go func() {
		for {
			log.Println("Current Queue: ", dbManager.GetQueueSize())
			log.Println("Recent Queue: ", len(recent.URLs))
			time.Sleep(time.Second * 15)
		}
	}()
	for cond {
		/*Preparation*/
		printStart(next)
		requester := HttpRequester.HttpRequester{
			URL:       next,
			UserAgent: "Url Scraper AC",
			Timeout:   30,
		}
		/*Actual Check*/
		document, err := requester.MakeCheck()

		//If request has problem remove URL from Mysql
		if err != nil {
			log.Println("This URL has something wrong, better skip")
			//Remove URL from MySQL
			go func() {
				dbManager.RemoveFromMysql(LinkManager.GetCleanUrl(next))
			}()

			/*Get Url For Next Request*/
			if dbManager.GetQueueSize() > 1 {
				next = dbManager.GetNextURLFromQueue(next, recent)
			} else {
				cond = false
				log.Fatal("No More URL")
			}
			/*Add done URL to last URL Memory Queue*/
			recent.URLs = recent.Push(URLsQueue.Url{
				Domain: LinkManager.GetDomainFromUrl(next),
				Clean:  LinkManager.GetCleanUrl(next),
			})
			continue
		}
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
		go func() {
			dbManager.RegisterURLToMysql(LinkManager.GetCleanedURLMap(extMap))
		}()

		/*Save to redis and save in Queue*/
		dbManager.RegisterURLToRedis(external)
		dbManager.RegisterURLToQueue(external)

		/*Add done URL to last URL Memory Queue*/
		recent.URLs = recent.Push(URLsQueue.Url{
			Domain: LinkManager.GetDomainFromUrl(next),
			Clean:  LinkManager.GetCleanUrl(next),
		})

		/*Get Url For Next Request*/
		if dbManager.GetQueueSize() > 1 {
			next = dbManager.GetNextURLFromQueue(next, recent)
		} else {
			cond = false
			log.Fatal("No More URL")
		}

		/*Wait 4 seconds before next request*/
		printComplete()
		time.Sleep(wait * time.Second)
	}

}

func printStart(url string) {
	fmt.Println("Request Started For: ", url)
}
func printComplete() {
	fmt.Printf("Complete.... Waiting %s seconds for next one.\n", os.Getenv("WAITING_TIME"))
	fmt.Println("-------------------------------------------")
	fmt.Println("")
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
