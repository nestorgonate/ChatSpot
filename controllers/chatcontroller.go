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
	ChatServices     *services.ChatServices
	GormServices     *services.GormServices
}

func NewChatController(services *services.ChatServices, utils *utils.Utils, chatrepository *repositories.ChatRepositories, gormServices *services.GormServices) *ChatController {
	return &ChatController{ChatServices: services, Utils: utils, ChatRepositories: chatrepository, GormServices: gormServices}
}

// Controla la conexion a websockets
func (r *ChatController) ChatController(c *gin.Context) {
	salaID := c.Query("salaID")
	log.Printf("salaID: %v", salaID)
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
	//Obtener usuario
	go r.ChatRepositories.HandleConnections(conn, salaID)
}

// Controla la vista del chat
func (r *ChatController) ChatView(c *gin.Context) {
	usuarioID := r.Utils.UsuarioIDJWT(c, "usuarioJWT", "usuarioID")
	usuario, _ := r.GormServices.GetUserByID(usuarioID)
	salaIDString := c.Param("id")
	salaID := r.Utils.StringToUint(salaIDString)
	sala, _ := r.GormServices.GetSalaByID(salaID)
	mensajes, _ := r.GormServices.GetLastMessages(salaID)
	log.Printf("Mensajes: %v", mensajes)
	c.HTML(http.StatusOK, "chat.html", gin.H{
		"UsuarioID": usuarioID,
		"Usuario": usuario,
		"Sala":      sala,
		"Mensajes":  mensajes,
	})
}
