package main

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const connectionStr = `
	host=127.0.0.1
	port=5432
	dbname=postgres
	user=postgres
	password=postgres
	sslmode=disable`

func connectDB() (*sql.DB, error) {
	conn, err := sql.Open("postgres", connectionStr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

type GenericResponse struct {
	Msg string `json:"message"`
}

type Tag struct {
	TagID     int    `json:"tag_id"`
	TagTypeID int    `json:"tag_type_id"`
	Name      string `json:"name"`
}

type TagsResponse struct {
	Tags []Tag `json:"tags"`
}

func main() {

	log.SetFlags(log.Lshortfile)
	engine := gin.Default()

	apiGrp := engine.Group("/")

	//
	// なくてもいいけどなんか返す
	//
	apiGrp.GET("/", func(c *gin.Context) {
		c.JSON(200, GenericResponse{"I'm Dummy API Server."})
	})

	//
	// すべてのタグ情報を取得する
	//
	apiGrp.GET("/tags", func(c *gin.Context) {
		conn, err := connectDB()
		if err != nil {
			log.Println(err)
			c.JSON(500, GenericResponse{"internal error"})
			return
		}
		defer conn.Close()
		rows, err := conn.Query(`
			SELECT
				tag_id
				, tag_type_id
				, name
			FROM Tags
			ORDER BY tag_id`)
		if err != nil {
			log.Println(err)
			c.JSON(500, GenericResponse{"internal error"})
			return
		}
		result := TagsResponse{}
		for rows.Next() {
			var t Tag
			err = rows.Scan(&t.TagID, &t.TagTypeID, &t.Name)
			if err != nil {
				log.Println(err)
				c.JSON(500, GenericResponse{"internal error"})
				return
			}
			result.Tags = append(result.Tags, t)
		}
		c.JSON(200, result)
	})

	//
	// 特定のタグの情報を取得する
	//
	apiGrp.GET("/tags/:tag_id", func(c *gin.Context) {
		tagIDStr := c.Param("tag_id")
		tagID, err := strconv.Atoi(tagIDStr)
		if err != nil {
			log.Println(err)
			c.JSON(500, GenericResponse{"internal error"})
			return
		}
		conn, err := connectDB()
		if err != nil {
			log.Println(err)
			c.JSON(500, GenericResponse{"internal error"})
			return
		}
		defer conn.Close()
		var t Tag
		err = conn.QueryRow(`
			SELECT
				tag_id
				, tag_type_id
				, name
			FROM Tags
			WHERE tag_id = $1`,
			tagID).Scan(&t.TagID, &t.TagTypeID, &t.Name)
			if err != nil {
				log.Println(err)
				c.JSON(500, GenericResponse{"internal error"})
				return
			}
		c.JSON(200, t)
	})

	//
	//
	//
	engine.Run(":3000")
}
