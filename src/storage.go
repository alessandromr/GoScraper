package main

import (
	"database/sql"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/iron-io/iron_go3/config"
	"github.com/iron-io/iron_go3/mq"
	"os"
)

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
