package analizador

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
)

type Mkdir struct {
	Path string
	R    bool
}

var consola_mkdir string

func (mkdir *Mkdir) VerificarParams(parametros map[string]string) {
	//Verificando parametros obligatorios
	consola_mkdir = ""
	if mkdir.Path == "" {
		consola_mkdir += "[-ERROR-] Falta el parametro path\n"
		return
	}
	if mkdir.Path[0] == '"' {
		mkdir.Path = mkdir.Path[1 : len(mkdir.Path)-1]
	}

	if Idlogueado == "" {
		consola_mkdir += "[-ERROR-] No hay ninguna sesión iniciada\n"
		return
	}

	mkdir.CrearCarpetas()
}

func (mkdir *Mkdir) CrearCarpetas() {
	montada := mkdir.RetornarStrictMontada(Idlogueado)
	if mkdir.IsParticionMontadaVacia(montada) {
		consola_mkdir += "[-ERROR-] La partición con id: " + Idlogueado + " no está montada\n"
		return
	}

	if !montada.Sistema_archivos {
		consola_mkdir += "[-ERROR-] La partición con id: " + Idlogueado + " no tiene un sistema de archivos\n"
	}

	if montada.User != "root" {
		consola_mkdir += "[-ERROR-] No se tienen los permisos suficientes para crear un grupo\n"
		return
	}

	//Abrir el archivo binario
	archivo, err := os.OpenFile(montada.Path, os.O_RDWR, 0666)
	if err != nil {
		consola_mkdir += "[-ERROR-] Error al abrir el archivo\n"
		return
	}
	defer archivo.Close()

	//Leer el MBR
	mbr := MBR{}
	archivo.Seek(int64(0), 0)
	err = binary.Read(archivo, binary.LittleEndian, &mbr)
	if err != nil {
		consola_mkdir += "[-ERROR-] Error al leer el MBR\n"
		return
	}
	fmt.Println("MBR DESDE REP")
	fmt.Println(mbr)
	particiones := mkdir.ObtenerParticiones(mbr)
	var ebrs []EBR
	logica_existe := false
	var particion_logica EBR
	for i := 0; i < len(particiones); i++ {
		if strings.ToLower(string(particiones[i].Part_type[0])) == "e" {
			ebrs = mkdir.ListadoEBR(particiones[i], montada.Path)
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
		consola_mkdir += "[-ERROR-] Error al leer el SuperBloque\n"
		return
	}

	if mkdir.R {
		mkdir.Recursivo(pos_inicio, montada.Path)
	} else {
		mkdir.NoRecursivo(pos_inicio, montada.Path)
	}
}

func (mkdir *Mkdir) Recursivo(pos_sb int, path string) {
	archivo, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		consola_mkdir += "[-ERROR-] Error al abrir el archivo\n"
		return
	}
	defer archivo.Close()

}

func (mkdir *Mkdir) NoRecursivo(pos_sb int, path string) {
	archivo, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		consola_mkdir += "[-ERROR-] Error al abrir el archivo\n"
		return
	}
	defer archivo.Close()

	//Leer el SuperBloque
	archivo.Seek(int64(pos_sb), 0)
	super_bloque := SuperBloque{}
	err = binary.Read(archivo, binary.LittleEndian, &super_bloque)
	if err != nil {
		consola_mkdir += "[-ERROR-] Error al leer el SuperBloque\n"
		return
	}

	//Leer el inodo de la carpeta raiz
	inodo := Inodo{}
	archivo.Seek(int64(super_bloque.S_inode_start), 0)
	err = binary.Read(archivo, binary.LittleEndian, &inodo)
	if err != nil {
		consola_mkdir += "[-ERROR-] Error al leer el Inodo\n"
		return
	}

	//Separar los nombres de las carpetas
	name_carpetas := strings.Split(mkdir.Path, "/")
	name_carpetas[0] = "/"
	nueva_carpeta := name_carpetas[len(name_carpetas)-1]
	nombre_padre := name_carpetas[len(name_carpetas)-2]
	//Verifica si esa carpeta ya existe
	if ExisteCarpetaPadre(name_carpetas, int(super_bloque.S_inode_start), path) {
		consola_mkdir += "[-ERROR-] Ya existe una carpeta con el nombre: " + nueva_carpeta + "\n"
		return
	}
	name_carpetas = name_carpetas[:len(name_carpetas)-1]
	fmt.Println("Nombre de nueva carpeta: ", nueva_carpeta)
	fmt.Println("final final")
	fmt.Println(ExisteCarpetaPadre(name_carpetas, int(super_bloque.S_inode_start), path))
	if !ExisteCarpetaPadre(name_carpetas, int(super_bloque.S_inode_start), path) {
		fmt.Println(mkdir.PosCarpetaPadre(name_carpetas, int(super_bloque.S_inode_start), path))
		return
	}

	//Actualiza el inodo de la carpeta padre
	posicion_padre := mkdir.PosCarpetaPadre(name_carpetas, int(super_bloque.S_inode_start), path)
	fmt.Println("Posicion padre: ", posicion_padre)
	archivo.Seek(int64(posicion_padre), 0)
	inodo_padre := Inodo{}
	err = binary.Read(archivo, binary.LittleEndian, &inodo_padre)
	if err != nil {
		consola_mkdir += "[-ERROR-] Error al leer el Inodo\n"
		return
	}
	//nombrare en que carpeta estoy
	archivo.Seek(int64(inodo_padre.I_block[0]), 0)
	Bloque_prueba := Bloque_Carpeta{}
	binary.Read(archivo, binary.LittleEndian, &Bloque_prueba)
	fmt.Println("Nombre de la carpeta actual: ", string(Bloque_prueba.B_content[0].B_name[:]))
	var uno byte = 1
	for i := 0; i < len(inodo_padre.I_block); i++ {
		if i != len(inodo_padre.I_block)-1 {
			if inodo_padre.I_block[i] != -1 {
				bloque_carpeta := Bloque_Carpeta{}
				archivo.Seek(int64(inodo_padre.I_block[i]), 0)
				err = binary.Read(archivo, binary.LittleEndian, &bloque_carpeta)
				if err != nil {
					consola_mkdir += "[-ERROR-] Error al leer el Bloque de Carpeta\n"
					return
				}
				nombre := string(bloque_carpeta.B_content[0].B_name[:])
				nombre2 := string(bloque_carpeta.B_content[1].B_name[:])
				var hay_espacio bool = false
				for j := 2; j < 4; j++ {
					if bloque_carpeta.B_content[j].B_inodo == -1 {
						bloque_carpeta.B_content[j].B_inodo = int32(super_bloque.S_first_ino)
						copy(bloque_carpeta.B_content[j].B_name[:], nueva_carpeta)
						hay_espacio = true
						archivo.Seek(int64(inodo_padre.I_block[i]), 0)
						binary.Write(archivo, binary.LittleEndian, &bloque_carpeta)
						fmt.Println("====================================================")
						fmt.Println("ACTUALIZANDO BLOQUE DE CARPETA")
						fmt.Println("B_content[0].B_name: ", nombre)
						fmt.Println("B_content[0].B_inodo: ", bloque_carpeta.B_content[0].B_inodo)
						fmt.Println("B_content[1].B_name: ", nombre2)
						fmt.Println("B_content[1].B_inodo: ", bloque_carpeta.B_content[1].B_inodo)
						fmt.Println("B_content[2].B_name: ", string(bloque_carpeta.B_content[2].B_name[:]))
						fmt.Println("B_content[2].B_inodo: ", bloque_carpeta.B_content[2].B_inodo)
						fmt.Println("B_content[3].B_name: ", string(bloque_carpeta.B_content[3].B_name[:]))
						fmt.Println("B_content[3].B_inodo: ", bloque_carpeta.B_content[3].B_inodo)
						fmt.Println("====================================================")
						break
					}
				}

				if !hay_espacio {
					//ACTUALIZO EL NODO PADRE
					inodo_padre.I_block[i+1] = int32(super_bloque.S_first_blo)
					archivo.Seek(int64(posicion_padre), 0)
					binary.Write(archivo, binary.LittleEndian, &inodo_padre)
					//CREO EL NUEVO BLOQUE DE CARPETA
					nuevo_bloque_carpeta := Bloque_Carpeta{}
					nuevo_bloque_carpeta.B_content[0].B_inodo = int32(bloque_carpeta.B_content[0].B_inodo)
					copy(nuevo_bloque_carpeta.B_content[0].B_name[:], nombre)
					nuevo_bloque_carpeta.B_content[1].B_inodo = int32(bloque_carpeta.B_content[1].B_inodo)
					copy(nuevo_bloque_carpeta.B_content[1].B_name[:], nombre2)
					nuevo_bloque_carpeta.B_content[2].B_inodo = int32(super_bloque.S_first_ino)
					copy(nuevo_bloque_carpeta.B_content[2].B_name[:], nueva_carpeta)
					nuevo_bloque_carpeta.B_content[3].B_inodo = -1
					archivo.Seek(int64(super_bloque.S_first_blo), 0)
					binary.Write(archivo, binary.LittleEndian, &nuevo_bloque_carpeta)
					fmt.Println("====================================================")
					fmt.Println("NUEVO BLOQUE DE CARPETA PARA APUNTADORES")
					fmt.Println("B_content[0].B_inodo: ", nuevo_bloque_carpeta.B_content[0].B_inodo)
					fmt.Println("B_content[0].B_name: ", string(nuevo_bloque_carpeta.B_content[0].B_name[:]))
					fmt.Println("B_content[1].B_inodo: ", nuevo_bloque_carpeta.B_content[1].B_inodo)
					fmt.Println("B_content[1].B_name: ", string(nuevo_bloque_carpeta.B_content[1].B_name[:]))
					fmt.Println("B_content[2].B_inodo: ", nuevo_bloque_carpeta.B_content[2].B_inodo)
					fmt.Println("B_content[2].B_name: ", string(nuevo_bloque_carpeta.B_content[2].B_name[:]))
					fmt.Println("B_content[3].B_inodo: ", nuevo_bloque_carpeta.B_content[3].B_inodo)
					fmt.Println("B_content[3].B_name: ", string(nuevo_bloque_carpeta.B_content[3].B_name[:]))
					fmt.Println("====================================================")
					//ACTUALIZO EL SUPER BLOQUE
					super_bloque.S_first_blo = super_bloque.S_first_blo + int32(binary.Size(Bloque_Carpeta{}))
					super_bloque.S_blocks_count--
					archivo.Seek(int64(pos_sb), 0)
					binary.Write(archivo, binary.LittleEndian, &super_bloque)
					//ACTUALIZO EL BITMAP DE BLOQUES
					posicion_bloque := int(super_bloque.S_blocks_count) - int(super_bloque.S_free_blocks_count)
					archivo.Seek(int64((posicion_bloque+1)+int(super_bloque.S_bm_block_start)), 0)
					binary.Write(archivo, binary.LittleEndian, &uno)
					break
				}
			} else {
				bloque_carpeta := Bloque_Carpeta{}
				archivo.Seek(int64(inodo_padre.I_block[i]), 0)
				err = binary.Read(archivo, binary.LittleEndian, &bloque_carpeta)
				if err != nil {
					consola_mkdir += "[-ERROR-] Error al leer el Bloque de Carpeta\n"
					return
				}
				for j := 2; j < 4; j++ {
					if bloque_carpeta.B_content[j].B_inodo == -1 {
						bloque_carpeta.B_content[j].B_inodo = int32(super_bloque.S_first_ino)
						copy(bloque_carpeta.B_content[j].B_name[:], nueva_carpeta)
						archivo.Seek(int64(inodo_padre.I_block[i]), 0)
						binary.Write(archivo, binary.LittleEndian, &bloque_carpeta)
						break
					}
				}
			}
		}
	}

	//Crear el nuevo inodo
	nuevo_inodo := Inodo{}
	nuevo_inodo.I_uid = 1
	nuevo_inodo.I_gid = 1
	nuevo_inodo.I_size = 0
	tiempo := time.Now()
	tiempoFormateado := tiempo.Format("2006-01-02 15:04:05")
	copy(nuevo_inodo.I_atime[:], tiempoFormateado)
	copy(nuevo_inodo.I_ctime[:], tiempoFormateado)
	copy(nuevo_inodo.I_mtime[:], tiempoFormateado)
	for i := 0; i < len(nuevo_inodo.I_block); i++ {
		nuevo_inodo.I_block[i] = -1
	}
	nuevo_inodo.I_block[0] = super_bloque.S_first_blo
	nuevo_inodo.I_type = 0
	nuevo_inodo.I_perm = 664

	fmt.Println("====================================================")
	fmt.Println("NUEVO INODO")
	fmt.Println("I_uid: ", nuevo_inodo.I_uid)
	fmt.Println("I_gid: ", nuevo_inodo.I_gid)
	fmt.Println("I_size: ", nuevo_inodo.I_size)
	fmt.Println("I_atime: ", string(nuevo_inodo.I_atime[:]))
	fmt.Println("I_ctime: ", string(nuevo_inodo.I_ctime[:]))
	fmt.Println("I_mtime: ", string(nuevo_inodo.I_mtime[:]))
	fmt.Println("I_block[0]: ", nuevo_inodo.I_block[0])
	fmt.Println("I_block[1]: ", nuevo_inodo.I_block[1])
	fmt.Println("I_block[2]: ", nuevo_inodo.I_block[2])
	fmt.Println("I_block[3]: ", nuevo_inodo.I_block[3])
	fmt.Println("I_block[4]: ", nuevo_inodo.I_block[4])
	fmt.Println("I_block[5]: ", nuevo_inodo.I_block[5])
	fmt.Println("I_block[6]: ", nuevo_inodo.I_block[6])
	fmt.Println("I_block[7]: ", nuevo_inodo.I_block[7])
	fmt.Println("I_block[8]: ", nuevo_inodo.I_block[8])
	fmt.Println("I_block[9]: ", nuevo_inodo.I_block[9])
	fmt.Println("I_block[10]: ", nuevo_inodo.I_block[10])
	fmt.Println("I_block[11]: ", nuevo_inodo.I_block[11])
	fmt.Println("I_block[12]: ", nuevo_inodo.I_block[12])
	fmt.Println("I_block[13]: ", nuevo_inodo.I_block[13])
	fmt.Println("I_block[14]: ", nuevo_inodo.I_block[14])
	fmt.Println("I_block[15]: ", nuevo_inodo.I_block[15])
	fmt.Println("I_type: ", nuevo_inodo.I_type)
	fmt.Println("I_perm: ", nuevo_inodo.I_perm)

	fmt.Println("====================================================")

	archivo.Seek(int64(super_bloque.S_first_ino), 0)
	err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
	if err != nil {
		consola_mkdir += "[-ERROR-] Error al escribir el Inodo\n"
		return
	}
	pos_inodo_actual := super_bloque.S_first_ino
	//Actualizo el bitmap de inodos
	posicion_inodo := int(super_bloque.S_inodes_count) - int(super_bloque.S_free_inodes_count)
	archivo.Seek(int64((posicion_inodo+1)+int(super_bloque.S_bm_inode_start)), 0)

	err = binary.Write(archivo, binary.LittleEndian, &uno)
	if err != nil {
		consola_mkdir += "[-ERROR-] Error al escribir el Bitmap de Inodos\n"
		return
	}
	//Actualizo el super bloque
	super_bloque.S_first_ino = super_bloque.S_first_ino + int32(binary.Size(Inodo{}))
	super_bloque.S_free_inodes_count = super_bloque.S_free_inodes_count - 1
	archivo.Seek(int64(pos_sb), 0)
	err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
	if err != nil {
		consola_mkdir += "[-ERROR-] Error al escribir el SuperBloque\n"
		return
	}

	//Crear el nuevo bloque
	nuevo_bloque := Bloque_Carpeta{}
	for i := 0; i < len(nuevo_bloque.B_content); i++ {
		nuevo_bloque.B_content[i].B_inodo = -1
	}
	nuevo_bloque.B_content[0].B_inodo = pos_inodo_actual
	copy(nuevo_bloque.B_content[0].B_name[:], nueva_carpeta)
	nuevo_bloque.B_content[1].B_inodo = posicion_padre
	copy(nuevo_bloque.B_content[1].B_name[:], nombre_padre)
	nuevo_bloque.B_content[2].B_inodo = -1
	nuevo_bloque.B_content[3].B_inodo = -1

	fmt.Println("====================================================")
	fmt.Println("NUEVO BLOQUE DE CARPETA")
	fmt.Println("B_content[0].B_inodo: ", nuevo_bloque.B_content[0].B_inodo)
	fmt.Println("B_content[0].B_name: ", string(nuevo_bloque.B_content[0].B_name[:]))
	fmt.Println("B_content[1].B_inodo: ", nuevo_bloque.B_content[1].B_inodo)
	fmt.Println("B_content[1].B_name: ", string(nuevo_bloque.B_content[1].B_name[:]))
	fmt.Println("B_content[2].B_inodo: ", nuevo_bloque.B_content[2].B_inodo)
	fmt.Println("B_content[2].B_name: ", string(nuevo_bloque.B_content[2].B_name[:]))
	fmt.Println("B_content[3].B_inodo: ", nuevo_bloque.B_content[3].B_inodo)
	fmt.Println("B_content[3].B_name: ", string(nuevo_bloque.B_content[3].B_name[:]))
	fmt.Println("====================================================")

	archivo.Seek(int64(super_bloque.S_first_blo), 0)
	err = binary.Write(archivo, binary.LittleEndian, &nuevo_bloque)
	if err != nil {
		consola_mkdir += "[-ERROR-] Error al escribir el Bloque de Carpeta\n"
		return
	}

	//Actualizo el bitmap de bloques
	posicion_bloque := int(super_bloque.S_blocks_count) - int(super_bloque.S_free_blocks_count)
	archivo.Seek(int64((posicion_bloque+1)+int(super_bloque.S_bm_block_start)), 0)
	err = binary.Write(archivo, binary.LittleEndian, &uno)
	if err != nil {
		consola_mkdir += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
		return
	}
	//Actualizo el super bloque
	super_bloque.S_first_blo = super_bloque.S_first_blo + int32(binary.Size(Bloque_Carpeta{}))
	super_bloque.S_free_blocks_count = super_bloque.S_free_blocks_count - 1
	archivo.Seek(int64(pos_sb), 0)
	err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
	if err != nil {
		consola_mkdir += "[-ERROR-] Error al escribir el SuperBloque\n"
		return
	}

	consola_mkdir += "[+OK-] Carpeta creada correctamente\n"
}
func ExisteCarpetaPadre(names []string, pos int, path string) bool {
	archivo, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		consola_mkdir += "[-ERROR-] Error al abrir el archivo\n"
		return false
	}
	defer archivo.Close()

	if len(names) == 0 {
		fmt.Println(true)
		return true
	}
	inodo := Inodo{}
	archivo.Seek(int64(pos), 0)
	err = binary.Read(archivo, binary.LittleEndian, &inodo)
	if err != nil {
		consola_mkdir += "[-ERROR-] Error al leer el Inodo\n"
		return false
	}
	for i := 0; i < len(inodo.I_block); i++ {
		if inodo.I_block[i] != -1 {
			bloque := Bloque_Carpeta{}
			archivo.Seek(int64(inodo.I_block[i]), 0)
			err = binary.Read(archivo, binary.LittleEndian, &bloque)
			if err != nil {
				consola_mkdir += "[-ERROR-] Error al leer el Bloque Carpeta\n"
				return false
			}
			for j := 0; j < len(bloque.B_content); j++ {
				if bloque.B_content[j].B_inodo != -1 {
					if strings.Contains(string(bloque.B_content[j].B_name[:]), names[0]) {
						if len(names) == 1 {
							fmt.Println(true)
							return true
						} else {
							retornar := ExisteCarpetaPadre(names[1:], int(bloque.B_content[j].B_inodo), path)
							fmt.Println(retornar)
							return retornar
						}
					}
				}
			}
		}
	}
	fmt.Println(false)
	return false
}

func (mkdir *Mkdir) PosCarpetaPadre(names []string, pos int, path string) int32 {
	archivo, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		consola_mkdir += "[-ERROR-] Error al abrir el archivo\n"
		return 0
	}
	defer archivo.Close()

	if len(names) == 0 {
		return int32(pos)
	}

	//Lee el inodo
	inodo := Inodo{}
	archivo.Seek(int64(pos), 0)
	err = binary.Read(archivo, binary.LittleEndian, &inodo)
	if err != nil {
		consola_mkdir += "[-ERROR-] Error al leer el Inodo\n"
		return 0
	}
	if inodo.I_type == 1 {
		return 0
	}

	for i := 0; i < len(inodo.I_block); i++ {
		if inodo.I_block[i] != -1 {
			bloque := Bloque_Carpeta{}
			archivo.Seek(int64(inodo.I_block[i]), 0)
			err = binary.Read(archivo, binary.LittleEndian, &bloque)
			if err != nil {
				consola_mkdir += "[-ERROR-] Error al leer el Bloque Carpeta\n"
				return 0
			}
			for j := 0; j < len(bloque.B_content); j++ {
				if bloque.B_content[j].B_inodo != -1 {
					if strings.Contains(string(bloque.B_content[j].B_name[:]), names[0]) {
						if len(names) == 1 {
							return bloque.B_content[j].B_inodo
						} else {
							return mkdir.PosCarpetaPadre(names[1:], int(bloque.B_content[j].B_inodo), path)
						}
					}
				}
			}
		}
	}
	return 0
}

func (mkdir *Mkdir) CadenaVacia(cadena [16]byte) bool {

	for _, v := range cadena {
		if v != 0 {
			return false
		}
	}
	return true

}

func (mkdir *Mkdir) VerificarID() bool {
	//Verificando si existe el id
	for i := 0; i < len(ParticionesMontadasList); i++ {
		if strings.Contains(strings.ToLower(ParticionesMontadasList[i].Id), strings.ToLower(Idlogueado)) {
			return true
		}
	}
	return false
}

func (mkdir *Mkdir) RetornarStrictMontada(id string) ParticionMontada {
	for i := 0; i < len(ParticionesMontadasList); i++ {
		if strings.Contains(strings.ToLower(ParticionesMontadasList[i].Id), strings.ToLower(id)) {
			return ParticionesMontadasList[i]
		}
	}
	return ParticionMontada{}
}

func (mkdir *Mkdir) IsParticionMontadaVacia(p ParticionMontada) bool {
	return !p.Sistema_archivos && p.Id == "" && p.Letra == "" && p.Numero == 0 && p.Path == "" && p.Type == "" && p.Name == ""
}

func (mkdir *Mkdir) ObtenerParticiones(Mbr MBR) []Partition {
	var particiones []Partition
	particiones = append(particiones, Mbr.Mbr_partition_1)
	particiones = append(particiones, Mbr.Mbr_partition_2)
	particiones = append(particiones, Mbr.Mbr_partition_3)
	particiones = append(particiones, Mbr.Mbr_partition_4)
	return particiones
}

func (mkdir *Mkdir) ListadoEBR(Extendida Partition, path string) []EBR {
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
		if !mkdir.CadenaVacia(ebr.Part_name) {
			ebrs = append(ebrs, ebr)
		} else if mkdir.CadenaVacia(ebr.Part_name) && ebr.Part_size != 0 {
			ebrs = append(ebrs, ebr)
		} else {
			break
		}
		temp = ebr.Part_next
	}
	return ebrs
}

func RetornarConsolamkdir() string {
	return consola_mkdir
}
