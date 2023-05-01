package analizador

import (
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Mkuser struct {
	User string
	Pwd  string
	Grp  string
}

var consola_mkuser string

func (mkuser *Mkuser) VerificarParams(parametros map[string]string) {
	consola_mkuser = ""
	//Verificando parametros obligatorios
	if mkuser.User == "" {
		consola_mkuser += "[-ERROR-] Falta el parametro user\n"
		return
	}
	if mkuser.Pwd == "" {
		consola_mkuser += "[-ERROR-] Falta el parametro pwd\n"
		return
	}

	if mkuser.User[0] == '"' && mkuser.User[len(mkuser.User)-1] == '"' {
		mkuser.User = mkuser.User[1 : len(mkuser.User)-1]
	}

	if mkuser.Pwd[0] == '"' && mkuser.Pwd[len(mkuser.Pwd)-1] == '"' {
		mkuser.Pwd = mkuser.Pwd[1 : len(mkuser.Pwd)-1]
	}

	if mkuser.Grp != "" {
		if mkuser.Grp[0] == '"' && mkuser.Grp[len(mkuser.Grp)-1] == '"' {
			mkuser.Grp = mkuser.Grp[1 : len(mkuser.Grp)-1]
		}
	} else {
		consola_mkuser += "[-ERROR-] Falta el parametro grp\n"
		return
	}

	if len(mkuser.User) > 10 {
		consola_mkuser += "[-ERROR-] El nombre de usuario no puede ser mayor a 10 caracteres\n"
		return
	}

	if len(mkuser.Pwd) > 10 {
		consola_mkuser += "[-ERROR-] La contraseña no puede ser mayor a 10 caracteres\n"
		return
	}

	if len(mkuser.Grp) > 10 {
		consola_mkuser += "[-ERROR-] El nombre del grupo no puede ser mayor a 10 caracteres\n"
		return
	}

	mkuser.CrearUsuario()

}

func (mkuser *Mkuser) CrearUsuario() {
	montada := mkuser.RetornarStrictMontada(Idlogueado)
	if mkuser.IsParticionMontadaVacia(montada) {
		consola_mkuser += "[-ERROR-] La partición con id: " + Idlogueado + " no está montada\n"
		return
	}

	if !montada.Sistema_archivos {
		consola_mkuser += "[-ERROR-] La partición con id: " + Idlogueado + " no tiene un sistema de archivos\n"
	}

	if montada.User != "root" {
		consola_mkuser += "[-ERROR-] No se tienen los permisos suficientes para crear un grupo\n"
		return
	}

	//Abrir el archivo binario
	archivo, err := os.OpenFile(montada.Path, os.O_RDWR, 0666)
	if err != nil {
		consola_mkuser += "[-ERROR-] Error al abrir el archivo\n"
		return
	}
	defer archivo.Close()

	//Leer el MBR
	mbr := MBR{}
	archivo.Seek(int64(0), 0)
	err = binary.Read(archivo, binary.LittleEndian, &mbr)
	if err != nil {
		consola_mkuser += "[-ERROR-] Error al leer el MBR\n"
		return
	}
	fmt.Println("MBR DESDE REP")
	fmt.Println(mbr)
	particiones := mkuser.ObtenerParticiones(mbr)
	var ebrs []EBR
	logica_existe := false
	var particion_logica EBR
	for i := 0; i < len(particiones); i++ {
		if strings.ToLower(string(particiones[i].Part_type[0])) == "e" {
			ebrs = mkuser.ListadoEBR(particiones[i], montada.Path)
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
		consola_mkuser += "[-ERROR-] Error al leer el SuperBloque\n"
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
		consola_mkuser += "[-ERROR-] Error al leer el Inodo del archivo users.txt\n"
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
			consola_mkuser += "[-ERROR-] Error al leer el Bloque de datos del archivo users.txt\n"
			return
		}
		usuariostxt += string(bloque.B_content[:])
	}

	usuariostxt = strings.Replace(usuariostxt, "\u0000", "", -1)
	fmt.Println("=====================")
	fmt.Println(usuariostxt)

	usuarios_grupos := strings.Split(usuariostxt, "\n")
	//Verifica si ya existe el usuario
	for i := 0; i < len(usuarios_grupos); i++ {
		datos := strings.Split(usuarios_grupos[i], ",")
		fmt.Println(datos)
		if len(datos) == 5 {
			if strings.Contains(datos[1], "U") {
				if !strings.Contains(datos[0], "0") {
					if strings.Contains(datos[2], mkuser.User) {
						consola_mkuser += "[-ERROR-] El usuario que quieres crear ya existe\n"
						return
					}
				}
			}
		}
	}

	error_grp := true
	//Verifica si ya existe el grupo
	for i := 0; i < len(usuarios_grupos); i++ {
		datos := strings.Split(usuarios_grupos[i], ",")
		fmt.Println(datos)
		if len(datos) == 3 {
			if strings.Contains(datos[1], "G") {
				if !strings.Contains(datos[0], "0") {
					if strings.Contains(datos[2], mkuser.Grp) {
						error_grp = false
					}
				}
			}
		}
	}

	if error_grp {
		consola_mkuser += "[-ERROR-] El grupo que quieres asignar no existe\n"
		return
	}

	fmt.Println(usuariostxt)
	//Obtengo el ultimo numero de grupo creado
	ultimo_usuario := 0
	for i := 0; i < len(usuarios_grupos); i++ {
		datos := strings.Split(usuarios_grupos[i], ",")
		if len(datos) == 5 {
			if strings.Contains(datos[1], "U") {
				ultimo_usuario++
			}
		}
	}
	nuevo_usuario := strconv.Itoa(ultimo_usuario + 1)
	//Creo el nuevo grupo
	nueva_linea := nuevo_usuario + ",U," + mkuser.Grp + "," + mkuser.User + "," + mkuser.Pwd + "\n"

	//ESCRIBIR EL CONTENIDO DEL ARCHIVO
	for i := 0; i < 16; i++ {
		if i != 15 {
			no_bloque := int(inodo_users.I_block[i])
			if no_bloque != -1 {
				if inodo_users.I_block[i+1] == -1 {
					bloque := Bloque_Archivo{}
					archivo.Seek(int64(no_bloque), 0)
					err = binary.Read(archivo, binary.LittleEndian, &bloque)
					if err != nil {
						consola_mkuser += "[-ERROR-] Error al leer el Bloque de datos del archivo users.txt\n"
						return
					}
					contenido := string(bloque.B_content[:])
					contenido = strings.Replace(contenido, "\u0000", "", -1)
					espacio := (64 - len(contenido)) - len(nueva_linea)
					if espacio >= 0 {
						if 64-len(contenido) == 0 {
							fmt.Println("PARTE 1")
							//Crear nuevo bloque Archivo
							nuevo := Bloque_Archivo{}
							copy(nuevo.B_content[:], nueva_linea)
							archivo.Seek(int64(super_bloque.S_first_blo), 0)
							err = binary.Write(archivo, binary.LittleEndian, &nuevo)
							if err != nil {
								consola_mkuser += "[-ERROR-] Error al escribir el Bloque de datos del archivo users.txt\n"
								return
							}
							//Actualizar el inodo
							inodo_users.I_size += int32(len(nueva_linea))
							tiempo := time.Now()
							tiempoFormateado := tiempo.Format("2006-01-02 15:04:05")
							copy(inodo_users.I_mtime[:], tiempoFormateado)
							inodo_users.I_block[i+1] = int32(super_bloque.S_first_blo)
							archivo.Seek(int64(pos_inodo), 0)
							err = binary.Write(archivo, binary.LittleEndian, &inodo_users)
							if err != nil {
								consola_mkuser += "[-ERROR-] Error al escribir el Inodo del archivo users.txt\n"
								return
							}
							//Actualizar el SuperBloque
							super_bloque.S_first_blo += int32(binary.Size(Bloque_Archivo{}))
							super_bloque.S_free_blocks_count--
							archivo.Seek(int64(pos_inicio), 0)
							err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
							if err != nil {
								consola_mkuser += "[-ERROR-] Error al escribir el SuperBloque\n"
								return
							}
							break
						} else {
							fmt.Println("PARTE 2")
							contenido += nueva_linea
							copy(bloque.B_content[:], contenido)
							archivo.Seek(int64(no_bloque), 0)
							err = binary.Write(archivo, binary.LittleEndian, &bloque)
							if err != nil {
								consola_mkuser += "[-ERROR-] Error al escribir el Bloque de datos del archivo users.txt\n"
								return
							}
							//Actualizar el inodo
							inodo_users.I_size += int32(len(nueva_linea))
							tiempo := time.Now()
							tiempoFormateado := tiempo.Format("2006-01-02 15:04:05")
							copy(inodo_users.I_mtime[:], tiempoFormateado)
							archivo.Seek(int64(pos_inodo), 0)
							err = binary.Write(archivo, binary.LittleEndian, &inodo_users)
							if err != nil {
								consola_mkuser += "[-ERROR-] Error al escribir el Inodo del archivo users.txt\n"
								return
							}
							break
						}
					} else if espacio < 0 {
						if 64-len(contenido) != 0 {
							fmt.Println("PARTE 3")
							espacio_disponible := 64 - len(contenido)
							parte1 := nueva_linea[:espacio_disponible]
							parte2 := nueva_linea[espacio_disponible:]
							contenido += parte1
							//SETEA EL BLOQUE QUE YA ESTABA LLENO CON LA PRIMERA PARTE
							copy(bloque.B_content[:], contenido)
							archivo.Seek(int64(no_bloque), 0)
							err = binary.Write(archivo, binary.LittleEndian, &bloque)
							if err != nil {
								consola_mkuser += "[-ERROR-] Error al escribir el Bloque de datos del archivo users.txt\n"
								return
							}
							//CREA EL NUEVO BLOQUE CON LA SEGUNDA PARTE
							nuevo := Bloque_Archivo{}
							copy(nuevo.B_content[:], parte2)
							archivo.Seek(int64(super_bloque.S_first_blo), 0)
							err = binary.Write(archivo, binary.LittleEndian, &nuevo)
							if err != nil {
								consola_mkuser += "[-ERROR-] Error al escribir el Bloque de datos del archivo users.txt\n"
								return
							}
							//Actualizar el inodo
							inodo_users.I_size += int32(len(nueva_linea))
							tiempo := time.Now()
							tiempoFormateado := tiempo.Format("2006-01-02 15:04:05")
							copy(inodo_users.I_mtime[:], tiempoFormateado)
							inodo_users.I_block[i+1] = int32(super_bloque.S_first_blo)
							archivo.Seek(int64(pos_inodo), 0)
							err = binary.Write(archivo, binary.LittleEndian, &inodo_users)
							if err != nil {
								consola_mkuser += "[-ERROR-] Error al escribir el Inodo del archivo users.txt\n"
								return
							}
							//Actualizar el SuperBloque
							super_bloque.S_first_blo += int32(binary.Size(Bloque_Archivo{}))
							super_bloque.S_free_blocks_count--
							archivo.Seek(int64(pos_inicio), 0)
							err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
							if err != nil {
								consola_mkuser += "[-ERROR-] Error al escribir el SuperBloque\n"
								return
							}
							break
						}
					}
				}
			}
		} else {
			no_bloque := inodo_users.I_block[i]
			bloque := Bloque_Archivo{}
			archivo.Seek(int64(no_bloque), 0)
			err = binary.Read(archivo, binary.LittleEndian, &bloque)
			if err != nil {
				consola_mkuser += "[-ERROR-] Error al leer el Bloque de datos del archivo users.txt\n"
				return
			}
			contenido := string(bloque.B_content[:])
			contenido = strings.Replace(contenido, "\u0000", "", -1)
			espacio := 64 - len(contenido) - len(nueva_linea)
			if espacio >= 0 {
				contenido += nueva_linea
				copy(bloque.B_content[:], contenido)
				archivo.Seek(int64(no_bloque), 0)
				err = binary.Write(archivo, binary.LittleEndian, &bloque)
				if err != nil {
					consola_mkuser += "[-ERROR-] Error al escribir el Bloque de datos del archivo users.txt\n"
					return
				}
				//Actualizar el inodo
				inodo_users.I_size += int32(len(nueva_linea))
				tiempo := time.Now()
				tiempoFormateado := tiempo.Format("2006-01-02 15:04:05")
				copy(inodo_users.I_mtime[:], tiempoFormateado)
				archivo.Seek(int64(pos_inodo), 0)
				err = binary.Write(archivo, binary.LittleEndian, &inodo_users)
				if err != nil {
					consola_mkuser += "[-ERROR-] Error al escribir el Inodo del archivo users.txt\n"
					return
				}
				break
			}
		}
	}

	consola_mkuser += "[*SUCCESS*] Se ha creado el usuario: " + mkuser.User + " exitosamente\n"
}

func (mkuser *Mkuser) CadenaVacia(cadena [16]byte) bool {

	for _, v := range cadena {
		if v != 0 {
			return false
		}
	}
	return true

}

func (mkuser *Mkuser) VerificarID() bool {
	//Verificando si existe el id
	for i := 0; i < len(ParticionesMontadasList); i++ {
		if strings.Contains(strings.ToLower(ParticionesMontadasList[i].Id), strings.ToLower(Idlogueado)) {
			return true
		}
	}
	return false
}

func (mkuser *Mkuser) RetornarStrictMontada(id string) ParticionMontada {
	for i := 0; i < len(ParticionesMontadasList); i++ {
		if strings.Contains(strings.ToLower(ParticionesMontadasList[i].Id), strings.ToLower(id)) {
			return ParticionesMontadasList[i]
		}
	}
	return ParticionMontada{}
}

func (mkuser *Mkuser) IsParticionMontadaVacia(p ParticionMontada) bool {
	return !p.Sistema_archivos && p.Id == "" && p.Letra == "" && p.Numero == 0 && p.Path == "" && p.Type == "" && p.Name == ""
}

func (mkuser *Mkuser) ObtenerParticiones(Mbr MBR) []Partition {
	var particiones []Partition
	particiones = append(particiones, Mbr.Mbr_partition_1)
	particiones = append(particiones, Mbr.Mbr_partition_2)
	particiones = append(particiones, Mbr.Mbr_partition_3)
	particiones = append(particiones, Mbr.Mbr_partition_4)
	return particiones
}

func (mkuser *Mkuser) ListadoEBR(Extendida Partition, path string) []EBR {
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
		if !mkuser.CadenaVacia(ebr.Part_name) {
			ebrs = append(ebrs, ebr)
		} else if mkuser.CadenaVacia(ebr.Part_name) && ebr.Part_size != 0 {
			ebrs = append(ebrs, ebr)
		} else {
			break
		}
		temp = ebr.Part_next
	}
	return ebrs
}

func RetornarConsolamkuser() string {
	return consola_mkuser
}
