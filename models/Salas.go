package models

import "gorm.io/gorm"

type Salas struct {
	gorm.Model
	Nombre string `json:"nombre" form:"nombre"`
	Privada bool `gorm:"default:false" json:"privada" form:"privada"`
}
