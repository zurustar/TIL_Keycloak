package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

var conn *sql.DB

const connStr ="host=127.0.0.1 port=5432 dbname=postgres user=postgres password=postgres sslmode=disable"


func close() {
	log.Println("close")
	conn.Close()	
}

func main() {
	var err error
	conn, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Println(err)
	}
	defer close()
	app := gin.Default()
	app.GET("/", func(c *gin.Context) {
		log.Println("GET /")
		rows, err := conn.Query("SELECT NOW()")
		if err != nil {
			log.Println(err)
			c.JSON(500, gin.H{})
			return
		}
		var now time.Time
		for rows.Next() {
			err = rows.Scan(&now)
			if err != nil {
				log.Println(err)
				c.JSON(500, gin.H{})
				return
			}
			log.Println(now)
		}
		c.JSON(200, gin.H{"now": now})
	})
	app.Run(":5000")
}
