package models

import "gorm.io/gorm"

type Salas struct {
	gorm.Model
	Nombre string `json:"nombre" form:"nombre"`
	Privado bool `gorm:"default:false" json:"privada" form:"privada"`
	UsuarioID uint
	Usuarios Usuarios `gorm:"foreignKey:UsuarioID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
