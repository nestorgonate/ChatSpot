package main

import (
	"ChatSpot/controllers"
	"ChatSpot/repositories"
	"ChatSpot/services"
	"ChatSpot/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	chatRepository := repositories.NewChatRepositories()
	service := services.NewChatServices(chatRepository)
	utils := utils.NewUtils()
	chatController := controllers.NewChatController(service, utils, chatRepository)
	r := gin.Default()
	r.LoadHTMLGlob("views/*.html")
	r.GET("/websockets", chatController.ChatController)
	r.GET("/chat", chatController.ChatView)
	r.Run()
}
