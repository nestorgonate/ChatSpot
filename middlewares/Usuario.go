package middlewares

import (
	"ChatSpot/models"
	"ChatSpot/services"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type Middleware struct {
	service *services.GormServices
}

func NewMiddleware(service *services.GormServices) *Middleware {
	return &Middleware{service: service}
}

func (r *Middleware) ValidarUsuario() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Obtener token JWT de las cookies
		getToken, err := ctx.Cookie("usuarioJWT")
		if err != nil {
			log.Print("No se encontró cookie de autenticación")
			redirectToLogin(ctx)
			return
		}

		// Parsear y validar token
		token, err := jwt.Parse(getToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			log.Print("Token JWT inválido")
			redirectToLogin(ctx)
			return
		}

		// Obtener claims del token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Print("Claims JWT inválidos")
			redirectToLogin(ctx)
			return
		}

		// Verificar 2FA si está habilitado
		is2fa, _ := claims["is2fa"].(bool)
		isAuthenticated2fa, _ := claims["isAuthenticated2fa"].(bool)
		if is2fa && !isAuthenticated2fa {
			log.Print("Requiere autenticación 2FA")
			redirectToLogin(ctx)
			return
		}

		// Obtener ID de usuario
		usuarioID, ok := claims["usuarioID"].(float64)
		if !ok {
			log.Print("ID de usuario inválido en JWT")
			redirectToLogin(ctx)
			return
		}

		// Obtener usuario completo de la base de datos
		var usuario *models.Usuarios
		usuario, err = r.service.GetUserByID(uint(usuarioID))
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Print("Usuario no encontrado en la base de datos")
			} else {
				log.Printf("Error al buscar usuario: %v", err)
			}
			redirectToLogin(ctx)
			return
		}

		// Establecer datos en el contexto
		ctx.Set("usuarioID", uint(usuarioID))
		ctx.Set("usuario", usuario) // Esto es lo que necesita tu controlador de password
		ctx.Set("claims", claims)

		ctx.Next()
	}
}

func redirectToLogin(ctx *gin.Context) {
	// Limpiar cookie si es inválida
	ctx.SetCookie("usuarioJWT", "", -1, "/", "", false, true)
	ctx.Redirect(http.StatusSeeOther, "/login")
	ctx.Abort()
}
