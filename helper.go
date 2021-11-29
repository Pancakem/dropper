package main

import (
	"github.com/asaskevich/govalidator"
)


func validateURL(uri string) bool {
	return govalidator.IsURL(uri)
}
