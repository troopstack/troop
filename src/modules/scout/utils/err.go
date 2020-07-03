package utils

import (
	"errors"
	"log"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s", msg, err)
	}
}

func CoverErrorMessage() {
	if message := recover(); message != nil {
		var err error
		switch x := message.(type) {
		case string:
			err = errors.New(x)
		case error:
			err = x
		default:
			err = errors.New("Unknow panic")
		}
		log.Println("Recovered panic error : ", err)
	}
}
