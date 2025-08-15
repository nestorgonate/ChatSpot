package models

import "gorm.io/gorm"

type Message struct {
	gorm.Model
	UsuarioID uint `json:"usuarioID"`
	SalaID uint `json:"salaID"`
	Mensaje string `gorm:"type:longtext;not null" json:"mensaje"`
	Usuarios Usuarios `gorm:"foreignKey:UsuarioID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Salas Salas `gorm:"foreignKey:SalaID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UsuarioNombre string `gorm:"-" json:"usuarioMessage"`
	ListaDeUsuarios []string `gorm:"-" json:"listaDeUsuarios"`
	Tipo string `gorm:"-" json:"tipoBroadcast"`
}
