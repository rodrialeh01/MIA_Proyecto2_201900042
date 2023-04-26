package analizador

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"strings"
	"time"
)

type Mkfs struct {
	Id   string
	Type string
}

var consola_mkfs string

func (mkfs *Mkfs) VerificarParams() {
	consola_mkfs = ""
	// Verificar la escritura de parametros
	if mkfs.Id == "" {
		consola_mkfs += "[-ERROR-] No agregó el parametro id\n"
		return
	}

	if mkfs.Type == "" {
		mkfs.Type = "full"
	} else {
		if strings.ToLower(mkfs.Type) != "full" {
			consola_mkfs += "[-ERROR-] El parametro type solo puede ser full\n"
			return
		}
	}

	if !mkfs.VerificarID() {
		consola_mkfs += "[-ERROR-] No existe la particion con el id: " + mkfs.Id + "\n"
		return
	}

	mkfs.FormateoEXT2()
}

func (mkfs *Mkfs) FormateoEXT2() {
	montada := mkfs.RetornarStrictMontada(mkfs.Id)
	if mkfs.IsParticionMontadaVacia(montada) {
		consola_rep += "[-ERROR-] La partición con id: " + mkfs.Id + " no está montada\n"
		return
	}

	//Abrir el archivo binario
	archivo, err := os.OpenFile(montada.Path, os.O_RDWR, 0666)
	if err != nil {
		consola_rep += "[-ERROR-] Error al abrir el archivo\n"
		return
	}
	defer archivo.Close()

	//Leer el MBR
	mbr := MBR{}
	archivo.Seek(int64(0), 0)
	err = binary.Read(archivo, binary.LittleEndian, &mbr)
	if err != nil {
		consola_rep += "[-ERROR-] Error al leer el MBR\n"
		return
	}
	fmt.Println("MBR DESDE REP")
	fmt.Println(mbr)
	particiones := mkfs.ObtenerParticiones(mbr)
	var ebrs []EBR
	logica_existe := false
	var particion_logica EBR
	for i := 0; i < len(particiones); i++ {
		if strings.ToLower(string(particiones[i].Part_type[0])) == "e" {
			ebrs = mkfs.ListadoEBR(particiones[i], montada.Path)
			for j := 0; j < len(ebrs); j++ {
				if strings.Contains(strings.ToLower(string(ebrs[j].Part_name[:])), strings.ToLower(montada.Name)) {
					particion_logica = ebrs[j]
					logica_existe = true
				}
			}
		}
	}

	pos_inicio := 0
	pos_final := 0
	if logica_existe {
		pos_inicio = int(particion_logica.Part_start) + binary.Size(EBR{})
		pos_final = int(particion_logica.Part_start) + int(particion_logica.Part_size)
	} else {
		for i := 0; i < len(particiones); i++ {
			if strings.Contains(strings.ToLower(string(particiones[i].Part_name[:])), strings.ToLower(montada.Name)) {
				pos_inicio = int(particiones[i].Part_start) + binary.Size(EBR{})
				pos_final = int(particiones[i].Part_start) + int(particiones[i].Part_size)
				break
			}
		}
	}

	//VACIA LA PARTICION
	archivo.Seek(int64(pos_inicio), 0)
	for i := pos_inicio; i < pos_final; i++ {
		archivo.Write([]byte{0})
	}

	//CALCULO DE N
	//FORMULA: n = (tamanio_particion - sizeof(superbloque))/(4+sizeof(inodos) + 3*sizeof(block))
	tamanio_particion := pos_final - pos_inicio
	n := float64(tamanio_particion-binary.Size(SuperBloque{})) / float64(4+binary.Size(Inodo{})+3*binary.Size(Bloque_Carpeta{}))
	n = math.Floor(n)

	tamanio_bitmap_inodos := int(n)
	tamanio_bitmap_bloques := int(n * 3)
	tamanio_inodos := int(n) * binary.Size(Inodo{})

	//CALCULO DE INICIO DE CADA SECCION
	inicio_bitmap_inodos := pos_inicio + binary.Size(SuperBloque{})
	inicio_bitmap_bloques := inicio_bitmap_inodos + tamanio_bitmap_inodos
	inicio_tabla_inodos := inicio_bitmap_bloques + tamanio_bitmap_bloques
	inicio_bloques := inicio_tabla_inodos + tamanio_inodos

	//CREACION DE SUPERBLOQUE
	super_bloque := SuperBloque{}
	super_bloque.S_filesystem_type = 2
	super_bloque.S_inodes_count = int32(tamanio_bitmap_inodos)
	super_bloque.S_blocks_count = int32(tamanio_bitmap_bloques)
	super_bloque.S_free_blocks_count = int32(tamanio_bitmap_bloques)
	super_bloque.S_free_inodes_count = int32(tamanio_bitmap_inodos)
	tiempo := time.Now()
	tiempoS := tiempo.String()
	copy(super_bloque.S_mtime[:], tiempoS)
	super_bloque.S_mnt_count = 1
	super_bloque.S_magic = 0xEF53
	super_bloque.S_inode_size = int32(binary.Size(Inodo{}))
	super_bloque.S_block_size = int32(binary.Size(Bloque_Carpeta{}))
	super_bloque.S_first_ino = int32(inicio_tabla_inodos)
	super_bloque.S_first_blo = int32(inicio_bloques)
	super_bloque.S_bm_inode_start = int32(inicio_bitmap_inodos)
	super_bloque.S_bm_block_start = int32(inicio_bitmap_bloques)
	super_bloque.S_inode_start = int32(inicio_tabla_inodos)
	super_bloque.S_block_start = int32(inicio_bloques)

	//ESCRIBIENDO SUPERBLOQUE
	archivo.Seek(int64(pos_inicio), 0)
	err = binary.Write(archivo, binary.LittleEndian, super_bloque)
	if err != nil {
		consola_rep += "[-ERROR-] Error al escribir el superbloque\n"
		return
	}

	//CREACION DE BITMAP DE INODOS
	bitmap_i := make([]byte, tamanio_bitmap_inodos)
	for i := 0; i < len(bitmap_i); i++ {
		bitmap_i[i] = 0
	}
	archivo.Seek(int64(inicio_bitmap_inodos), 0)
	archivo.Write(bitmap_i)

	//CREACION DE BITMAP DE BLOQUES
	bitmap_b := make([]byte, tamanio_bitmap_bloques)
	for i := 0; i < len(bitmap_b); i++ {
		bitmap_b[i] = 0
	}
	archivo.Seek(int64(inicio_bitmap_bloques), 0)
	archivo.Write(bitmap_b)

	//========================================CREACION DEL /USERS.TXT =====================================

}

func (mkfs *Mkfs) CadenaVacia(cadena [16]byte) bool {

	for _, v := range cadena {
		if v != 0 {
			return false
		}
	}
	return true

}

func (mkfs *Mkfs) VerificarID() bool {
	//Verificando si existe el id
	for i := 0; i < len(ParticionesMontadasList); i++ {
		if strings.Contains(strings.ToLower(ParticionesMontadasList[i].Id), strings.ToLower(mkfs.Id)) {
			return true
		}
	}
	return false
}

func (mkfs *Mkfs) RetornarStrictMontada(id string) ParticionMontada {
	for i := 0; i < len(ParticionesMontadasList); i++ {
		if strings.Contains(strings.ToLower(ParticionesMontadasList[i].Id), strings.ToLower(id)) {
			return ParticionesMontadasList[i]
		}
	}
	return ParticionMontada{}
}

func (mkfs *Mkfs) IsParticionMontadaVacia(p ParticionMontada) bool {
	return !p.Sistema_archivos && p.Id == "" && p.Letra == "" && p.Numero == 0 && p.Path == "" && p.Type == "" && p.Name == ""
}

func (mkfs *Mkfs) ObtenerParticiones(Mbr MBR) []Partition {
	var particiones []Partition
	particiones = append(particiones, Mbr.Mbr_partition_1)
	particiones = append(particiones, Mbr.Mbr_partition_2)
	particiones = append(particiones, Mbr.Mbr_partition_3)
	particiones = append(particiones, Mbr.Mbr_partition_4)
	return particiones
}

func (mkfs *Mkfs) ListadoEBR(Extendida Partition, path string) []EBR {
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
		if !mkfs.CadenaVacia(ebr.Part_name) {
			ebrs = append(ebrs, ebr)
		} else if mkfs.CadenaVacia(ebr.Part_name) && ebr.Part_size != 0 {
			ebrs = append(ebrs, ebr)
		} else {
			break
		}
		temp = ebr.Part_next
	}
	return ebrs
}

func RetornarConsolamkfs() string {
	return consola_mkfs
}
