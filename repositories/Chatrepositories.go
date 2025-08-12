package repositories

import (
	"ChatSpot/models"
	"log"

	"github.com/gorilla/websocket"
)

type IChatRepositories interface{
	HandleConnections(c *websocket.Conn)
}

type ChatRepositories struct {
	Clients map[*websocket.Conn]struct{} //Al usar struct, el mapa solo guarda claves sin datos
}

func NewChatRepositories() *ChatRepositories{
	return &ChatRepositories{Clients: make(map[*websocket.Conn]struct{})}
}

func (r *ChatRepositories) HandleConnections(c *websocket.Conn){
	defer func ()  {
		delete(r.Clients, c)
		log.Print("Cerrando conexion websocket")
		c.Close()
	}()
	r.Clients[c]=struct{}{}
	for{
		var msg models.Message
		err := c.ReadJSON(&msg)
		if err != nil{
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure){
				log.Printf("Cliente desconectado: %v", err)
				return
			}else{
				log.Printf("Error de websocket: %v", err)
			}
			return
		}
		r.broadcast(msg)
	}
}

func (r *ChatRepositories) broadcast(msg models.Message){
	for conn := range r.Clients{
		conn.WriteJSON(msg)
	}
}
