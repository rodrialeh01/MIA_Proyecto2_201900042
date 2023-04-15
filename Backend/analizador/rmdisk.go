package analizador

import (
	"os"
)

type Rmdisk struct {
	Path string
}

var consola_rmdisk string

func (rmdisk *Rmdisk) VerificarParams(parametros map[string]string) {
	consola_rmdisk = ""
	//Verificando parametros obligatorios
	if rmdisk.Path == "" {
		consola_rmdisk += "[-ERROR-] Falta el parametro path\n"
		return
	}
	//Verificando si el disco existe
	if !rmdisk.ExisteDisco() {
		consola_rmdisk += "[-ERROR-] El disco no existe\n"
		return
	}
	//Eliminando el disco
	rmdisk.EliminarDisco()
}

func (rmdisk *Rmdisk) EliminarDisco() {
	err := os.Remove(rmdisk.Path)
	if err != nil {
		consola_rmdisk += "[-ERROR-] No se pudo eliminar el disco\n"
		return
	}
	consola_rmdisk += "[*SUCCESS*] Disco ha sido eliminado con exito\n"
}

func (rmdisk *Rmdisk) ExisteDisco() bool {
	_, err := os.Stat(rmdisk.Path)
	if os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func RetornarConsolarmdisk() string {
	return consola_rmdisk
}
