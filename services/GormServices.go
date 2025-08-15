package services

import "ChatSpot/models"

type IGormRepository interface {
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
	GetLastMessages(salaID uint) []models.Message
	AddMessageToDatabase(mensaje *models.Message)
	DeleteSalaByID(salaID uint, usuarioID uint) bool
}

type GormServices struct {
	repository IGormRepository
}

func NewGormServices(repository IGormRepository) *GormServices {
	return &GormServices{repository: repository}
}

func (r *GormServices) ListUsers() ([]models.Usuarios, error) {
	return r.repository.GetAll() //Delega la funcion al repository
}

func (r *GormServices) GetUserByID(id uint) (*models.Usuarios, error) {
	return r.repository.GetByID(id) //Delega la funcion al repository
}

func (r *GormServices) GetUserByUsuario(usuario string) (*models.Usuarios, error) {
	return r.repository.GetUserByUser(usuario) //Delega la funcion al repository
}

func (r *GormServices) AddUser(usuario *models.Usuarios) (*models.Usuarios, error) {
	return r.repository.AddUser(usuario) //Delega la funcion al repository
}

func (r *GormServices) SaveUserGoogle(content []byte) (*models.Usuarios, error) {
	return r.repository.SaveUserGoogle(content) //Delega la funcion al repository
}

func (r *GormServices) UpdatePassword(newPassword string, id uint) error {
	return r.repository.UpdatePassword(newPassword, id) //Delega la funcion al repository
}

func (r *GormServices) IsPreviousPassword(id uint, newPassword string) bool {
	return r.repository.IsPreviousPassword(id, newPassword) //Delega la funcion al repository
}

func (r *GormServices) AddSalaToDatabase(sala *models.Salas) error {
	return r.repository.AddSalaToDatabase(sala) //Delega la funcion al repository
}

func (r *GormServices) GetAllSalas() ([]models.Salas, error) {
	return r.repository.GetAllSalas() //Delega la funcion al repository
}

func (r *GormServices) GetSalaByID(salaID uint) (*models.Salas, error) {
	return r.repository.GetSalaByID(salaID)
}

func (r *GormServices) GetLastMessages(salaID uint) []models.Message {
	return r.repository.GetLastMessages(salaID)
}

func (r *GormServices) AddMessageToDatabase(mensaje *models.Message) {
	r.repository.AddMessageToDatabase(mensaje)
}

func (r *GormServices) DeleteSalaByID(salaID uint, usuarioID uint) bool {
	return r.repository.DeleteSalaByID(salaID, usuarioID)
}
