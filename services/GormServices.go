package services

import "ChatSpot/models"

type IGormRepository interface {
	GetAll() ([]models.Usuarios, error)
	GetByID(id uint) (*models.Usuarios, error)
	GetByEmail(email string) (*models.Usuarios, error)
	AddUser(*models.Usuarios) (*models.Usuarios, error)
	SaveUserGoogle(content []byte) (*models.Usuarios, error)
	UpdatePassword(newPassword string, id uint) error
	IsPreviousPassword(id uint, newPassword string) bool
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

func (r *GormServices) GetUserByEmail(email string) (*models.Usuarios, error) {
	return r.repository.GetByEmail(email) //Delega la funcion al repository
}

func (r *GormServices) AddUser(usuario *models.Usuarios) (*models.Usuarios, error) {
	return r.repository.AddUser(usuario) //Delega la funcion al repository
}

func (r *GormServices) SaveUserGoogle(content []byte) (*models.Usuarios, error) {
	return r.repository.SaveUserGoogle(content)
}

func (r *GormServices) UpdatePassword(newPassword string, id uint) error {
	return r.repository.UpdatePassword(newPassword, id)
}

func (r *GormServices) IsPreviousPassword(id uint, newPassword string) bool {
	return r.repository.IsPreviousPassword(id, newPassword)
}
