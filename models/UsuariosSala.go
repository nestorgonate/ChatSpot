package models

import "gorm.io/gorm"

//Aun no se usa
type UsuariosSala struct {
	gorm.Model
	UsuarioID uint
	SalaID uint
	Usuarios Usuarios `gorm:"foreignKey:UsuarioID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Salas Salas `gorm:"foreignKey:SalaID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}