package analizador

import (
	"os"
	"strings"
)

type MkDisk struct {
	Path string
	Size int
	Unit string
	Fit  string
}

var consola_mkdisk string

func (mkdisk *MkDisk) VerificarParams(parametros map[string]string) {
	//Verificando parametros obligatorios
	if mkdisk.Path == "" {
		consola_mkdisk += "[-ERROR-] Falta el parametro path\n"
		return
	}
	if mkdisk.Size == 0 {
		consola_mkdisk += "[-ERROR-] Falta el parametro size\n"
		return
	}
	if mkdisk.Size < 0 {
		consola_mkdisk += "[-ERROR-] El parametro size no puede ser negativo\n"
		return
	}

	//Verificando parametros opcionales
	if mkdisk.Fit == "" {
		mkdisk.Fit = "ff"
	}
	if mkdisk.Unit == "" {
		mkdisk.Unit = "m"
	}
	mkdisk.Fit = strings.ToLower(mkdisk.Fit)
	mkdisk.Unit = strings.ToLower(mkdisk.Unit)
	if mkdisk.Fit != "bf" && mkdisk.Fit != "ff" && mkdisk.Fit != "wf" {
		consola_mkdisk += "[-ERROR-] El parametro fit no es valido\n"
		return
	}
	if mkdisk.Unit != "m" && mkdisk.Unit != "k" {
		consola_mkdisk += "[-ERROR-] El parametro unit no es valido\n"
		return
	}

	//Verificando si el disco ya existe
	if mkdisk.ExisteDisco() {
		consola_mkdisk += "[-ERROR-] El disco ya existe\n"
		return
	}

	//Cambiando el path
	if mkdisk.Path[0] == '"' && mkdisk.Path[len(mkdisk.Path)-1] == '"' {
		mkdisk.Path = mkdisk.Path[1 : len(mkdisk.Path)-1]
	}

	//Crear disco
	mkdisk.CrearDisco()
}

func (mkdisk *MkDisk) CrearDisco() {
	//Crear disco
}

func (mkdisk *MkDisk) ExisteDisco() bool {
	archivo := mkdisk.Path
	_, err := os.Stat(archivo)
	return err == nil
}

func RetornarConsolamkdisk() string {
	return consola_mkdisk
}
