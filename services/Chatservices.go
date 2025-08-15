package services

import (
	"github.com/gorilla/websocket"
)

type IChatServies interface{
	HandleConnections(conn *websocket.Conn, salaID string, usuarioUsuario string)
}

type ChatServices struct {
	repository IChatServies
}

func NewChatServices(repository IChatServies) *ChatServices{
	return &ChatServices{repository: repository}
}

func (r *ChatServices) HandleConnections(c *websocket.Conn, salaID string, usuarioUsuario string){
	r.repository.HandleConnections(c, salaID, usuarioUsuario)
}
