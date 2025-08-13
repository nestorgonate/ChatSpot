package utils

import (
	"ChatSpot/models"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Utils struct {
	AllowedOrigins []string
	Usuarios       *models.Usuarios
}

func NewUtils() *Utils {
	return &Utils{AllowedOrigins: []string{"http://localhost:8080"}}
}

// Verifica si la contraseña actual es correcta
func (r *Utils) GetJWT(usuarioID uint, email string, is2fa bool, isAutenticated2fa bool) (string, error) {
	//Asigna el ID del usuario a una cookie, si tiene 2fa y si se ha autenticado con 2fa para validarlos en los middlewares
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"usuarioID":          usuarioID,
		"exp":                time.Now().Add(time.Hour * 24).Unix(),
		"email":              email,
		"is2fa":              is2fa,
		"isAuthenticated2fa": isAutenticated2fa,
	})
	//Obtiene el secret del .env para validar el JWT
	tokenID, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}
	return tokenID, nil
}

func (r *Utils) Hash_password(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (r *Utils) ValidatePassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (r *Utils) UsuarioIDJWT(c *gin.Context, cookie, valorJWT string) float64 {
	obtenerJWT, _ := c.Cookie(cookie)
	token, _ := jwt.Parse(obtenerJWT, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metodo incorrecto firma JWT")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	claims, _ := token.Claims.(jwt.MapClaims)
	usuarioID := claims[valorJWT].(float64)
	return usuarioID
}

func ConectarDB() (*gorm.DB, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("no se cargo el archivo .env: %w", err)
	}
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True", user, password, host, port, dbname)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("error conectando con DB: %w", err)
	}
	return db, nil
}
