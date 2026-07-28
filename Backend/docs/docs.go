package docs

import "github.com/swaggo/swag"

type s struct{}

func (s *s) ReadDoc() string {
	return "{}"
}

func init() {
	swag.Register(swag.Name, &s{})
}
