package analizador

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

type Fdisk struct {
	Size int
	Unit string
	Path string
	Type string
	Fit  string
	Name string
}

var consola_fdisk string

func (fdisk *Fdisk) VerificarParams(parametros map[string]string) {
	consola_fdisk = ""
	//Verificando parametros obligatorios
	if fdisk.Size == 0 {
		consola_fdisk += "[-ERROR-] Falta el parametro size\n"
		return
	}
	if fdisk.Size < 0 {
		consola_fdisk += "[-ERROR-] El parametro size no puede ser negativo\n"
		return
	}
	if fdisk.Path == "" {
		consola_fdisk += "[-ERROR-] Falta el parametro path\n"
		return
	}
	if fdisk.Name == "" {
		consola_fdisk += "[-ERROR-] Falta el parametro name\n"
		return
	}

	//Verificando parametros opcionales
	if fdisk.Fit == "" {
		fdisk.Fit = "ff"
	}
	if fdisk.Unit == "" {
		fdisk.Unit = "m"
	}
	if fdisk.Type == "" {
		fdisk.Type = "p"
	}
	fdisk.Fit = strings.ToLower(fdisk.Fit)
	fdisk.Unit = strings.ToLower(fdisk.Unit)
	fdisk.Type = strings.ToLower(fdisk.Type)
	if fdisk.Fit != "bf" && fdisk.Fit != "ff" && fdisk.Fit != "wf" {
		consola_fdisk += "[-ERROR-] El parametro fit no es valido\n"
		return
	}
	if fdisk.Unit != "m" && fdisk.Unit != "k" && fdisk.Unit != "b" {
		consola_fdisk += "[-ERROR-] El parametro unit no es valido\n"
		return
	}
	if fdisk.Type != "p" && fdisk.Type != "e" && fdisk.Type != "l" {
		consola_fdisk += "[-ERROR-] El parametro type no es valido\n"
		return
	}

	//Verificando si el disco existe
	if !fdisk.ExisteDisco() {
		consola_fdisk += "[-ERROR-] El disco no existe\n"
		return
	}

	//Verificando si la particion ya existe
	/*if fdisk.ExisteParticion() {
		consola_fdisk += "[-ERROR-] La particion ya existe\n"
		return
	}*/

	//Creando la particion
	fdisk.CrearParticion()
}

func (fdisk *Fdisk) CrearParticion() {
	//Lee el disco
	archivo, err := os.Open(fdisk.Path)
	if err != nil {
		consola_fdisk += "[-ERROR-] No se pudo leer el disco\n"
		return
	}
	defer archivo.Close()

	// Lee el MBR
	var mbr MBR
	tamanio := binary.Size(MBR{})
	fmt.Println("Tamaño del MBR: ", binary.Size(MBR{}))
	archivo.Seek(int64(tamanio), 0)
	err = binary.Read(archivo, binary.LittleEndian, &mbr)
	if err != nil {
		consola_fdisk += "[-ERROR-] No se pudo leer el MBR\n"
		return
	}

	//Verificando si se leyo bien el MBR
	fmt.Println("DESDE EL FDISK")
	fmt.Println("Fecha de creacion: ", string(mbr.mbr_fecha_creacion[:]))
	fmt.Println("Tamaño del disco: ", mbr.mbr_tamano)
	fmt.Println("Signature: ", mbr.mbr_dsk_signature)
}

func (fdisk *Fdisk) ExisteDisco() bool {
	_, err := os.Stat(fdisk.Path)
	if os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func RetornarConsolafdisk() string {
	return consola_fdisk
}
