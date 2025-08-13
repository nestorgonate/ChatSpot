package repositories

import (
	"ChatSpot/models"
	"ChatSpot/utils"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/rabbitmq/amqp091-go"
)

type IChatRepositories interface{
	HandleConnections(c *websocket.Conn, salaID string)
}

type ChatRepositories struct {
	Clients map[*websocket.Conn]string //Solo tiene llaves y el struct no tiene valor, map de conexiones websocket por sala
	Utils *utils.Utils
	SalaConsumers map[string]bool //Consumidores por sala
	db *GormRepositories
}

func NewChatRepositories(utils *utils.Utils, db *GormRepositories) *ChatRepositories{
	return &ChatRepositories{
		Clients: make(map[*websocket.Conn]string),
		SalaConsumers: make(map[string]bool),
		Utils: utils,
		db: db,
	}
}

func (r *ChatRepositories) HandleConnections(conn *websocket.Conn, salaID string){
	fmt.Println("HandelConnections")
	defer func ()  {
		delete(r.Clients, conn)
		log.Print("Cerrando conexion websocket")
		conn.Close()
	}()
	//Cada conexion sabe su salaID
	r.Clients[conn] = salaID
	//Validar si el consumidor no existe
	if !r.SalaConsumers[salaID]{
		r.SalaConsumers[salaID] = true
		go r.ConsummerRabbitMQ(salaID)
	}
	for{
		var mensaje models.Message
		err := conn.ReadJSON(&mensaje)
		if err != nil{
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure){
				log.Printf("Cliente desconectado: %v", err)
				break
			}else{
				log.Printf("Error de websocket: %v", err)
			}
			return
		}
		//Publicar mensajes en RabbitMQ
		salaID := r.Utils.UintToString(mensaje.SalaID)
		err = r.Utils.Channel.Publish(
			"chat_exchange", //Exchange
			salaID, //Routing key, debe coincidir con RabbitMQ.ChannelQueBinding
			false, //Mandatory
			false, //Inmediate
			amqp091.Publishing{
				ContentType: "application/json",
				Body: r.messageToJSON(mensaje),
			},
		)
		if err != nil{
			fmt.Printf("No se publico el mensaje a RabbitMQ: %v", err)
		}
	}
}

//Declara exchange, queue, binding, consume y reenvia mensajes
func (r *ChatRepositories) ConsummerRabbitMQ(salaID string) {
	fmt.Println("ConsummerRabbitMQ")
	var mensajes models.Message
	var usuario models.Usuarios
	//Declarar el exchange
	err := r.Utils.Channel.ExchangeDeclare(
		"chat_exchange",
		"direct",
		true,     // durable
        false,    // auto-delete
        false,    // internal
        false,    // no-wait
        nil,
	)
	if err != nil{
		log.Printf("No se pudo declarar el exchange de RabbitMQ: %v", err)
	}
	//Declarar queue
	queueIdentity := "queue_" + salaID
	queue, err := r.Utils.Channel.QueueDeclare(
		queueIdentity,
		true,  // durable
        false, // delete cuando no se use
        false, // exclusiva
        false, // no-wait
        nil,
	)
	if err != nil{
		log.Printf("No se pudo declarar la queue de RabbitMQ: %v", err)
	}
	//Binding del exchange
	err = r.Utils.Channel.QueueBind(
		queue.Name,
		salaID,
		"chat_exchange",
		false,
        nil,
	)
	if err != nil{
		log.Printf("No se pudo hacer el binding de RabbitMQ: %v", err)
	}
	//Consumir mensajes de la queue
	getMensajes, err := r.Utils.Channel.Consume(
		queue.Name, //Debe ser el mismo de QueueDeclare
		"",
		true,   // auto-ack
        false,  // exclusive
        false,  // no local
        false,  // no wait
        nil,
	)
	if err != nil{
		log.Printf("error consumiendo mensajes de RabbitMQ: %v", err)
	}
	//Reenviar mensaje por websockets
	go func(){
		for d := range getMensajes{
			err := json.Unmarshal(d.Body, &mensajes)
			if err != nil{
				log.Printf("No se pudo parsear el JSON al struct mensajes: %v", err)
				continue
			}
			r.db.db.First(&usuario, mensajes.UsuarioID)
			mensajes.UsuarioNombre = usuario.Usuario
			r.db.db.Create(&mensajes)
			r.broadcast(mensajes, salaID)
		}
	}()
}

//Envia mensajes a los clientes conectados a la sala
func (r *ChatRepositories) broadcast(mensaje models.Message, salaID string){
	fmt.Println("Broadcast")
	for conn, salaIDinRabbitMQ := range r.Clients{
		if salaID == salaIDinRabbitMQ{
			conn.WriteJSON(mensaje)
		}
	}
}

func (r *ChatRepositories) messageToJSON(mensaje models.Message) []byte{
	fmt.Println("messageToJSON")
	data, _ := json.Marshal(mensaje)
	return data
}
