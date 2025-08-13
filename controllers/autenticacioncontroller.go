package controllers

import (
	"ChatSpot/models"
	"ChatSpot/services"
	"ChatSpot/utils"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AutenticacionController struct {
	service *services.GormServices
	utils   *utils.Utils
}

func NewAutenticacionController(service *services.GormServices) *AutenticacionController {
	return &AutenticacionController{service: service}
}

func (r *AutenticacionController) LoginGET(c *gin.Context){
	c.HTML(http.StatusOK, "login.html", nil)
}

func (r *AutenticacionController) LoginPOST(c *gin.Context) {
	// Verificar si es una solicitud AJAX/JSON
	isJsonRequest := strings.Contains(c.GetHeader("Accept"), "application/json") ||
		c.GetHeader("X-Requested-With") == "XMLHttpRequest"

	email := c.PostForm("email")
	password := c.PostForm("password")

	var usuario *models.Usuarios
	usuario, err := r.service.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if isJsonRequest {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"errors": map[string]string{
						"email": "Email no registrado",
					},
				})
			} else {
				c.HTML(http.StatusOK, "login.html", gin.H{
					"Error": "Email no registrado",
					"Email": email,
				})
			}
			return
		}

		if isJsonRequest {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Error en la base de datos",
			})
		} else {
			c.HTML(http.StatusOK, "login.html", gin.H{
				"Error": "Error en la base de datos",
				"Email": email,
			})
		}
		return
	}

	if !r.utils.ValidatePassword(password, *usuario.Password) {
		if isJsonRequest {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"errors": map[string]string{
					"password": "Contraseña incorrecta",
				},
			})
		} else {
			c.HTML(http.StatusOK, "login.html", gin.H{
				"Error": "Contraseña incorrecta",
				"Email": email,
			})
		}
		return
	}

	isAuthenticated2fa := false
	tokenID, err := r.utils.GetJWT(usuario.ID, usuario.Email, usuario.Is_2fa, isAuthenticated2fa)
	if err != nil {
		if isJsonRequest {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Error al generar token de sesión",
			})
		} else {
			c.HTML(http.StatusOK, "login.html", gin.H{
				"Error": "Error al generar token de sesión",
				"Email": email,
			})
		}
		return
	}

	c.SetCookie("usuarioJWT", tokenID, 86400, "/", "", false, true)

	if isJsonRequest {
		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"redirect": "/autenticado/perfil",
		})
	} else {
		if usuario.Is_2fa {
			c.Redirect(http.StatusSeeOther, "/verificarFA")
		} else {
			c.Redirect(http.StatusSeeOther, "/autenticado/salas")
		}
	}
}

func (r *AutenticacionController) RegistroGET(c *gin.Context) {
	c.HTML(http.StatusOK, "registro.html", nil)
}

func (r *AutenticacionController) RegistroPOST(c *gin.Context) {
	var usuario models.Usuarios
	if err := c.ShouldBind(&usuario); err != nil {
		c.Redirect(http.StatusSeeOther, "/registro?error=datos_incorrectos")
		return
	}
	log.Printf("Usuario form: %v", usuario)
	usuario.GoogleID = nil
	// Validar nombre (que coincida con tu formulario HTML)
	if strings.TrimSpace(usuario.Nombre) == "" {
		c.Redirect(http.StatusSeeOther, "/registro?error=nombre_requerido")
		return
	}
	//Hashear
		hash, err := r.utils.Hash_password(*usuario.Password)
		if err != nil {
			c.Redirect(http.StatusSeeOther, "/registro?error=error_hasheo")
			return
		}
		previousPassword, _ := json.Marshal([]string{hash})
		usuario.PreviousPasswords = previousPassword
		usuario.Password = &hash // asignamos puntero al hash

	// Establecer foto por defecto si no viene de Google
	if usuario.FotoPerfil == "" {
		usuario.FotoPerfil = "/assets/perfil-user.jpg"
	}

	// Crear usuario
	_, err = r.service.AddUser(&usuario)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			c.Redirect(http.StatusSeeOther, "/registro?error=email_ya_registrado")
			return
		}
	}
	// Generar JWT y redirigir
	isAuthenticated2fa := false
	tokenID, err := r.utils.GetJWT(usuario.ID, usuario.Email, usuario.Is_2fa, isAuthenticated2fa)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/registro?error=error_jwt")
		return
	}

	c.SetCookie("usuarioJWT", tokenID, 86400, "/", "", false, true)
	c.Redirect(http.StatusSeeOther, "/autenticado/salas")
}

func (r *AutenticacionController) Logout(c *gin.Context){
	c.SetCookie("usuarioJWT", "", -1, "/", "", false, true)
	c.Redirect(http.StatusSeeOther, "/")
}
