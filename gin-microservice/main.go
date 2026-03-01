package main

import (
	"log"
	// "net/http"

	"github.com/gin-gonic/gin"
)

func main(){
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	router.GET("/", showIndexPage)
	router.GET("/article/view/:article_id", getArticle)
	log.Print("Success!")
	router.Run()
}