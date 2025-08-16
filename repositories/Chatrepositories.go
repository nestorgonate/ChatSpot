package repositories

import (
	"ChatSpot/models"
	"ChatSpot/utils"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type IChatRepositories interface {
	HandleConnections(conn *websocket.Conn, salaID string, usuarioUsuario string)
}

type ChatRepositories struct {
	ConexionesASalas     map[*websocket.Conn]string //Solo tiene llaves y el struct no tiene valor, la clave es la conexion y el valor es el ID de la sala map[0x2ec4b68:1]
	Utils                *utils.Utils
	SalaConsumers        map[string]bool //Mapa de salas que ya tienen un consumidor en RabbitMQ, evita duplicar consumidores para la misma sala map[1:true]
	db                   *GormRepositories
	redisClient          *redis.Client
	ConexionesDeUsuarios map[*websocket.Conn]string //La clave es la conexion y el valor es el nombre del usuario

}

func NewChatRepositories(utils *utils.Utils, db *GormRepositories) *ChatRepositories {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	return &ChatRepositories{
		ConexionesASalas:     make(map[*websocket.Conn]string),
		SalaConsumers:        make(map[string]bool),
		Utils:                utils,
		db:                   db,
		redisClient:          redisClient,
		ConexionesDeUsuarios: make(map[*websocket.Conn]string),
	}
}

func (r *ChatRepositories) HandleConnections(conn *websocket.Conn, salaID string, usuarioUsuario string) {
	fmt.Println("HandelConnections")
	defer func() {
		delete(r.ConexionesASalas, conn)
		delete(r.ConexionesDeUsuarios, conn)
		//Enviar broadcast con la lista de usuarios cuando se desconectan
		r.broadcastDeUsuariosEnSala(salaID)
		log.Print("Cerrando conexion websocket")
		conn.Close()
	}()
	//Cada conexion sabe su salaID
	r.ConexionesASalas[conn] = salaID
	log.Printf("Conexiones de salas: %v", r.ConexionesASalas)
	//Cada conexion sabe el usuario
	r.ConexionesDeUsuarios[conn] = usuarioUsuario
	log.Printf("Conexiones de usuarios: %v", r.ConexionesDeUsuarios)
	//Enviar broadcast con la lista de usuarios
	r.broadcastDeUsuariosEnSala(salaID)
	//Validar si la sala tiene un consumidor
	if !r.SalaConsumers[salaID] {
		r.SalaConsumers[salaID] = true
		go r.ConsummerRabbitMQ(salaID)
	}
	for {
		mensaje := models.Message{}
		//Tipo de broadcast
		mensaje.Tipo = "broadcastDeMensaje"
		err := conn.ReadJSON(&mensaje)
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("Cliente desconectado: %v", err)
				break
			} else {
				log.Printf("Error de websocket: %v", err)
			}
			return
		}
		//Publicar mensajes en RabbitMQ
		salaID := r.Utils.UintToString(mensaje.SalaID)
		err = r.Utils.Channel.Publish(
			"chat_exchange", //Exchange
			salaID,          //Routing key, debe coincidir con RabbitMQ.ChannelQueBinding
			false,           //Mandatory
			false,           //Inmediate
			amqp091.Publishing{
				DeliveryMode: amqp091.Persistent,
				ContentType:  "application/json",
				Body:         r.messageToJSON(mensaje),
			},
		)
		if err != nil {
			fmt.Printf("No se publico el mensaje a RabbitMQ: %v", err)
		}
	}
}

// Declara exchange, queue, binding, consume y reenvia mensajes
func (r *ChatRepositories) ConsummerRabbitMQ(salaID string) {
	key := "latest_messages"
	fmt.Println("ConsummerRabbitMQ")
	//Declarar el exchange
	err := r.Utils.Channel.ExchangeDeclare(
		"chat_exchange",
		"direct",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,
	)
	if err != nil {
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
	if err != nil {
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
	if err != nil {
		log.Printf("No se pudo hacer el binding de RabbitMQ: %v", err)
	}
	//Consumir mensajes de la queue
	getMensajes, err := r.Utils.Channel.Consume(
		queue.Name, //Debe ser el mismo de QueueDeclare
		"",
		false, // auto-ack
		false, // exclusive
		false, // no local
		false, // no wait
		nil,
	)
	if err != nil {
		log.Printf("error consumiendo mensajes de RabbitMQ: %v", err)
	}
	//Reenviar mensaje por websockets
	go func() {
		for d := range getMensajes {
			var mensajes models.Message
			var usuario models.Usuarios
			err := json.Unmarshal(d.Body, &mensajes)
			if err != nil {
				log.Printf("No se pudo parsear el JSON al struct mensajes: %v", err)
				continue
			}
			r.db.db.Model(models.Usuarios{}).Where("id = ?", mensajes.UsuarioID).Find(&usuario)
			mensajes.UsuarioNombre = usuario.Usuario
			log.Printf("Tipo de mensaje recibido: %v", mensajes.Tipo)
			//Si es broadcast de mensaje es el mensaje de un usuario, si es lista de usuarios es la lista de usuarios conectados
			switch mensajes.Tipo {
			case "broadcastDeMensaje":
				r.db.db.Create(&mensajes)
				r.redisClient.Del(context.Background(), key)
				log.Print("Cache de redis borrada al enviar un mensaje")
				r.broadcast(mensajes, salaID)
			case "listaDeUsuarios":
				r.broadcast(mensajes, salaID)
			}
		}
	}()
}

// Envia mensajes a los clientes conectados a la sala
func (r *ChatRepositories) broadcast(mensaje models.Message, salaID string) {
	fmt.Println("Broadcast")
	for conn, salaIDinRabbitMQ := range r.ConexionesASalas {
		if salaID == salaIDinRabbitMQ {
			conn.WriteJSON(mensaje)
		}
	}
}

func (r *ChatRepositories) messageToJSON(mensaje models.Message) []byte {
	fmt.Println("messageToJSON")
	data, _ := json.Marshal(mensaje)
	return data
}

// Itera en el mapa de Conexiones a salas, si sala coincide con salaID, se agrega al slices Usuarios el valor que tenga la misma conexion en salas y usuarios
func (r *ChatRepositories) usuariosEnSala(salaID string) []string {
	var usuarios []string
	for conn, sala := range r.ConexionesASalas {
		if sala == salaID {
			usuarios = append(usuarios, r.ConexionesDeUsuarios[conn])
		}
	}
	return usuarios
}

func (r *ChatRepositories) broadcastDeUsuariosEnSala(salaID string) {
	log.Print("Broadcast de la lista de usuarios")
	listaDeUsuariosEnSala := r.usuariosEnSala(salaID)
	salaIDuint := r.Utils.StringToUint(salaID)
	var mensaje models.Message = models.Message{
		SalaID:          salaIDuint,
		ListaDeUsuarios: listaDeUsuariosEnSala,
		Tipo:            "listaDeUsuarios",
	}
	err := r.Utils.Channel.Publish(
		"chat_exchange", //Exchange
		salaID,          //Routing key, debe coincidir con RabbitMQ.ChannelQueBinding
		false,           //Mandatory
		false,           //Inmediate
		amqp091.Publishing{
			DeliveryMode: amqp091.Persistent,
			ContentType:  "application/json",
			Body:         r.messageToJSON(mensaje),
		},
	)
	if err != nil {
		fmt.Printf("No se publico el mensaje a RabbitMQ: %v", err)
	}
}
