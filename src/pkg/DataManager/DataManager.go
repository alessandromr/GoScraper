package DataManager

import (
	"../LinkManager"
	"../MemQueue"
	"database/sql"
	"github.com/go-redis/redis"
	"github.com/iron-io/iron_go3/mq"
	"log"
)

type DataManager struct {
	RedisClient *redis.Client
	MySqlClient *sql.DB
	QueueClient mq.Queue
}

func (ins DataManager) RegisterURLToMysql(external map[string]int) {
	stmtIns, err := ins.MySqlClient.Prepare("INSERT IGNORE INTO sites (url) VALUES( ? )")
	if err != nil {
		panic(err.Error())
	}
	defer stmtIns.Close()

	for key := range external {
		key = LinkManager.GetCleanUrl(key)
		_, err = stmtIns.Exec(key)
		checkErr(err)
	}

}

func (ins DataManager) RegisterURLToRedis(external []string) {
	for _, element := range external {
		err := ins.RedisClient.SetNX(element, "0", 0).Err()
		checkErr(err)
	}
}

func (ins DataManager) RegisterURLToQueue(external []string) {
	for _, element := range external {
		val, err := ins.RedisClient.Get(element).Result()
		if err != nil {
			panic(err)
		} else if err == redis.Nil {
			continue
		}
		if val == "0" {
			_, err := ins.QueueClient.PushMessage(mq.Message{Delay: 0, Body: element})
			checkErr(err)
		}
	}
	return
}

func (ins DataManager) GetQueueSize() int {
	info, _ := ins.QueueClient.Info()
	return info.Size
}

func (ins DataManager) GetNextURLFromQueue(current string, recents MemQueue.Queue) string {

	//Set Last URL as done on REDIS
	checkErr(ins.RedisClient.Set(current, "1", 0).Err())

	//If there are no recent elements just get one from queue
	if len(recents.URLs) < 1 {
		mess, err := ins.QueueClient.Pop()
		checkErr(err)
		return mess.Body
	}

	for {

		//Peek last n elements from queue
		messages, err := ins.QueueClient.PeekN(recents.Qlen)
		checkErr(err)

		var good string
		bad := false

		//Check if one element in queue is not a recent one
		for _, inQueue := range messages {
			bad = false
			for _, recent := range recents.URLs {
				if LinkManager.GetDomainFromUrl(inQueue.Body) == recent.Domain {
					bad = true
				}
			}
			if !bad {
				good = inQueue.Body
				ins.QueueClient.DeleteMessage(inQueue.Id, "GetNextURLFromQueue")
				break
			}
		}

		//If no element is found return empty
		if good == "" {
			ins.trimQueue(recents.Qlen)
			log.Println("No new URLs in queue, trimming it and trying again")
			continue
		}
		return good
	}
}

func (ins DataManager) RemoveFromMysql(url string) {
	stmtIns, err := ins.MySqlClient.Prepare("DELETE FROM sites WHERE url = ?")
	checkErr(err)
	defer stmtIns.Close()
	_, err = stmtIns.Exec(url)
	checkErr(err)
}

func (ins DataManager) trimQueue(qty int) {
	//Peek last n elements from queue
	messages, err := ins.QueueClient.PeekN(qty)
	checkErr(err)

	var ids []string
	for _, inQueue := range messages {
		ids = append(ids, inQueue.Id)
	}
	ins.QueueClient.DeleteMessages(ids)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
