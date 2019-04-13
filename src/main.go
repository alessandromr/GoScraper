package main

import (
	"./pkg/DataManager"
	"./pkg/DomManager"
	"./pkg/HttpRequester"
	"./pkg/LinkManager"
	"./pkg/MemQueue"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

func main() {
	loadEnv()
	var next string
	var recent = MemQueue.Queue{Qlen: 80} //Must be below 100

	//Prepare the Database clients
	var dbManager = DataManager.DataManager{
		RedisClient: getRedisClient(),
		MySqlClient: getMysqlClient(),
		QueueClient: getQueueClient(),
	}
	getQueueClient().Clear()
	// getRedisClient().FlushDB()

	wait := getWaitTime()

	go func() {
		for {
			log.Println("Current Queue: ", dbManager.GetQueueSize())
			log.Println("Recent Queue: ", len(recent.URLs))
			time.Sleep(time.Second * 20)
		}
	}()
	for {
		/*Get Url For Next Request*/
		if next == "" {
			next = os.Getenv("START_URL")
		} else if dbManager.GetQueueSize() > 1 {
			next = dbManager.GetNextURLFromQueue(next, recent)
		} else {
			log.Println("NO MORE URL")
			break
		}
		dbManager.FlushQueueIfTooBig()
		printStart(next)

		/*Prepare request*/
		requester := HttpRequester.HttpRequester{
			URL:       next,
			UserAgent: "Url Scraper AC",
			Timeout:   4,
		}
		/*Actual Check*/
		document, err := requester.MakeCheck()
		if checkRequestError(err, dbManager, next, recent) {
			dbManager.SetUrlOnRedis(next)
			continue
		}

		/*Get Link from response document*/
		dom := DomManager.DomManager{
			Document: document,
		}
		urls := dom.GetURLFromDocument()
		extMap, intMap := LinkManager.DivideURLByType(urls, LinkManager.GetDomainFromUrl(next))
		storeResponse(next, extMap, intMap, dbManager, &recent)

		/*Wait 4 seconds before next request*/
		printComplete()
		time.Sleep(wait * time.Second)
	}

}

func getWaitTime() time.Duration {
	wTime, err := strconv.ParseInt(os.Getenv("WAITING_TIME"), 10, 64)
	checkErr(err)
	return time.Duration(wTime)
}

func checkRequestError(err error, dbManager DataManager.DataManager, next string, recent MemQueue.Queue) bool {
	if err != nil {
		log.Printf("\nThis URL has something wrong, better skip\n")
		go func() {
			dbManager.RemoveFromMysql(LinkManager.GetCleanUrl(next))
		}()

		/*Add done URL to last URL Memory Queue*/
		recent.URLs = recent.Push(MemQueue.Url{
			Domain: LinkManager.GetDomainFromUrl(next),
			Clean:  LinkManager.GetCleanUrl(next),
		})
		return true
	}
	return false
}

func storeResponse(current string, extMap map[string]int, intMap map[string]int, dbManager DataManager.DataManager, recent *MemQueue.Queue) {
	var wg sync.WaitGroup

	/*Convert Link From Map To Slice (Array)*/
	var external []string
	var internal []string

	wg.Add(2)
	go func() {
		external = mapToSlice(extMap) //Get a slice from a map
		wg.Done()
	}()
	go func() {
		internal = mapToSlice(intMap) //Get a slice from a map
		wg.Done()
	}()
	_ = internal //Remove Declared but not used error

	go func() {
		//Insert DB to mysql
		dbManager.RegisterURLToMysql(LinkManager.GetCleanedURLMap(extMap))
	}()

	wg.Wait() //Wait mapToSlice to be completed

	wg.Add(3)
	go func() {
		/*Save to redis*/
		dbManager.RegisterURLToRedis(external)
		wg.Done()
	}()
	go func() {
		/*Save in Queue*/
		dbManager.RegisterURLToQueue(external)
		wg.Done()
	}()
	go func() {
		/*Add done URL to last URL Memory Queue*/
		recent.URLs = recent.Push(MemQueue.Url{
			Domain: LinkManager.GetDomainFromUrl(current),
			Clean:  LinkManager.GetCleanUrl(current),
		})
		wg.Done()
	}()
	wg.Wait()
}
