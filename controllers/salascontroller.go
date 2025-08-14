package controllers

import (
	"ChatSpot/models"
	"ChatSpot/services"
	"ChatSpot/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SalasController struct {
	GormServices *services.GormServices
	Utils        *utils.Utils
}

func NewSalasController(gormServices *services.GormServices) *SalasController {
	return &SalasController{GormServices: gormServices}
}

// Agrega salas POST
func (r *SalasController) NuevaSala(c *gin.Context) {
	var salaForm struct {
		Nombre  string `form:"nombre"`
		Privado bool   `form:"privado"`
	}
	var sala models.Salas
	err := c.ShouldBind(&salaForm)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	sala.Nombre = salaForm.Nombre
	sala.Privado = salaForm.Privado
	usuarioID := r.Utils.GetUsuarioIdFromJWT(c, "usuarioJWT", "usuarioID")
	sala.UsuarioID = usuarioID
	log.Printf("ID del usuario que crea la sala: %v", sala.UsuarioID)
	err = r.GormServices.AddSalaToDatabase(&sala)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.Redirect(http.StatusSeeOther, "/autenticado/salas")
}

// Obtener todas las salas para enviarlos a la view GET
func (r *SalasController) ListarSalas(c *gin.Context) {
	salas, _ := r.GormServices.GetAllSalas()
	c.HTML(http.StatusOK, "salas.html", gin.H{
		"Salas": salas,
	})
}

func (r *SalasController) BorrarSala(c *gin.Context) {
	salaIDString := c.Query("id")
	salaID := r.Utils.StringToUint(salaIDString)
	usuarioID := r.Utils.GetUsuarioIdFromJWT(c, "usuarioJWT", "usuarioID")
	canDeleteSala := r.GormServices.DeleteSalaByID(salaID, usuarioID)
	//canDeleteSala es false, no se puede borrar la sala porque el propietario no es el que borro la sala
	if !canDeleteSala{
		c.Redirect(http.StatusSeeOther, "/autenticado/salas?error=no_eres_el_propietario_de_la_sala")
		return
	}
	c.Redirect(http.StatusSeeOther, "/autenticado/salas?success=sala_borrada")
}
