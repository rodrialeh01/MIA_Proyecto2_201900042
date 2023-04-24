package analizador

import (
	"os"
	"strings"
)

type Rep struct {
	Name string
	Path string
	Id   string
	Ruta string
}

var consola_rep string

func (rep *Rep) VerificarParams(parametros map[string]string) {
	consola_rep = ""
	//Verificando parametros obligatorios
	if rep.Name == "" {
		consola_rep += "[-ERROR-] Falta el parametro name\n"
		return
	}
	if rep.Path == "" {
		consola_rep += "[-ERROR-] Falta el parametro path\n"
		return
	}
	if rep.Id == "" {
		consola_rep += "[-ERROR-] Falta el parametro id\n"
		return
	}

	//Verificando parametros opcionales
	if strings.ToLower(rep.Name) != "disk" && strings.ToLower(rep.Name) != "tree" && strings.ToLower(rep.Name) != "file" && strings.ToLower(rep.Name) != "sb" {
		consola_rep += "[-ERROR-] El parametro name no es valido\n"
		return
	}

	//Validando el parametro ruta
	if strings.ToLower(rep.Name) == "file" {
		if rep.Ruta == "" {
			consola_rep += "[-ERROR-] Falta el parametro ruta\n"
			return
		}
	}

	//Cambiando el path
	if rep.Path[0] == '"' && rep.Path[len(rep.Path)-1] == '"' {
		rep.Path = rep.Path[1 : len(rep.Path)-1]
	}

}

func (rep *Rep) ExisteDisco() bool {
	_, err := os.Stat(rep.Path)
	if os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}
