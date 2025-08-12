package utils

type Utils struct {
	AllowedOrigins []string
}

func NewUtils() *Utils {
	return &Utils{AllowedOrigins: []string{"http://localhost:8080"}}
}
