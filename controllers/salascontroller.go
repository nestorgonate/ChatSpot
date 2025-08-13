package controllers

import (
	"ChatSpot/models"
	"ChatSpot/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SalasController struct {
	GormServices *services.GormServices
}

func NewSalasController(gormServices *services.GormServices) *SalasController {
	return &SalasController{GormServices: gormServices}
}


//Agrega salas POST
func (r *SalasController) NuevaSala(c *gin.Context){
	var datosSala struct{
		Nombre string `json:"nombre" form:"nombre"`
		Privada bool `json:"privada" form:"privada"`
	}
	var sala models.Salas
	err := c.ShouldBind(&datosSala)
	if err != nil{
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	sala.Nombre = datosSala.Nombre
	sala.Privada = datosSala.Privada
	err = r.GormServices.AddSalaToDatabase(&sala)
	if err != nil{
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.Redirect(http.StatusSeeOther, "/autenticado/salas")
}


//Obtener todas las salas para enviarlos a la view GET
func (r *SalasController) ListarSalas(c *gin.Context){
	salas, _ := r.GormServices.GetAllSalas()
	c.HTML(http.StatusOK, "salas.html", gin.H{
		"Salas": salas,
	})
}

