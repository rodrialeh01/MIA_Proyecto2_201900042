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
	name_carpetas := strings.Split(mkdir.Path, "/")
	if name_carpetas[0] == "" {
		name_carpetas[0] = "/"
	} else {
		consola_mkdir += "[-ERROR-] La ruta no es absoluta\n"
	}
	mkdir.CreacionRecursiva(name_carpetas, pos_sb, path)
	consola_mkdir += "[*SUCCESS*] Se han creado las carpetas correctamente\n"
}
func (mkdir *Mkdir) CreacionRecursiva(nombres_carpetas []string, pos_sb int, path string) {
	fmt.Println("CARPETAS")
	fmt.Println(nombres_carpetas)
	archivo, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		consola_mkdir += "[-ERROR-] Error al abrir el archivo\n"
		return
	}
	defer archivo.Close()

	//LEE EL SUPERBLOQUE

	archivo.Seek(int64(pos_sb), 0)
	super_bloque := SuperBloque{}
	fmt.Println("POS SB: ", pos_sb)
	err = binary.Read(archivo, binary.LittleEndian, &super_bloque)
	if err != nil {
		consola_mkdir += "[-ERROR-] Error al leer el SuperBloque\n"
		return
	}
	fmt.Println("SUPERBLOQUE")
	fmt.Println(super_bloque)

	for i := 0; i < len(nombres_carpetas)-1; i++ {
		fmt.Println(i)
		if mkdir.ExisteCarpetaPadre(nombres_carpetas[:i+1], int(super_bloque.S_inode_start), path) && !mkdir.ExisteCarpetaPadre(nombres_carpetas[:i+2], int(super_bloque.S_inode_start), path) {
			posicion_padre := mkdir.PosCarpetaPadre(nombres_carpetas[:i+1], int(super_bloque.S_inode_start), path)
			fmt.Println("POSICION PADRE: ", posicion_padre)
			nueva_carpeta := nombres_carpetas[i+1]
			if len(nueva_carpeta) > 12 {
				consola_mkdir += "[-ERROR-] El nombre de la carpeta no puede ser mayor a 12 caracteres\n"
				return
			}
			nombre_padre := nombres_carpetas[i]
			archivo.Seek(int64(posicion_padre), 0)
			//LEE EL INODO PADRE
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

			tiempo := time.Now()
			tiempoFormateado := tiempo.Format("2006-01-02 15:04:05")

			for i := 0; i < len(inodo_padre.I_block); i++ {
				if i != 15 {
					no_bloque := int(inodo_padre.I_block[i])
					if no_bloque != -1 {
						if inodo_padre.I_block[i+1] == -1 {
							bloque := Bloque_Carpeta{}
							archivo.Seek(int64(no_bloque), 0)
							err = binary.Read(archivo, binary.LittleEndian, &bloque)
							if err != nil {
								consola_mkdir += "[-ERROR-] Error al leer el Bloque\n"
								return
							}
							var hay_espacio bool = false
							var pos_b int = 0
							for j := 2; j < 4; j++ {
								name_comp := string(bloque.B_content[j].B_name[:])
								name_comp = strings.Replace(name_comp, "\u0000", "", -1)
								if name_comp == nueva_carpeta {
									consola_mkdir += "[-ERROR-] Ya existe una carpeta con ese nombre\n"
									return
								}
								if bloque.B_content[j].B_inodo == -1 {
									hay_espacio = true
									pos_b = j
									break
								}
							}
							if hay_espacio {
								//Crear el nuevo inodo
								nuevo_inodo := Inodo{}
								nuevo_inodo.I_uid = int32(Id_UserLogueado)
								nuevo_inodo.I_gid = int32(Id_GroupLogueado)
								nuevo_inodo.I_size = 0
								copy(nuevo_inodo.I_atime[:], tiempoFormateado)
								copy(nuevo_inodo.I_ctime[:], tiempoFormateado)
								copy(nuevo_inodo.I_mtime[:], tiempoFormateado)
								nuevo_inodo.I_type = 0
								nuevo_inodo.I_perm = 664
								for j := 0; j < 16; j++ {
									nuevo_inodo.I_block[j] = -1
								}
								nuevo_inodo.I_block[0] = super_bloque.S_first_blo
								//Escribir el nuevo inodo
								posicion_nuevo_inodo := super_bloque.S_first_ino
								archivo.Seek(int64(super_bloque.S_first_ino), 0)
								err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
								if err != nil {
									consola_mkdir += "[-ERROR-] Error al escribir el Inodo\n"
									return
								}
								//Actualizar el super bloque
								super_bloque.S_first_ino += int32(binary.Size(nuevo_inodo))
								super_bloque.S_free_inodes_count -= 1
								archivo.Seek(int64(pos_sb), 0)
								err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
								if err != nil {
									consola_mkdir += "[-ERROR-] Error al escribir el Super Bloque\n"
									return
								}
								//Actualizar el bitmap de inodos
								pos_bitmap := super_bloque.S_inodes_count - super_bloque.S_free_inodes_count
								archivo.Seek(int64(super_bloque.S_bm_inode_start+pos_bitmap), 0)
								var uno byte = 1
								err = binary.Write(archivo, binary.LittleEndian, &uno)
								if err != nil {
									consola_mkdir += "[-ERROR-] Error al escribir el Bitmap de Inodos\n"
									return
								}
								//Actualizar el bloque de carpetas
								bloque.B_content[pos_b].B_inodo = int32(posicion_nuevo_inodo)
								copy(bloque.B_content[pos_b].B_name[:], nueva_carpeta)
								archivo.Seek(int64(no_bloque), 0)
								err = binary.Write(archivo, binary.LittleEndian, &bloque)
								if err != nil {
									consola_mkdir += "[-ERROR-] Error al escribir el Bloque\n"
									return
								}
								//Crear el nuevo bloque de carpetas
								nuevo_bloque := Bloque_Carpeta{}
								nuevo_bloque.B_content[0].B_inodo = int32(posicion_nuevo_inodo)
								copy(nuevo_bloque.B_content[0].B_name[:], nueva_carpeta)
								nuevo_bloque.B_content[1].B_inodo = int32(posicion_padre)
								copy(nuevo_bloque.B_content[1].B_name[:], nombre_padre)
								for j := 2; j < 4; j++ {
									nuevo_bloque.B_content[j].B_inodo = -1
								}
								//Escribir el nuevo bloque de carpetas
								archivo.Seek(int64(super_bloque.S_first_blo), 0)
								err = binary.Write(archivo, binary.LittleEndian, &nuevo_bloque)
								if err != nil {
									consola_mkdir += "[-ERROR-] Error al escribir el Bloque\n"
									return
								}
								//Actualizar el super bloque
								super_bloque.S_first_blo += int32(binary.Size(nuevo_bloque))
								super_bloque.S_free_blocks_count -= 1
								archivo.Seek(int64(pos_sb), 0)
								err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
								if err != nil {
									consola_mkdir += "[-ERROR-] Error al escribir el Super Bloque\n"
									return
								}
								//Actualizar el bitmap de bloques
								pos_bitmap = super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
								archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
								err = binary.Write(archivo, binary.LittleEndian, &uno)
								if err != nil {
									consola_mkdir += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
									return
								}

								mkdir.CreacionRecursiva(nombres_carpetas, pos_sb, path)
							} else {
								siguiente_bloque := Bloque_Carpeta{}
								siguiente_bloque.B_content[0].B_inodo = int32(Bloque_prueba.B_content[0].B_inodo)
								copy(siguiente_bloque.B_content[0].B_name[:], Bloque_prueba.B_content[0].B_name[:])
								siguiente_bloque.B_content[1].B_inodo = int32(Bloque_prueba.B_content[1].B_inodo)
								copy(siguiente_bloque.B_content[1].B_name[:], Bloque_prueba.B_content[1].B_name[:])
								for j := 2; j < 4; j++ {
									siguiente_bloque.B_content[j].B_inodo = -1
								}
								//Escribir el nuevo bloque de carpetas
								pos_sig_bloque := super_bloque.S_first_blo
								archivo.Seek(int64(pos_sig_bloque), 0)
								err = binary.Write(archivo, binary.LittleEndian, &siguiente_bloque)
								if err != nil {
									consola_mkdir += "[-ERROR-] Error al escribir el Bloque\n"
									return
								}
								//Actualizar el super bloque
								super_bloque.S_first_blo += int32(binary.Size(siguiente_bloque))
								super_bloque.S_free_blocks_count -= 1
								archivo.Seek(int64(pos_sb), 0)
								err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
								if err != nil {
									consola_mkdir += "[-ERROR-] Error al escribir el Super Bloque\n"
									return
								}
								var uno byte = 1
								//Actualizar el bitmap de bloques
								pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
								archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
								err = binary.Write(archivo, binary.LittleEndian, &uno)
								if err != nil {
									consola_mkdir += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
									return
								}
								//Actualizar el inodo padre
								inodo_padre.I_block[i+1] = int32(pos_sig_bloque)
								archivo.Seek(int64(Bloque_prueba.B_content[0].B_inodo), 0)
								err = binary.Write(archivo, binary.LittleEndian, &inodo_padre)
								if err != nil {
									consola_mkdir += "[-ERROR-] Error al escribir el Inodo\n"
									return
								}

								//CREAR LA NUEVA CARPETA
								nuevo_inodo := Inodo{}
								nuevo_inodo.I_uid = int32(Id_UserLogueado)
								nuevo_inodo.I_gid = int32(Id_GroupLogueado)
								nuevo_inodo.I_size = 0
								copy(nuevo_inodo.I_atime[:], []byte(tiempoFormateado))
								copy(nuevo_inodo.I_ctime[:], []byte(tiempoFormateado))
								copy(nuevo_inodo.I_mtime[:], []byte(tiempoFormateado))
								for j := 0; j < 16; j++ {
									nuevo_inodo.I_block[j] = -1
								}
								nuevo_inodo.I_block[0] = int32(super_bloque.S_first_blo)
								nuevo_inodo.I_type = 0
								nuevo_inodo.I_perm = 664
								//Escribir el nuevo inodo
								pos_nuevo_inodo := super_bloque.S_first_ino
								archivo.Seek(int64(pos_nuevo_inodo), 0)
								err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
								if err != nil {
									consola_mkdir += "[-ERROR-] Error al escribir el Inodo\n"
									return
								}
								//Actualizar el super bloque
								super_bloque.S_first_ino += int32(binary.Size(nuevo_inodo))
								super_bloque.S_free_inodes_count -= 1
								archivo.Seek(int64(pos_sb), 0)
								err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
								if err != nil {
									consola_mkdir += "[-ERROR-] Error al escribir el Super Bloque\n"
									return
								}
								//Actualizar el bitmap de inodos
								pos_bitmap = super_bloque.S_inodes_count - super_bloque.S_free_inodes_count
								archivo.Seek(int64(super_bloque.S_bm_inode_start+pos_bitmap), 0)
								err = binary.Write(archivo, binary.LittleEndian, &uno)
								if err != nil {
									consola_mkdir += "[-ERROR-] Error al escribir el Bitmap de Inodos\n"
									return
								}
								//Actualizar el bloque de carpeta
								siguiente_bloque.B_content[2].B_inodo = int32(pos_nuevo_inodo)
								copy(siguiente_bloque.B_content[2].B_name[:], nueva_carpeta)
								archivo.Seek(int64(pos_sig_bloque), 0)
								err = binary.Write(archivo, binary.LittleEndian, &siguiente_bloque)
								if err != nil {
									consola_mkdir += "[-ERROR-] Error al escribir el Bloque\n"
									return
								}
								//Crea el bloque carpeta
								nuevo_bloque := Bloque_Carpeta{}
								nuevo_bloque.B_content[0].B_inodo = int32(pos_nuevo_inodo)
								copy(nuevo_bloque.B_content[0].B_name[:], nueva_carpeta)
								nuevo_bloque.B_content[1].B_inodo = int32(Bloque_prueba.B_content[0].B_inodo)
								copy(nuevo_bloque.B_content[1].B_name[:], Bloque_prueba.B_content[0].B_name[:])
								for j := 2; j < 4; j++ {
									nuevo_bloque.B_content[j].B_inodo = -1
								}
								//Escribir el nuevo bloque
								pos_nuevo_bloque := super_bloque.S_first_blo
								archivo.Seek(int64(pos_nuevo_bloque), 0)
								err = binary.Write(archivo, binary.LittleEndian, &nuevo_bloque)
								if err != nil {
									consola_mkdir += "[-ERROR-] Error al escribir el Bloque\n"
									return
								}
								//Actualizar el super bloque
								super_bloque.S_first_blo += int32(binary.Size(nuevo_bloque))
								super_bloque.S_free_blocks_count -= 1
								archivo.Seek(int64(pos_sb), 0)
								err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
								if err != nil {
									consola_mkdir += "[-ERROR-] Error al escribir el Super Bloque\n"
									return
								}
								//Actualizar el bitmap de bloques
								pos_bitmap = super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
								archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
								err = binary.Write(archivo, binary.LittleEndian, &uno)
								if err != nil {
									consola_mkdir += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
									return
								}
								mkdir.CreacionRecursiva(nombres_carpetas, pos_sb, path)
							}
						}
					}
				}
			}
		}
	}

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
	if name_carpetas[0] == "" {
		name_carpetas[0] = "/"
	} else {
		consola_mkdir += "[-ERROR-] La ruta no es absoluta\n"
	}
	nueva_carpeta := name_carpetas[len(name_carpetas)-1]
	if len(nueva_carpeta) > 12 {
		consola_mkdir += "[-ERROR-] El nombre de la carpeta no puede ser mayor a 12 caracteres\n"
		return
	}
	nombre_padre := name_carpetas[len(name_carpetas)-2]
	//Verifica si esa carpeta ya existe
	if mkdir.ExisteCarpetaPadre(name_carpetas, int(super_bloque.S_inode_start), path) {
		consola_mkdir += "[-ERROR-] Ya existe una carpeta con el nombre: " + nueva_carpeta + "\n"
		return
	}
	name_carpetas = name_carpetas[:len(name_carpetas)-1]
	fmt.Println("Nombre de nueva carpeta: ", nueva_carpeta)
	fmt.Println("final final")
	fmt.Println(mkdir.ExisteCarpetaPadre(name_carpetas, int(super_bloque.S_inode_start), path))
	if !mkdir.ExisteCarpetaPadre(name_carpetas, int(super_bloque.S_inode_start), path) {
		consola_mkdir += "[-ERROR-] No existe la carpeta donde quieres agregar una nueva carpeta\n"
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
	tiempo := time.Now()
	tiempoFormateado := tiempo.Format("2006-01-02 15:04:05")
	for i := 0; i < len(inodo_padre.I_block); i++ {
		if i != 15 {
			no_bloque := int(inodo_padre.I_block[i])
			if no_bloque != -1 {
				if inodo_padre.I_block[i+1] == -1 {
					bloque := Bloque_Carpeta{}
					archivo.Seek(int64(no_bloque), 0)
					err = binary.Read(archivo, binary.LittleEndian, &bloque)
					if err != nil {
						consola_mkdir += "[-ERROR-] Error al leer el Bloque\n"
						return
					}
					var hay_espacio bool = false
					var pos_b int = 0
					for j := 2; j < 4; j++ {
						if bloque.B_content[j].B_inodo == -1 {
							hay_espacio = true
							pos_b = j
							break
						}
					}
					if hay_espacio {
						//Crear el nuevo inodo
						nuevo_inodo := Inodo{}
						nuevo_inodo.I_uid = int32(Id_UserLogueado)
						nuevo_inodo.I_gid = int32(Id_GroupLogueado)
						nuevo_inodo.I_size = 0
						copy(nuevo_inodo.I_atime[:], tiempoFormateado)
						copy(nuevo_inodo.I_ctime[:], tiempoFormateado)
						copy(nuevo_inodo.I_mtime[:], tiempoFormateado)
						nuevo_inodo.I_type = 0
						nuevo_inodo.I_perm = 664
						for j := 0; j < 16; j++ {
							nuevo_inodo.I_block[j] = -1
						}
						nuevo_inodo.I_block[0] = super_bloque.S_first_blo
						//Escribir el nuevo inodo
						posicion_nuevo_inodo := super_bloque.S_first_ino
						archivo.Seek(int64(super_bloque.S_first_ino), 0)
						err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
						if err != nil {
							consola_mkdir += "[-ERROR-] Error al escribir el Inodo\n"
							return
						}
						//Actualizar el super bloque
						super_bloque.S_first_ino += int32(binary.Size(nuevo_inodo))
						super_bloque.S_free_inodes_count -= 1
						archivo.Seek(int64(pos_sb), 0)
						err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
						if err != nil {
							consola_mkdir += "[-ERROR-] Error al escribir el Super Bloque\n"
							return
						}
						//Actualizar el bitmap de inodos
						pos_bitmap := super_bloque.S_inodes_count - super_bloque.S_free_inodes_count
						archivo.Seek(int64(super_bloque.S_bm_inode_start+pos_bitmap), 0)
						var uno byte = 1
						err = binary.Write(archivo, binary.LittleEndian, &uno)
						if err != nil {
							consola_mkdir += "[-ERROR-] Error al escribir el Bitmap de Inodos\n"
							return
						}
						//Actualizar el bloque de carpetas
						bloque.B_content[pos_b].B_inodo = int32(posicion_nuevo_inodo)
						copy(bloque.B_content[pos_b].B_name[:], nueva_carpeta)
						archivo.Seek(int64(no_bloque), 0)
						err = binary.Write(archivo, binary.LittleEndian, &bloque)
						if err != nil {
							consola_mkdir += "[-ERROR-] Error al escribir el Bloque\n"
							return
						}
						//Crear el nuevo bloque de carpetas
						nuevo_bloque := Bloque_Carpeta{}
						nuevo_bloque.B_content[0].B_inodo = int32(posicion_nuevo_inodo)
						copy(nuevo_bloque.B_content[0].B_name[:], []byte(name_carpetas[len(name_carpetas)-1]))
						nuevo_bloque.B_content[1].B_inodo = int32(posicion_padre)
						copy(nuevo_bloque.B_content[1].B_name[:], nombre_padre)
						for j := 2; j < 4; j++ {
							nuevo_bloque.B_content[j].B_inodo = -1
						}
						//Escribir el nuevo bloque de carpetas
						archivo.Seek(int64(super_bloque.S_first_blo), 0)
						err = binary.Write(archivo, binary.LittleEndian, &nuevo_bloque)
						if err != nil {
							consola_mkdir += "[-ERROR-] Error al escribir el Bloque\n"
							return
						}
						//Actualizar el super bloque
						super_bloque.S_first_blo += int32(binary.Size(nuevo_bloque))
						super_bloque.S_free_blocks_count -= 1
						archivo.Seek(int64(pos_sb), 0)
						err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
						if err != nil {
							consola_mkdir += "[-ERROR-] Error al escribir el Super Bloque\n"
							return
						}
						//Actualizar el bitmap de bloques
						pos_bitmap = super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
						archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
						err = binary.Write(archivo, binary.LittleEndian, &uno)
						if err != nil {
							consola_mkdir += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
							return
						}

						consola_mkdir += "[*SUCCESS*] Se ha creado la carpeta " + nueva_carpeta + " correctamente\n"
						return
					} else {
						siguiente_bloque := Bloque_Carpeta{}
						siguiente_bloque.B_content[0].B_inodo = int32(Bloque_prueba.B_content[0].B_inodo)
						copy(siguiente_bloque.B_content[0].B_name[:], Bloque_prueba.B_content[0].B_name[:])
						siguiente_bloque.B_content[1].B_inodo = int32(Bloque_prueba.B_content[1].B_inodo)
						copy(siguiente_bloque.B_content[1].B_name[:], Bloque_prueba.B_content[1].B_name[:])
						for j := 2; j < 4; j++ {
							siguiente_bloque.B_content[j].B_inodo = -1
						}
						//Escribir el nuevo bloque de carpetas
						pos_sig_bloque := super_bloque.S_first_blo
						archivo.Seek(int64(pos_sig_bloque), 0)
						err = binary.Write(archivo, binary.LittleEndian, &siguiente_bloque)
						if err != nil {
							consola_mkdir += "[-ERROR-] Error al escribir el Bloque\n"
							return
						}
						//Actualizar el super bloque
						super_bloque.S_first_blo += int32(binary.Size(siguiente_bloque))
						super_bloque.S_free_blocks_count -= 1
						archivo.Seek(int64(pos_sb), 0)
						err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
						if err != nil {
							consola_mkdir += "[-ERROR-] Error al escribir el Super Bloque\n"
							return
						}
						var uno byte = 1
						//Actualizar el bitmap de bloques
						pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
						archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
						err = binary.Write(archivo, binary.LittleEndian, &uno)
						if err != nil {
							consola_mkdir += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
							return
						}
						//Actualizar el inodo padre
						inodo_padre.I_block[i+1] = int32(pos_sig_bloque)
						archivo.Seek(int64(Bloque_prueba.B_content[0].B_inodo), 0)
						err = binary.Write(archivo, binary.LittleEndian, &inodo_padre)
						if err != nil {
							consola_mkdir += "[-ERROR-] Error al escribir el Inodo\n"
							return
						}

						//CREAR LA NUEVA CARPETA
						nuevo_inodo := Inodo{}
						nuevo_inodo.I_uid = int32(Id_UserLogueado)
						nuevo_inodo.I_gid = int32(Id_GroupLogueado)
						nuevo_inodo.I_size = 0
						copy(nuevo_inodo.I_atime[:], []byte(tiempoFormateado))
						copy(nuevo_inodo.I_ctime[:], []byte(tiempoFormateado))
						copy(nuevo_inodo.I_mtime[:], []byte(tiempoFormateado))
						for j := 0; j < 16; j++ {
							nuevo_inodo.I_block[j] = -1
						}
						nuevo_inodo.I_block[0] = int32(super_bloque.S_first_blo)
						nuevo_inodo.I_type = 0
						nuevo_inodo.I_perm = 664
						//Escribir el nuevo inodo
						pos_nuevo_inodo := super_bloque.S_first_ino
						archivo.Seek(int64(pos_nuevo_inodo), 0)
						err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
						if err != nil {
							consola_mkdir += "[-ERROR-] Error al escribir el Inodo\n"
							return
						}
						//Actualizar el super bloque
						super_bloque.S_first_ino += int32(binary.Size(nuevo_inodo))
						super_bloque.S_free_inodes_count -= 1
						archivo.Seek(int64(pos_sb), 0)
						err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
						if err != nil {
							consola_mkdir += "[-ERROR-] Error al escribir el Super Bloque\n"
							return
						}
						//Actualizar el bitmap de inodos
						pos_bitmap = super_bloque.S_inodes_count - super_bloque.S_free_inodes_count
						archivo.Seek(int64(super_bloque.S_bm_inode_start+pos_bitmap), 0)
						err = binary.Write(archivo, binary.LittleEndian, &uno)
						if err != nil {
							consola_mkdir += "[-ERROR-] Error al escribir el Bitmap de Inodos\n"
							return
						}
						//Actualizar el bloque de carpeta
						siguiente_bloque.B_content[2].B_inodo = int32(pos_nuevo_inodo)
						copy(siguiente_bloque.B_content[2].B_name[:], nueva_carpeta)
						archivo.Seek(int64(pos_sig_bloque), 0)
						err = binary.Write(archivo, binary.LittleEndian, &siguiente_bloque)
						if err != nil {
							consola_mkdir += "[-ERROR-] Error al escribir el Bloque\n"
							return
						}
						//Crea el bloque carpeta
						nuevo_bloque := Bloque_Carpeta{}
						nuevo_bloque.B_content[0].B_inodo = int32(pos_nuevo_inodo)
						copy(nuevo_bloque.B_content[0].B_name[:], nueva_carpeta)
						nuevo_bloque.B_content[1].B_inodo = int32(Bloque_prueba.B_content[0].B_inodo)
						copy(nuevo_bloque.B_content[1].B_name[:], Bloque_prueba.B_content[0].B_name[:])
						for j := 2; j < 4; j++ {
							nuevo_bloque.B_content[j].B_inodo = -1
						}
						//Escribir el nuevo bloque
						pos_nuevo_bloque := super_bloque.S_first_blo
						archivo.Seek(int64(pos_nuevo_bloque), 0)
						err = binary.Write(archivo, binary.LittleEndian, &nuevo_bloque)
						if err != nil {
							consola_mkdir += "[-ERROR-] Error al escribir el Bloque\n"
							return
						}
						//Actualizar el super bloque
						super_bloque.S_first_blo += int32(binary.Size(nuevo_bloque))
						super_bloque.S_free_blocks_count -= 1
						archivo.Seek(int64(pos_sb), 0)
						err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
						if err != nil {
							consola_mkdir += "[-ERROR-] Error al escribir el Super Bloque\n"
							return
						}
						//Actualizar el bitmap de bloques
						pos_bitmap = super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
						archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
						err = binary.Write(archivo, binary.LittleEndian, &uno)
						if err != nil {
							consola_mkdir += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
							return
						}

						consola_mkdir += "[*SUCCESS*] Se ha creado la carpeta " + nueva_carpeta + " correctamente\n"
						return
					}
				}
			}
		}
	}

}
func (mkdir *Mkdir) ExisteCarpetaPadre(names []string, pos int, path string) bool {
	fmt.Println("NAMES: ", names)
	fmt.Println("POS: ", pos)
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
			fmt.Println("Bloque: ", inodo.I_block[i])
			bloque := Bloque_Carpeta{}
			archivo.Seek(int64(inodo.I_block[i]), 0)
			err = binary.Read(archivo, binary.LittleEndian, &bloque)
			if err != nil {
				consola_mkdir += "[-ERROR-] Error al leer el Bloque Carpeta\n"
				return false
			}
			for j := 0; j < len(bloque.B_content); j++ {
				if bloque.B_content[j].B_inodo != -1 {
					name_comp := string(bloque.B_content[j].B_name[:])
					name_comp = strings.ReplaceAll(name_comp, "\u0000", "")
					if name_comp == names[0] {
						if len(names) == 1 {
							fmt.Println(true)
							return true
						} else {
							retornar := mkdir.ExisteCarpetaPadre(names[1:], int(bloque.B_content[j].B_inodo), path)
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
					name_com := string(bloque.B_content[j].B_name[:])
					name_com = strings.ReplaceAll(name_com, "\u0000", "")
					if name_com == names[0] {
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
