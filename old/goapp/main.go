package main

import "github.com/gin-gonic/gin"


const AUTH_ENDPOINT = "http://192.168.0.200:8080/auth/realms/demo/protocol/openid-connect/auth"

func main() {
	app := gin.Default()

	app.Static("/static", "./static")

	app.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "OK"})
	})
	app.GET("/login", func(c *gin.Context) {
		c.Redirect(302, AUTH_ENDPOINT)
	})
	app.Run(":5000")
}