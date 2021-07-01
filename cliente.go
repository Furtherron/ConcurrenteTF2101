package main

import (
	"encoding/gob"
	"fmt"
	"net"
)

type Parametros struct {
	KNearest string
	GRUPO    string
	EDAD     string
	SEXO     string
	DOSIS    string
	UBIGEO   string
	Eleccion string
	//RESULTADO sortedClassVotes
}

func cliente(parametros Parametros) {
	c, err := net.Dial("tcp", ":9999")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = gob.NewEncoder(c).Encode(parametros)
	if err != nil {
		fmt.Println(err)
	}

	c.Close()
}

func main() {
	parametros := Parametros{
		KNearest: "1",
		GRUPO:    "1",
		EDAD:     "1",
		SEXO:     "1",
		DOSIS:    "1",
		UBIGEO:   "1",
		Eleccion: "1",
	}
	go cliente(parametros)
	fmt.Println(parametros)
}
