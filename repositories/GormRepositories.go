package repositories

import (
	"ChatSpot/models"
	"encoding/json"
	"errors"
	"log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type IGormRepositories interface {
	GetAll() ([]models.Usuarios, error)
	GetByID(id uint) (*models.Usuarios, error)
	GetUserByUser(usuarioUsuario string) (*models.Usuarios, error)
	AddUser(*models.Usuarios) (*models.Usuarios, error)
	SaveUserGoogle(content []byte) (*models.Usuarios, error)
	UpdatePassword(newPassword string, id uint) error
	IsPreviousPassword(id uint, newPassword string) bool
	AddSalaToDatabase(sala *models.Salas) error
	GetAllSalas() ([]models.Salas, error)
	GetSalaByID(salaID uint) (*models.Salas, error)
	GetLastMessages(salaID uint) ([]models.Message, error)
	AddMessageToDatabase(mensaje *models.Message)
	DeleteSalaByID(salaID uint, usuarioID uint) bool
}

type GormRepositories struct {
	db *gorm.DB
}

func NewGormRepositories(db *gorm.DB) *GormRepositories {
	return &GormRepositories{db: db}
}

func (r *GormRepositories) GetAll() ([]models.Usuarios, error) {
	var users []models.Usuarios
	err := r.db.Find(&users)
	return users, err.Error
}

func (r *GormRepositories) GetByID(id uint) (*models.Usuarios, error) {
	var usuario models.Usuarios
	result := r.db.Where("id = ?", id).First(&usuario)
	return &usuario, result.Error
}

func (r *GormRepositories) GetUserByUser(usuarioUsuario string) (*models.Usuarios, error) {
	var usuario models.Usuarios
	result := r.db.Where("usuario = ?", usuarioUsuario).First(&usuario)
	log.Print("Resultado getbyemail: ", result)
	return &usuario, result.Error
}

func (r *GormRepositories) AddUser(usuario *models.Usuarios) (*models.Usuarios, error) {
	result := r.db.Create(usuario)
	return usuario, result.Error
}

func (r *GormRepositories) SaveUserGoogle(content []byte) (*models.Usuarios, error) {
	var googleUser models.GoogleUser
	err := json.Unmarshal(content, &googleUser)
	if err != nil {
		return nil, err
	}
	var usuario *models.Usuarios
	usuario, err = r.GetUserByUser(googleUser.Email)
	//Si hay error en la solicitud a db es porque el usuario no existe
	if errors.Is(err, gorm.ErrRecordNotFound) {
		usuario = &models.Usuarios{
			Nombre:     googleUser.Nombre,
			Password:   nil,
			FotoPerfil: googleUser.FotoPerfil,
			GoogleID:   &googleUser.GoogleID,
		}
		log.Print("Registrando usuario en db")
		r.AddUser(usuario)
	}
	log.Print("Login usuario en db")
	return usuario, nil
}

func (r *GormRepositories) UpdatePassword(newPassword string, id uint) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	hashedStr := string(hashedPassword)
	//Leer usuario en la db para saber las contraseñas
	var usuario models.Usuarios
	r.db.Model(models.Usuarios{}).Where("id = ?", id).First(&usuario)

	//Unmarshal de las contraseñas anteriores
	var previousPassword []string
	json.Unmarshal(usuario.PreviousPasswords, &previousPassword)
	log.Print("Unmarshal de passwords: ", previousPassword)
	if usuario.Password != nil {
		previousPassword = append(previousPassword, *usuario.Password)
	}

	//Marshal de las contraseñas para agregar la nueva contraseña
	previousPasswordMarshall, _ := json.Marshal(previousPassword)
	return r.db.Model(models.Usuarios{}).Where("id = ?", id).Updates(map[string]interface{}{
		"password":           hashedStr,
		"previous_passwords": datatypes.JSON(previousPasswordMarshall),
	}).Error
}

func (r *GormRepositories) IsPreviousPassword(id uint, newPassword string) bool {
	var usuario models.Usuarios
	r.db.Model(models.Usuarios{}).Where("id = ?", id).First(&usuario)
	//Unmarshal de las contraseñas anteriores
	var previousPasswords []string
	json.Unmarshal(usuario.PreviousPasswords, &previousPasswords)
	if err := json.Unmarshal(usuario.PreviousPasswords, &previousPasswords); err != nil {
		return false
	}
	//Comprar las contraseñas
	for _, hash := range previousPasswords {
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(newPassword)) == nil {
			//Coincide la contraseña
			return true
		}
	}
	return false
}

func (r *GormRepositories) AddSalaToDatabase(sala *models.Salas) error{
	err := r.db.Model(models.Salas{}).Create(sala)
	return err.Error
}

func (r *GormRepositories) GetAllSalas() ([]models.Salas, error){
	var salas []models.Salas
	err := r.db.Find(&salas)
	return salas, err.Error
}

func (r *GormRepositories) GetSalaByID(id uint) (*models.Salas, error){
	var sala models.Salas

	err := r.db.Model(models.Salas{}).Where("id = ?", id).First(&sala)
	return &sala, err.Error
}

func (r *GormRepositories) GetLastMessages(salaID uint) ([]models.Message, error){
	var mensajes []models.Message
	err := r.db.Preload("Usuarios").Model(models.Message{}).Where("sala_id = ?", salaID).Order("created_at desc").Limit(50).Find(&mensajes)
	//Inventir slice para mostrar el mensaje nuevo al final del scroll del chat
	for i, j := 0, len(mensajes)-1; i < j; i, j = i+1, j-1 {
        mensajes[i], mensajes[j] = mensajes[j], mensajes[i]
    }
	return mensajes, err.Error
}

func (r *GormRepositories) AddMessageToDatabase(mensaje *models.Message){
	r.db.Create(&mensaje)
}

func (r *GormRepositories) DeleteSalaByID(salaID uint, usuarioID uint) bool{
	resultado := r.db.Model(models.Salas{}).Where("id = ? and usuario_id = ?", salaID, usuarioID).Unscoped().Delete(models.Salas{})
	var canDelete bool = true
	//No se borro porque la sala no existe o el usuario no es el propietario
	if resultado.RowsAffected == 0 {
		//Establece que no se puede borrar
		canDelete = false
	}
	return canDelete
}