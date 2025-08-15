package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Usuarios struct {
	gorm.Model
	Nombre             string         `json:"nombre" form:"nombre"`
	Usuario            string         `form:"usuario"`
	Password           *string        `json:"password" form:"password"`
	Secret_2fa         string         `json:"secret_2fa" form:"secret_2fa"`
	Is_2fa             bool           `gorm:"default:false"`
	FotoPerfil         string         `gorm:"size:255;default:'/assets/perfil-user.jpg'" json:"foto_perfil"`
	GoogleID           *string        `gorm:"size:255;uniqueIndex" json:"-"` // Para usuarios de Google
	LastPasswordChange time.Time      `gorm:"autoUpdateTime" json:"-"`
	PreviousPasswords  datatypes.JSON `gorm:"type:json" json:"-"` // Para almacenar hashes de contraseñas anteriores
}
