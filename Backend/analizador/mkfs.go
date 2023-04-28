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

func (mkfs *Mkfs) VerificarParams(parametros map[string]string) {
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

	if mkfs.Id[0] == '"' && mkfs.Id[len(mkfs.Id)-1] == '"' {
		mkfs.Id = mkfs.Id[1 : len(mkfs.Id)-1]
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
		consola_mkfs += "[-ERROR-] La partición con id: " + mkfs.Id + " no está montada\n"
		return
	}

	//Abrir el archivo binario
	archivo, err := os.OpenFile(montada.Path, os.O_RDWR, 0666)
	if err != nil {
		consola_mkfs += "[-ERROR-] Error al abrir el archivo\n"
		return
	}
	defer archivo.Close()

	//Leer el MBR
	mbr := MBR{}
	archivo.Seek(int64(0), 0)
	err = binary.Read(archivo, binary.LittleEndian, &mbr)
	if err != nil {
		consola_mkfs += "[-ERROR-] Error al leer el MBR\n"
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
				pos_inicio = int(particiones[i].Part_start)
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
	tiempoFormateado := tiempo.Format("2006-01-02 15:04:05")
	copy(super_bloque.S_mtime[:], tiempoFormateado)
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
	err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
	if err != nil {
		consola_mkfs += "[-ERROR-] Error al escribir el superbloque\n"
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
	//ESCRITURA DEL INNODO DE LA CARPETA RAIZ
	inodo_root := Inodo{}
	inodo_root.I_uid = 1
	inodo_root.I_gid = 1
	inodo_root.I_size = 0
	copy(inodo_root.I_atime[:], tiempoFormateado)
	copy(inodo_root.I_ctime[:], tiempoFormateado)
	copy(inodo_root.I_mtime[:], tiempoFormateado)
	for i := 0; i < len(inodo_root.I_block); i++ {
		inodo_root.I_block[i] = -1
	}
	inodo_root.I_block[0] = super_bloque.S_block_start
	inodo_root.I_type = 0
	inodo_root.I_perm = 664

	archivo.Seek(int64(inicio_tabla_inodos), 0)
	err = binary.Write(archivo, binary.LittleEndian, &inodo_root)
	if err != nil {
		consola_mkfs += "[-ERROR-] Error al escribir el inodo de la carpeta raiz\n"
		return
	}

	super_bloque.S_free_inodes_count--
	super_bloque.S_first_ino = super_bloque.S_first_ino + int32(binary.Size(Inodo{}))
	archivo.Seek(int64(pos_inicio), 0)
	err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
	if err != nil {
		consola_mkfs += "[-ERROR-] Error al actualizar el superbloque\n"
		return
	}
	fmt.Println("SUPERBLOQUE")
	fmt.Println(super_bloque.S_filesystem_type)
	fmt.Println(super_bloque.S_inodes_count)
	fmt.Println(super_bloque.S_blocks_count)
	fmt.Println(super_bloque.S_free_blocks_count)
	fmt.Println(super_bloque.S_free_inodes_count)
	fmt.Println(super_bloque.S_mtime)
	fmt.Println(super_bloque.S_mnt_count)
	fmt.Println(super_bloque.S_magic)
	fmt.Println(super_bloque.S_inode_size)
	fmt.Println(super_bloque.S_block_size)
	fmt.Println(super_bloque.S_first_ino)
	fmt.Println(super_bloque.S_first_blo)
	fmt.Println(super_bloque.S_bm_inode_start)
	fmt.Println(super_bloque.S_bm_block_start)
	fmt.Println(super_bloque.S_inode_start)
	fmt.Println(super_bloque.S_block_start)

	//ESCRITURA AL BITMAP DE INODOS
	var uno byte = 1
	archivo.Seek(int64(inicio_bitmap_inodos), 0)
	archivo.Write([]byte{uno})

	//ESCRITURA DEL BLOQUE DE LA CARPETA RAIZ
	bloque_root := Bloque_Carpeta{}
	for i := 0; i < len(bloque_root.B_content); i++ {
		bloque_root.B_content[i].B_inodo = -1
	}
	bloque_root.B_content[0].B_inodo = super_bloque.S_inode_start
	copy(bloque_root.B_content[0].B_name[:], "/")
	bloque_root.B_content[1].B_inodo = super_bloque.S_inode_start
	copy(bloque_root.B_content[1].B_name[:], "/")
	bloque_root.B_content[2].B_inodo = super_bloque.S_inode_start + int32(binary.Size(Inodo{}))
	copy(bloque_root.B_content[2].B_name[:], "users.txt")
	bloque_root.B_content[3].B_inodo = -1

	archivo.Seek(int64(inicio_bloques), 0)
	err = binary.Write(archivo, binary.LittleEndian, &bloque_root)
	if err != nil {
		consola_mkfs += "[-ERROR-] Error al escribir el bloque de la carpeta raiz\n"
		return
	}

	super_bloque.S_free_blocks_count--
	super_bloque.S_first_blo = super_bloque.S_first_blo + int32(binary.Size(Bloque_Carpeta{}))
	archivo.Seek(int64(pos_inicio), 0)
	err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
	if err != nil {
		consola_mkfs += "[-ERROR-] Error al actualizar el superbloque\n"
		return
	}

	//ESCRITURA AL BITMAP DE BLOQUES
	archivo.Seek(int64(inicio_bitmap_bloques), 0)
	archivo.Write([]byte{uno})

	//ESCRITURA DEL INODO USERS.TXT
	inodo_users := Inodo{}
	inodo_users.I_uid = 1
	inodo_users.I_gid = 1
	inodo_users.I_size = 27
	copy(inodo_users.I_atime[:], tiempoFormateado)
	copy(inodo_users.I_ctime[:], tiempoFormateado)
	copy(inodo_users.I_mtime[:], tiempoFormateado)
	for i := 0; i < len(inodo_users.I_block); i++ {
		inodo_users.I_block[i] = -1
	}
	inodo_users.I_block[0] = super_bloque.S_block_start + int32(binary.Size(Bloque_Carpeta{}))
	inodo_users.I_type = 1
	inodo_users.I_perm = 664

	archivo.Seek(int64(super_bloque.S_inode_start+int32(binary.Size(Inodo{}))), 0)
	err = binary.Write(archivo, binary.LittleEndian, &inodo_users)
	if err != nil {
		consola_mkfs += "[-ERROR-] Error al escribir el inodo del archivo users.txt\n"
		return
	}

	super_bloque.S_free_inodes_count--
	super_bloque.S_first_ino = super_bloque.S_first_ino + int32(binary.Size(Inodo{}))
	archivo.Seek(int64(pos_inicio), 0)
	err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
	if err != nil {
		consola_mkfs += "[-ERROR-] Error al actualizar el superbloque\n"
		return
	}

	//ESCRITURA AL BITMAP DE INODOS
	archivo.Seek(int64(inicio_bitmap_inodos+1), 0)
	archivo.Write([]byte{uno})

	//ESCRITURA DEL BLOQUE DEL ARCHIVO USERS.TXT
	bloque_users := Bloque_Archivo{}
	copy(bloque_users.B_content[:], "1,G,root\n1,U,root,root,123\n")
	archivo.Seek(int64(super_bloque.S_block_start+int32(binary.Size(Bloque_Carpeta{}))), 0)
	err = binary.Write(archivo, binary.LittleEndian, &bloque_users)
	if err != nil {
		consola_mkfs += "[-ERROR-] Error al escribir el bloque del archivo users.txt\n"
		return
	}

	super_bloque.S_free_blocks_count--
	super_bloque.S_first_blo = super_bloque.S_first_blo + int32(binary.Size(Bloque_Archivo{}))
	archivo.Seek(int64(pos_inicio), 0)
	err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
	if err != nil {
		consola_mkfs += "[-ERROR-] Error al actualizar el superbloque\n"
		return
	}

	//ESCRITURA AL BITMAP DE BLOQUES
	archivo.Seek(int64(inicio_bitmap_bloques+1), 0)
	archivo.Write([]byte{uno})

	//actualizo el mount
	for i := 0; i < len(ParticionesMontadasList); i++ {
		if montada == ParticionesMontadasList[i] {
			ParticionesMontadasList[i].Sistema_archivos = true
		}
	}

	//leer el sb
	archivo.Seek(int64(pos_inicio), 0)
	sbprueba := SuperBloque{}
	err1 := binary.Read(archivo, binary.LittleEndian, &sbprueba)
	if err1 != nil {
		fmt.Println("f")
	}

	fmt.Println("INICIO PARTICION SB: ", pos_inicio)
	fmt.Println("=============================================")
	fmt.Println("SUPERBLOQUE")
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

	consola_mkfs += "[*SUCCESS*] Se acaba de formatear la partición y se creó el sistema de archivos EXT2\n"

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
