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
	err = dbConn.AutoMigrate(&models.Usuarios{})
	if err != nil{
		log.Printf("Error de db: %v", err)
	}
	//Repositorios
	gormRepositories := repositories.NewGormRepositories(dbConn)
	chatRepositories := repositories.NewChatRepositories()
	//Servicios
	gormServices := services.NewGormServices(gormRepositories)
	chatServices := services.NewChatServices(chatRepositories)
	//Utils
	utils := utils.NewUtils()
	//Controladores
	chatController := controllers.NewChatController(chatServices, utils, chatRepositories, gormServices)
	autenticacionController := controllers.NewAutenticacionController(gormServices)
	//Middlewares
	middlewares := middlewares.NewMiddleware(gormServices)
	r := gin.Default()
	r.LoadHTMLGlob("views/*.html")
	r.GET("/registro", autenticacionController.RegistroGET)
	r.POST("/registro", autenticacionController.RegistroPOST)
	r.GET("/login", autenticacionController.LoginGET)
	r.POST("/login", autenticacionController.LoginPOST)
	autenticado := r.Group("/autenticado")
	autenticado.Use(middlewares.ValidarUsuario())
	{
		autenticado.GET("/websockets", chatController.ChatController)
		autenticado.GET("/chat", chatController.ChatView)
		autenticado.GET("/logout", autenticacionController.Logout)
	}
	r.Run()
}
