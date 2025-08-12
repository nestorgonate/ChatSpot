package controllers

import (
	"ChatSpot/repositories"
	"ChatSpot/services"
	"ChatSpot/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ChatController struct {
	Utils            *utils.Utils
	ChatRepositories *repositories.ChatRepositories
	Services         *services.ChatServices
}

func NewChatController(services *services.ChatServices, utils *utils.Utils, chatrepository *repositories.ChatRepositories) *ChatController {
	return &ChatController{Services: services, Utils: utils, ChatRepositories: chatrepository}
}

func (r *ChatController) ChatController(c *gin.Context) {
	allowedOrigin := r.Utils.AllowedOrigins
	//Valida si se debe cambiar el protocolo http a websocket
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		origin := c.GetHeader("Origin")
		for _, allowed := range allowedOrigin {
			if origin == allowed {
				return true
			}
		}
		return false
	},
	}
	//Cambia el protocolo http a websocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("No sirve websockets: %v", err)
		return
	}
	go r.ChatRepositories.HandleConnections(conn)
}

func (r *ChatController) ChatView(c *gin.Context) {
	c.HTML(200, "chat.html", nil)
}
