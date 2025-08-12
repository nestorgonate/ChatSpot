package services

import (
	"github.com/gorilla/websocket"
)

type IChatServies interface{
	HandleConnections(c *websocket.Conn)
}

type ChatServices struct {
	repository IChatServies
}

func NewChatServices(repository IChatServies) *ChatServices{
	return &ChatServices{repository: repository}
}

func (r *ChatServices) HandleConnections(c *websocket.Conn){
	r.repository.HandleConnections(c)
}
