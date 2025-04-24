package main

import (
	"scira2api/config"
	"scira2api/service"

	"scira2api/log"
	"scira2api/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	config := config.NewConfig()
	handler := service.NewChatHandler(config)
	router := gin.Default()

	router.Use(middleware.AuthMiddleware(config))
	router.Use(middleware.CorsMiddleware())

	router.GET("/v1/models", handler.ModelGetHandler)

	router.POST("/v1/chat/completions", handler.ChatCompletionsHandler)

	log.Info("Server is running on port %s", config.Port)

	router.Run(":" + config.Port)
}
