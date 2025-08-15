package utils

import (
	"ChatSpot/models"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/rabbitmq/amqp091-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Utils struct {
	AllowedOrigins []string
	Usuario        models.Usuarios
	Conn           *amqp091.Connection
	Channel        *amqp091.Channel
}

func NewUtils() *Utils {
	return &Utils{
		AllowedOrigins: []string{"http://localhost:8080"},
	}
}

// Verifica si la contraseña actual es correcta
func (r *Utils) GetJWT(usuarioID uint, is2fa bool, isAutenticated2fa bool) (string, error) {
	//Asigna el ID del usuario a una cookie, si tiene 2fa y si se ha autenticado con 2fa para validarlos en los middlewares
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"usuarioID":          usuarioID,
		"exp":                time.Now().Add(time.Hour * 24).Unix(),
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

// Obtiene el usuarioID del JWT
func (r *Utils) GetUsuarioIdFromJWT(c *gin.Context, cookie, valorJWT string) uint {
	obtenerJWT, _ := c.Cookie(cookie)
	token, _ := jwt.Parse(obtenerJWT, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metodo incorrecto firma JWT")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	claims, _ := token.Claims.(jwt.MapClaims)
	usuarioIDfloat := claims[valorJWT].(float64)
	usuarioID := uint(usuarioIDfloat)
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

func (r *Utils) ConectarRabbitMQ() (*Utils, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("no se cargo el archivo .env: %w", err)
	}
	user := os.Getenv("RABBITMQ_USER")
	password := os.Getenv("RABBITMQ_PASS")
	host := os.Getenv("RABBITMQ_HOST")
	port := os.Getenv("RABBITMQ_PORT")
	dsn := fmt.Sprintf("amqp://%s:%s@%s:%s", user, password, host, port)
	conn, err := amqp091.Dial(dsn)
	if err != nil {
		fmt.Printf("No se pudo conectar a RabbitMQ: %v", err)
	}
	channel, err := conn.Channel()
	if err != nil {
		fmt.Printf("No se pudo cargar el canal de RabbitMQ: %v", err)
	}
	r.Channel = channel
	return &Utils{
		Conn:    conn,
		Channel: channel,
	}, err
}

func (r *Utils) CloseRabbitMQ() {
	r.Channel.Close()
	r.Conn.Close()
}

func (r *Utils) UintToString(number uint) string {
	uintString := strconv.FormatUint(uint64(number), 10)
	return uintString
}

func (r *Utils) StringToUint(number string) uint {
	uintUint, _ := strconv.ParseUint(number, 10, 64)
	return uint(uintUint)
}
