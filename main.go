package main

import (
	"ChatSpot/controllers"
	"ChatSpot/middlewares"
	"ChatSpot/models"
	"ChatSpot/repositories"
	"ChatSpot/services"
	"ChatSpot/utils"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	dbConn, err := utils.ConectarDB()
	if err != nil {
		fmt.Printf("Error de db: %v", err)
	}
	err = dbConn.AutoMigrate(&models.Usuarios{}, &models.Salas{}, &models.UsuariosSala{}, &models.Message{})
	if err != nil{
		log.Printf("Error de db: %v", err)
	}
	//Repositorios
	utils := utils.NewUtils()
	gormRepositories := repositories.NewGormRepositories(dbConn)
	chatRepositories := repositories.NewChatRepositories(utils, gormRepositories)
	//Servicios
	gormServices := services.NewGormServices(gormRepositories)
	chatServices := services.NewChatServices(chatRepositories)
	//Utils
	//RabbitMQ
	rabbitMQ, err := utils.ConectarRabbitMQ()
	if err != nil{
		fmt.Printf("no se conecto a RabbitMQ: %v", err)
	}
	defer rabbitMQ.CloseRabbitMQ()
	
	//Controladores
	chatController := controllers.NewChatController(chatServices, utils, chatRepositories, gormServices)
	autenticacionController := controllers.NewAutenticacionController(gormServices)
	salasController := controllers.NewSalasController(gormServices)
	//Middlewares
	middlewares := middlewares.NewMiddleware(gormServices)
	r := gin.Default()
	r.LoadHTMLGlob("views/*.html")
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})
	r.GET("/registro", autenticacionController.RegistroGET)
	r.POST("/registro", autenticacionController.RegistroPOST)
	r.GET("/login", autenticacionController.LoginGET)
	r.POST("/login", autenticacionController.LoginPOST)
	autenticado := r.Group("/autenticado")
	autenticado.Use(middlewares.ValidarUsuario())
	{
		autenticado.GET("/websockets", chatController.ChatController)
		autenticado.GET("/chat/:id", chatController.ChatView)
		autenticado.GET("/logout", autenticacionController.Logout)
		autenticado.GET("/salas", salasController.ListarSalas)
		autenticado.POST("/salas", salasController.NuevaSala)
	}
	r.Run()
}
