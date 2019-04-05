package DbManager

import (
	"../LinkManager"
	"database/sql"
	"github.com/go-redis/redis"
	"github.com/iron-io/iron_go3/mq"
	"log"
)

type DbManager struct {
	RedisClient *redis.Client
	MySqlClient *sql.DB
	QueueClient mq.Queue
}

func (ins DbManager) RegisterURLToMysql(external map[string]int) {
	stmtIns, err := ins.MySqlClient.Prepare("INSERT IGNORE INTO sites (url) VALUES( ? )")
	if err != nil {
		panic(err.Error())
	}
	defer stmtIns.Close()

	for key := range external {
		key = LinkManager.GetCleanUrl(key)
		_, err = stmtIns.Exec(key)
		if err != nil {
			log.Fatal(err)
		}
	}

}

func (ins DbManager) RegisterURLToRedis(external []string) {
	for _, element := range external {
		err := ins.RedisClient.SetNX(element, "0", 0).Err()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (ins DbManager) RegisterURLToQueue(external []string) {
	for _, element := range external {
		val, err := ins.RedisClient.Get(element).Result()
		if err != nil {
			panic(err)
		} else if err == redis.Nil {
			continue
		}
		if val == "0" {
			_, err := ins.QueueClient.PushMessage(mq.Message{Delay: 0, Body: element})
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return
}

func (ins DbManager) registerWorkedToMysql(current string, db *sql.DB) {
	stmtIns, err := db.Prepare("INSERT IGNORE INTO worked_sites (url) VALUES( ? )")
	if err != nil {
		panic(err.Error())
	}
	defer stmtIns.Close()
	_, err = stmtIns.Exec(current)
	if err != nil {
		log.Fatal(err)
	}
}

func (ins DbManager) GetQueueSize() int {
	info, _ := ins.QueueClient.Info()
	return info.Size
}

func (ins DbManager) GetNextURLFromQueue(current string) string {
	err := ins.RedisClient.Set(current, "1", 0).Err()
	if err != nil {
		log.Fatal(err)
	}
	message, _ := ins.QueueClient.Pop()
	if message.Body == "" {
		return ""
	}
	return message.Body
}
