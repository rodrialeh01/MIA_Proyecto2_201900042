package analizador

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
)

type Rmusr struct {
	User string
}

var consola_rmusr string

func (rmusr *Rmusr) VerificarParams(parametros map[string]string) {
	//Verificando parametros obligatorios
	consola_rmusr = ""
	if rmusr.User == "" {
		consola_rmusr += "[-ERROR-] Falta el parametro user\n"
		return
	}
	if rmusr.User[0] == '"' && rmusr.User[len(rmusr.User)-1] == '"' {
		rmusr.User = rmusr.User[1 : len(rmusr.User)-1]
	}

	rmusr.EliminarUsuario()
}

func (rmusr *Rmusr) EliminarUsuario() {
	montada := rmusr.RetornarStrictMontada(Idlogueado)
	if rmusr.IsParticionMontadaVacia(montada) {
		consola_rmusr += "[-ERROR-] La partición con id: " + Idlogueado + " no está montada\n"
		return
	}

	if !montada.Sistema_archivos {
		consola_rmusr += "[-ERROR-] La partición con id: " + Idlogueado + " no tiene un sistema de archivos\n"
	}

	if montada.User != "root" {
		consola_rmusr += "[-ERROR-] No se tienen los permisos suficientes para crear un grupo\n"
		return
	}

	//Abrir el archivo binario
	archivo, err := os.OpenFile(montada.Path, os.O_RDWR, 0666)
	if err != nil {
		consola_rmusr += "[-ERROR-] Error al abrir el archivo\n"
		return
	}
	defer archivo.Close()

	//Leer el MBR
	mbr := MBR{}
	archivo.Seek(int64(0), 0)
	err = binary.Read(archivo, binary.LittleEndian, &mbr)
	if err != nil {
		consola_rmusr += "[-ERROR-] Error al leer el MBR\n"
		return
	}
	fmt.Println("MBR DESDE REP")
	fmt.Println(mbr)
	particiones := rmusr.ObtenerParticiones(mbr)
	var ebrs []EBR
	logica_existe := false
	var particion_logica EBR
	for i := 0; i < len(particiones); i++ {
		if strings.ToLower(string(particiones[i].Part_type[0])) == "e" {
			ebrs = rmusr.ListadoEBR(particiones[i], montada.Path)
			for j := 0; j < len(ebrs); j++ {
				if strings.Contains(strings.ToLower(string(ebrs[j].Part_name[:])), strings.ToLower(montada.Name)) {
					particion_logica = ebrs[j]
					logica_existe = true
				}
			}
		}
	}

	pos_inicio := 0
	if logica_existe {
		pos_inicio = int(particion_logica.Part_start) + binary.Size(EBR{})
		//pos_final = int(particion_logica.Part_start) + int(particion_logica.Part_size)
	} else {
		for i := 0; i < len(particiones); i++ {
			if strings.Contains(strings.ToLower(string(particiones[i].Part_name[:])), strings.ToLower(montada.Name)) {
				pos_inicio = int(particiones[i].Part_start)
				//pos_final = int(particiones[i].Part_start) + int(particiones[i].Part_size)
				break
			}
		}
	}

	//Leer el SuperBloque
	archivo.Seek(int64(pos_inicio), 0)
	super_bloque := SuperBloque{}
	err = binary.Read(archivo, binary.LittleEndian, &super_bloque)
	if err != nil {
		consola_rmusr += "[-ERROR-] Error al leer el SuperBloque\n"
		return
	}

	fmt.Println("SUPERBLOQUE DESDE MKGRP")
	sbprueba := super_bloque
	fmt.Println("S_filesystem_type : ", sbprueba.S_filesystem_type)
	fmt.Println("S_inodes_count: ", sbprueba.S_inodes_count)
	fmt.Println("S_blocks_count:", sbprueba.S_blocks_count)
	fmt.Println("S_free_blocks_count:", sbprueba.S_free_blocks_count)
	fmt.Println("S_free_inodes_count:", sbprueba.S_free_inodes_count)
	fmt.Println("S_mtime:", sbprueba.S_mtime)
	fmt.Println("S_mnt_count:", sbprueba.S_mnt_count)
	fmt.Println("S_magic:", sbprueba.S_magic)
	fmt.Println("S_block_size:", sbprueba.S_block_size)
	fmt.Println("S_first_ino:", sbprueba.S_first_ino)
	fmt.Println("S_first_blo:", sbprueba.S_first_blo)
	fmt.Println("S_bm_inode_start:", sbprueba.S_bm_inode_start)
	fmt.Println("S_bm_block_start:", sbprueba.S_bm_block_start)
	fmt.Println("S_inode_start:", sbprueba.S_inode_start)
	fmt.Println("S_block_start:", sbprueba.S_block_start)

	//Leer el Inodo del archivo users.txt
	inodo_users := Inodo{}
	pos_inodo := super_bloque.S_inode_start + int32(binary.Size(Inodo{}))
	archivo.Seek(int64(pos_inodo), 0)
	err = binary.Read(archivo, binary.LittleEndian, &inodo_users)
	if err != nil {
		consola_rmusr += "[-ERROR-] Error al leer el Inodo del archivo users.txt\n"
		return
	}

	usuariostxt := ""
	//Leer el Bloque de datos del archivo users.txt
	for i := 0; i < len(inodo_users.I_block); i++ {
		no_bloque := int(inodo_users.I_block[i])
		if no_bloque == -1 {
			break
		}
		bloque := Bloque_Archivo{}
		archivo.Seek(int64(no_bloque), 0)
		err = binary.Read(archivo, binary.LittleEndian, &bloque)
		if err != nil {
			consola_rmusr += "[-ERROR-] Error al leer el Bloque de datos del archivo users.txt\n"
			return
		}
		usuariostxt += string(bloque.B_content[:])
	}

	usuariostxt = strings.Replace(usuariostxt, "\u0000", "", -1)
	fmt.Println("=====================")
	fmt.Println(usuariostxt)

	nuevo_usuariostxt := ""
	encontrado := false
	usuarios_grupos := strings.Split(usuariostxt, "\n")
	for i := 0; i < len(usuarios_grupos); i++ {
		datos := strings.Split(usuarios_grupos[i], ",")
		if len(datos) == 5 {
			if strings.Contains(datos[1], "U") {
				if strings.Contains(datos[3], rmusr.User) {
					if strings.Contains(datos[0], "0") {
						consola_rmusr += "[-ERROR-] Ya ha sido eliminado el usuario: " + rmusr.User + "\n"
						return
					} else {
						nuevo_usuariostxt += "0," + datos[1] + "," + datos[2] + "," + datos[3] + "," + datos[4] + "\n"
						encontrado = true
					}
				} else {
					nuevo_usuariostxt += usuarios_grupos[i] + "\n"
				}
			}
		} else if len(datos) > 0 {
			nuevo_usuariostxt += usuarios_grupos[i] + "\n"
		}
	}

	if !encontrado {
		consola_rmusr += "[-ERROR-] No se encontro el usuario: " + rmusr.User + "\n"
		return
	}

	fmt.Println(nuevo_usuariostxt)
	insertar := nuevo_usuariostxt
	//ESCRIBIR EL CONTENIDO DEL ARCHIVO
	for i := 0; i < len(inodo_users.I_block); i++ {
		no_bloque := int(inodo_users.I_block[i])
		if no_bloque == -1 {
			break
		}
		if len(insertar) > 63 {
			bloque := Bloque_Archivo{}
			copy(bloque.B_content[:], insertar[:63])
			insertar = insertar[63:]
			archivo.Seek(int64(no_bloque), 0)
			err = binary.Write(archivo, binary.LittleEndian, bloque)
			if err != nil {
				consola_rmusr += "[-ERROR-] Error al escribir el Bloque de datos del archivo users.txt\n"
				return
			}
		} else {
			bloque := Bloque_Archivo{}
			copy(bloque.B_content[:], insertar)
			archivo.Seek(int64(no_bloque), 0)
			err = binary.Write(archivo, binary.LittleEndian, bloque)
			if err != nil {
				consola_rmusr += "[-ERROR-] Error al escribir el Bloque de datos del archivo users.txt\n"
				return
			}
		}
	}

	tiempo := time.Now()
	tiempoFormateado := tiempo.Format("2006-01-02 15:04:05")
	copy(inodo_users.I_mtime[:], tiempoFormateado)
	archivo.Seek(int64(pos_inodo), 0)
	err = binary.Write(archivo, binary.LittleEndian, &inodo_users)
	if err != nil {
		consola_rmusr += "[-ERROR-] Error al escribir el Inodo del archivo users.txt\n"
		return
	}

	consola_rmusr += "[*SUCCESS*] Se ha eliminado el usuario: " + rmusr.User + " exitosamente\n"
}

func (rmusr *Rmusr) CadenaVacia(cadena [16]byte) bool {

	for _, v := range cadena {
		if v != 0 {
			return false
		}
	}
	return true

}

func (rmusr *Rmusr) VerificarID() bool {
	//Verificando si existe el id
	for i := 0; i < len(ParticionesMontadasList); i++ {
		if strings.Contains(strings.ToLower(ParticionesMontadasList[i].Id), strings.ToLower(Idlogueado)) {
			return true
		}
	}
	return false
}

func (rmusr *Rmusr) RetornarStrictMontada(id string) ParticionMontada {
	for i := 0; i < len(ParticionesMontadasList); i++ {
		if strings.Contains(strings.ToLower(ParticionesMontadasList[i].Id), strings.ToLower(id)) {
			return ParticionesMontadasList[i]
		}
	}
	return ParticionMontada{}
}

func (rmusr *Rmusr) IsParticionMontadaVacia(p ParticionMontada) bool {
	return !p.Sistema_archivos && p.Id == "" && p.Letra == "" && p.Numero == 0 && p.Path == "" && p.Type == "" && p.Name == ""
}

func (rmusr *Rmusr) ObtenerParticiones(Mbr MBR) []Partition {
	var particiones []Partition
	particiones = append(particiones, Mbr.Mbr_partition_1)
	particiones = append(particiones, Mbr.Mbr_partition_2)
	particiones = append(particiones, Mbr.Mbr_partition_3)
	particiones = append(particiones, Mbr.Mbr_partition_4)
	return particiones
}

func (rmusr *Rmusr) ListadoEBR(Extendida Partition, path string) []EBR {
	ebrs := []EBR{}
	archivox, _ := os.OpenFile(path, os.O_RDWR, 0666)
	defer archivox.Close()

	temp := Extendida.Part_start
	for temp != -1 {
		archivox.Seek(int64(temp), 0)
		ebr := EBR{}
		err := binary.Read(archivox, binary.LittleEndian, &ebr)
		if err != nil {
			return ebrs
		}
		if !rmusr.CadenaVacia(ebr.Part_name) {
			ebrs = append(ebrs, ebr)
		} else if rmusr.CadenaVacia(ebr.Part_name) && ebr.Part_size != 0 {
			ebrs = append(ebrs, ebr)
		} else {
			break
		}
		temp = ebr.Part_next
	}
	return ebrs
}

func RetornarConsolarmusr() string {
	return consola_rmusr
}
