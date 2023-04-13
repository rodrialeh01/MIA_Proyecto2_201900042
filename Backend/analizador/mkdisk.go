package analizador

import (
	"bytes"
	"encoding/binary"
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

	//CERO DEFAULT
	var temp int8 = 0
	s := &temp
	var binario bytes.Buffer
	binary.Write(&binario, binary.BigEndian, s)

	//Crea el tamaño del disco
	if mkdisk.Unit == "k" {
		size := int32(mkdisk.Size * 1024)
		bytes_size := make([]byte, 4)
		binary.PutVarint(bytes_size, int64(size))
		copy(mbr.mbr_tamano[:], bytes_size)
		for i := 0; i < (mkdisk.Size * 1024); i++ {
			err = binary.Write(archivo, binary.LittleEndian, byte('0'))
			if err != nil {
				consola_mkdisk += "[-ERROR-] No se pudo crear el disco\n"
				return
			}
		}
	} else if mkdisk.Unit == "m" {
		size := int32(mkdisk.Size * 1024 * 1024)
		bytes_size := make([]byte, 4)
		binary.PutVarint(bytes_size, int64(size))
		copy(mbr.mbr_tamano[:], bytes_size)
		for i := 0; i < (mkdisk.Size * 1024 * 1024); i++ {
			err = binary.Write(archivo, binary.LittleEndian, byte('0'))
			if err != nil {
				consola_mkdisk += "[-ERROR-] No se pudo crear el disco\n"
				return
			}
		}
	}

	//Agrega la fecha de creacion
	tiempo := time.Now()
	tiempoS := tiempo.String()

	copy(mbr.mbr_fecha_creacion[:], tiempoS)

	//Agrega signature
	signature := rand.Intn(999999999) + 1
	entero := int32(signature)
	bytes := make([]byte, 4)
	binary.PutVarint(bytes, int64(entero))
	copy(mbr.mbr_dsk_signature[:], bytes)

	//Agrega el fit
	if mkdisk.Fit == "bf" {
		tipo := []byte{byte('B')}
		copy(mbr.mbr_fit[:], tipo)
	} else if mkdisk.Fit == "ff" {
		tipo := []byte{byte('F')}
		copy(mbr.mbr_fit[:], tipo)
	} else if mkdisk.Fit == "wf" {
		tipo := []byte{byte('W')}
		copy(mbr.mbr_fit[:], tipo)
	}

	//Inicializa las particiones

	status := []byte{byte('0')}
	copy(mbr.mbr_partition_1.part_status[:], status)
	copy(mbr.mbr_partition_2.part_status[:], status)
	copy(mbr.mbr_partition_3.part_status[:], status)
	copy(mbr.mbr_partition_4.part_status[:], status)

	type_mbr := []byte{byte('0')}
	copy(mbr.mbr_partition_1.part_type[:], type_mbr)
	copy(mbr.mbr_partition_2.part_type[:], type_mbr)
	copy(mbr.mbr_partition_3.part_type[:], type_mbr)
	copy(mbr.mbr_partition_4.part_type[:], type_mbr)

	fit := []byte{byte('0')}
	copy(mbr.mbr_partition_1.part_fit[:], fit)
	copy(mbr.mbr_partition_2.part_fit[:], fit)
	copy(mbr.mbr_partition_3.part_fit[:], fit)
	copy(mbr.mbr_partition_4.part_fit[:], fit)

	start_p := make([]byte, 8)
	binary.PutVarint(start_p, int64(-1))
	copy(mbr.mbr_partition_1.part_start[:], start_p)
	copy(mbr.mbr_partition_2.part_start[:], start_p)
	copy(mbr.mbr_partition_3.part_start[:], start_p)
	copy(mbr.mbr_partition_4.part_start[:], start_p)

	part_name := []byte{byte('0')}
	copy(mbr.mbr_partition_1.part_name[:], part_name)
	copy(mbr.mbr_partition_2.part_name[:], part_name)
	copy(mbr.mbr_partition_3.part_name[:], part_name)
	copy(mbr.mbr_partition_4.part_name[:], part_name)

	//Escribir el MBR
	posicion := int64(0)
	_, err = archivo.Seek(posicion, 0)
	if err != nil {
		consola_mkdisk += "[-ERROR-] No se pudo crear el disco\n"
		return
	}

	mbr_byte := make([]byte, binary.Size(mbr))
	_, err = archivo.Write(mbr_byte)
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
