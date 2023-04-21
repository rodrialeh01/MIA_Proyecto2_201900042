package analizador

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

type MkDisk struct {
	Path string
	Size int
	Unit string
	Fit  string
}

var consola_mkdisk string

func (mkdisk *MkDisk) VerificarParams(parametros map[string]string) {
	consola_mkdisk = ""
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
	//Crear las carpetas
	carpetas := obtener_path_carpetas(mkdisk.Path)
	if _, err := os.Stat(carpetas); os.IsNotExist(err) {
		err = os.MkdirAll(carpetas, 0777)
		if err != nil {
			consola_mkdisk += "[-ERROR-] No se pudo crear las carpetas del disco\n"
			return
		}
	}

	//Crear el archivo
	archivo, err := os.Create(mkdisk.Path)
	if err != nil {
		consola_mkdisk += "[-ERROR-] No se pudo crear el disco\n"
		return
	}
	defer archivo.Close()

	//Crear el MBR
	mbr := MBR{}
	fmt.Println("=================MBR=================")
	//Crea el tamaño del disco
	if mkdisk.Unit == "k" {
		mbr.Mbr_tamano = int32(mkdisk.Size * 1024)
		for i := 0; i < (mkdisk.Size * 1024); i++ {
			err = binary.Write(archivo, binary.LittleEndian, byte('0'))
			if err != nil {
				consola_mkdisk += "[-ERROR-] No se pudo crear el disco\n"
				return
			}
		}
	} else if mkdisk.Unit == "m" {
		mbr.Mbr_tamano = int32(mkdisk.Size * 1024 * 1024)
		for i := 0; i < (mkdisk.Size * 1024 * 1024); i++ {
			err = binary.Write(archivo, binary.LittleEndian, byte('0'))
			if err != nil {
				consola_mkdisk += "[-ERROR-] No se pudo crear el disco\n"
				return
			}
		}
	}
	fmt.Println("Tamaño del disco: ", mbr.Mbr_tamano)
	//Agrega la fecha de creacion
	tiempo := time.Now()
	tiempoS := tiempo.String()

	copy(mbr.Mbr_fecha_creacion[:], tiempoS)
	fmt.Println("Fecha de creacion: ", string(mbr.Mbr_fecha_creacion[:]))
	//Agrega signature
	signature := rand.Intn(999999999) + 1
	mbr.Mbr_dsk_signature = int32(signature)
	fmt.Println("Signature: ", mbr.Mbr_dsk_signature)
	//Agrega el fit
	if mkdisk.Fit == "bf" {
		tipo := []byte{byte('B')}
		copy(mbr.Mbr_fit[:], tipo)
	} else if mkdisk.Fit == "ff" {
		tipo := []byte{byte('F')}
		copy(mbr.Mbr_fit[:], tipo)
	} else if mkdisk.Fit == "wf" {
		tipo := []byte{byte('W')}
		copy(mbr.Mbr_fit[:], tipo)
	}

	//Inicializa las particiones

	status := []byte{byte('0')}
	copy(mbr.Mbr_partition_1.Part_status[:], status)
	copy(mbr.Mbr_partition_2.Part_status[:], status)
	copy(mbr.Mbr_partition_3.Part_status[:], status)
	copy(mbr.Mbr_partition_4.Part_status[:], status)

	type_mbr := []byte{byte('0')}
	copy(mbr.Mbr_partition_1.Part_type[:], type_mbr)
	copy(mbr.Mbr_partition_2.Part_type[:], type_mbr)
	copy(mbr.Mbr_partition_3.Part_type[:], type_mbr)
	copy(mbr.Mbr_partition_4.Part_type[:], type_mbr)

	fit := []byte{byte('0')}
	copy(mbr.Mbr_partition_1.Part_fit[:], fit)
	copy(mbr.Mbr_partition_2.Part_fit[:], fit)
	copy(mbr.Mbr_partition_3.Part_fit[:], fit)
	copy(mbr.Mbr_partition_4.Part_fit[:], fit)

	start_p := int32(-1)
	mbr.Mbr_partition_1.Part_start = start_p
	mbr.Mbr_partition_2.Part_start = start_p
	mbr.Mbr_partition_3.Part_start = start_p
	mbr.Mbr_partition_4.Part_start = start_p

	part_name := []byte{byte('0')}
	copy(mbr.Mbr_partition_1.Part_name[:], part_name)
	copy(mbr.Mbr_partition_2.Part_name[:], part_name)
	copy(mbr.Mbr_partition_3.Part_name[:], part_name)
	copy(mbr.Mbr_partition_4.Part_name[:], part_name)

	//Escribir el MBR
	archivo.Seek(0, 0)
	err = binary.Write(archivo, binary.LittleEndian, &mbr)
	if err != nil {
		consola_mkdisk += "[-ERROR-] No se pudo crear el disco\n"
		return
	}

	consola_mkdisk += "[*SUCCESS*] Disco creado con exito\n"

}

func obtener_path_carpetas(path string) string {
	var aux_path int
	var aux_ruta string

	for i := len(path) - 1; i >= 0; i-- {
		aux_path++
		if string(path[i]) == "/" {
			break
		}
	}

	for i := 0; i < ((len(path)) - (aux_path - 1)); i++ {
		aux_ruta += string(path[i])
	}
	return aux_ruta

}

func (mkdisk *MkDisk) ExisteDisco() bool {
	_, err := os.Stat(mkdisk.Path)
	if os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func RetornarConsolamkdisk() string {
	return consola_mkdisk
}
