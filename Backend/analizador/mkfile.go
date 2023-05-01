package analizador

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

type Mkfile struct {
	Path string
	R    bool
	Size int
	Cont string
}

var consola_mkfile string

func (mkfile *Mkfile) VerificarParams(parametros map[string]string) {
	consola_mkfile = ""
	if mkfile.Path == "" {
		consola_mkfile += "[-ERROR-] Falta el parametro path\n"
		return
	}

	if mkfile.Size < 0 {
		consola_mkfile += "[-ERROR-] No se aceptan numeros negativos en el Size\n"
		return
	}

	//Dando prioridad al Cont
	if mkfile.Size > 0 && mkfile.Cont != "" {
		mkfile.Size = 0
	}

	if mkfile.Path[0] == '"' {
		mkfile.Path = mkfile.Path[1 : len(mkfile.Path)-1]
	}

	if mkfile.Size > 1024 {
		consola_mkfile += "[-ERROR-] El tamaño del archivo no puede ser mayor a 1MB\n"
		return
	}

	//Verificando si el archivo existe
	if mkfile.Cont != "" {
		if mkfile.Cont[0] == '"' {
			mkfile.Cont = mkfile.Cont[1 : len(mkfile.Cont)-1]
		}
		if !mkfile.ExisteArchivo() {
			consola_mkfile += "[-ERROR-] El archivo registrado en el paramentro CONT no existe\n"
			return
		}
	}
	mkfile.CrearArchivo()
}

func (mkfile *Mkfile) CrearArchivo() {
	montada := mkfile.RetornarStrictMontada(Idlogueado)
	if mkfile.IsParticionMontadaVacia(montada) {
		consola_mkfile += "[-ERROR-] La partición con id: " + Idlogueado + " no está montada\n"
		return
	}

	if !montada.Sistema_archivos {
		consola_mkfile += "[-ERROR-] La partición con id: " + Idlogueado + " no tiene un sistema de archivos\n"
	}

	//Abrir el archivo binario
	archivo, err := os.OpenFile(montada.Path, os.O_RDWR, 0666)
	if err != nil {
		consola_mkfile += "[-ERROR-] Error al abrir el archivo\n"
		return
	}
	defer archivo.Close()

	//Leer el MBR
	mbr := MBR{}
	archivo.Seek(int64(0), 0)
	err = binary.Read(archivo, binary.LittleEndian, &mbr)
	if err != nil {
		consola_mkfile += "[-ERROR-] Error al leer el MBR\n"
		return
	}
	fmt.Println("MBR DESDE REP")
	fmt.Println(mbr)
	particiones := mkfile.ObtenerParticiones(mbr)
	var ebrs []EBR
	logica_existe := false
	var particion_logica EBR
	for i := 0; i < len(particiones); i++ {
		if strings.ToLower(string(particiones[i].Part_type[0])) == "e" {
			ebrs = mkfile.ListadoEBR(particiones[i], montada.Path)
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
		consola_mkfile += "[-ERROR-] Error al leer el SuperBloque\n"
		return
	}

	if mkfile.R {
		if mkfile.Cont != "" {
			mkfile.CrearArchivoComputadoraRecursivo(pos_inicio, montada.Path)
		} else if mkfile.Size > 0 {
			mkfile.CrearArchivoConTamañoRecursivo(pos_inicio, montada.Path)
		} else {
			mkfile.CrearArchivoRecursivo(pos_inicio, montada.Path)
		}
	} else {
		if mkfile.Cont != "" {
			mkfile.CrearArchivoComputadora(pos_inicio, montada.Path)
		} else if mkfile.Size > 0 {
			mkfile.CrearArchivoConTamaño(pos_inicio, montada.Path)
		} else {
			mkfile.CrearArchivoNoRecursivo(pos_inicio, montada.Path)
		}
	}

}

func (mkfile *Mkfile) CrearArchivoComputadoraRecursivo(pos_sb int, path string) {
	archivo, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		consola_mkfile += "[-ERROR-] Error al abrir el archivo\n"
		return
	}
	defer archivo.Close()
	name_carpetas := strings.Split(mkfile.Path, "/")
	if name_carpetas[0] == "" {
		name_carpetas[0] = "/"
	} else {
		consola_mkfile += "[-ERROR-] La ruta no es absoluta\n"
	}
	nombre_archivo := name_carpetas[len(name_carpetas)-1]
	name_carpetas = name_carpetas[:len(name_carpetas)-1]

	//Leer el SuperBloque
	archivo.Seek(int64(pos_sb), 0)
	super_bloque := SuperBloque{}
	err = binary.Read(archivo, binary.LittleEndian, &super_bloque)
	if err != nil {
		consola_mkfile += "[-ERROR-] Error al leer el SuperBloque\n"
		return
	}

	if !mkfile.ExisteCarpetaPadre(name_carpetas, int(super_bloque.S_inode_start), path) {
		consola_mkfile += "[-ERROR-] No existe la carpeta donde quieres agregar el archivo\n"
		return
	}

	//Leer el Inodo de la carpeta
	posicion_carpeta := mkfile.PosCarpetaPadre(name_carpetas, int(super_bloque.S_inode_start), path)
	fmt.Println("Posicion padre: ", posicion_carpeta)
	archivo.Seek(int64(posicion_carpeta), 0)
	inodo_carpeta := Inodo{}
	err = binary.Read(archivo, binary.LittleEndian, &inodo_carpeta)
	if err != nil {
		consola_mkfile += "[-ERROR-] Error al leer el Inodo\n"
		return
	}
	//nombrare en que carpeta estoy
	archivo.Seek(int64(inodo_carpeta.I_block[0]), 0)
	Bloque_prueba := Bloque_Carpeta{}
	binary.Read(archivo, binary.LittleEndian, &Bloque_prueba)
	fmt.Println("Nombre de la carpeta actual: ", string(Bloque_prueba.B_content[0].B_name[:]))
	tiempo := time.Now()
	tiempoFormateado := tiempo.Format("2006-01-02 15:04:05")

	for i := 0; i < len(inodo_carpeta.I_block); i++ {
		if i != 15 {
			no_bloque := int(inodo_carpeta.I_block[i])
			if no_bloque != -1 {
				if inodo_carpeta.I_block[i+1] == -1 {
					bloque_c := Bloque_Carpeta{}
					archivo.Seek(int64(no_bloque), 0)
					err = binary.Read(archivo, binary.LittleEndian, &bloque_c)
					if err != nil {
						consola_mkfile += "[-ERROR-] Error al leer el Bloque\n"
						return
					}
					var hay_espacio bool = false
					var pos_b int = 0
					for j := 2; j < 4; j++ {
						if bloque_c.B_content[j].B_inodo == -1 {
							hay_espacio = true
							pos_b = j
							break
						}
					}
					//Crea el nuevo bloque Archivo
					contenido := mkfile.RetornarContenidoArchivoComputadora()
					if len(contenido) > 1024 {
						consola_mkfile += "[-ERROR-] El contenido del archivo es mayor a 1Kb\n"
						return
					}
					fmt.Println("CONTENIDO DEL ARCHIVO")
					fmt.Println(contenido)
					if contenido == "" {
						return
					}
					if hay_espacio {
						//Crear el inodo de archivo
						nuevo_inodo := Inodo{}
						nuevo_inodo.I_uid = int32(Id_UserLogueado)
						nuevo_inodo.I_gid = int32(Id_GroupLogueado)
						nuevo_inodo.I_size = int32(len(contenido))
						copy(nuevo_inodo.I_atime[:], tiempoFormateado)
						copy(nuevo_inodo.I_ctime[:], tiempoFormateado)
						copy(nuevo_inodo.I_mtime[:], tiempoFormateado)
						nuevo_inodo.I_type = 1
						nuevo_inodo.I_perm = 664
						for j := 0; j < 16; j++ {
							nuevo_inodo.I_block[j] = -1
						}
						nuevo_inodo.I_block[0] = super_bloque.S_first_blo
						//Escribir el inodo en el archivo
						posicion_nuevo_inodo := super_bloque.S_first_ino
						archivo.Seek(int64(posicion_nuevo_inodo), 0)
						err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
							return
						}
						//Actualizar el super bloque
						super_bloque.S_first_ino += int32(binary.Size(nuevo_inodo))
						super_bloque.S_free_inodes_count -= 1
						archivo.Seek(int64(pos_sb), 0)
						err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
							return
						}
						//Actualizar el bitmap de inodos
						pos_bitmap := super_bloque.S_inodes_count - super_bloque.S_free_inodes_count
						archivo.Seek(int64(super_bloque.S_bm_inode_start+pos_bitmap), 0)
						var uno byte = 1
						err = binary.Write(archivo, binary.LittleEndian, &uno)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Inodos\n"
							return
						}
						//Actualizar el bloque de carpetas
						bloque_c.B_content[pos_b].B_inodo = int32(posicion_nuevo_inodo)
						copy(bloque_c.B_content[pos_b].B_name[:], nombre_archivo)
						archivo.Seek(int64(no_bloque), 0)
						err = binary.Write(archivo, binary.LittleEndian, &bloque_c)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bloque\n"
							return
						}

						if len(contenido) < 64 {
							//Creo el primer bloque_archivo
							bloque := Bloque_Archivo{}
							copy(bloque.B_content[:], contenido)
							archivo.Seek(int64(super_bloque.S_first_blo), 0)
							err = binary.Write(archivo, binary.LittleEndian, &bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
								return
							}
							//Actualizar el super bloque
							super_bloque.S_first_blo += int32(binary.Size(bloque))
							super_bloque.S_free_blocks_count -= 1
							archivo.Seek(int64(pos_sb), 0)
							err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
								return
							}
							//Actualizar el bitmap de bloques
							pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
							archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
							var uno byte = 1
							err = binary.Write(archivo, binary.LittleEndian, &uno)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
								return
							}
							consola_mkfile += "[*SUCCESS*] Se ha creado el archivo " + nombre_archivo + " correctamente\n"
							return
						} else {
							//Creo el primer bloque_archivo
							bloque := Bloque_Archivo{}
							copy(bloque.B_content[:], contenido[:64])
							contenido = contenido[64:]
							archivo.Seek(int64(super_bloque.S_first_blo), 0)
							err = binary.Write(archivo, binary.LittleEndian, &bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
								return
							}
							//Actualizar el super bloque
							super_bloque.S_first_blo += int32(binary.Size(bloque))
							super_bloque.S_free_blocks_count -= 1
							archivo.Seek(int64(pos_sb), 0)
							err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
								return
							}
							//Actualizar el bitmap de bloques
							pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
							archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
							var uno byte = 1
							err = binary.Write(archivo, binary.LittleEndian, &uno)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
								return
							}
							for j := 0; j < len(nuevo_inodo.I_block); j++ {
								no_bloque2 := int64(nuevo_inodo.I_block[j])
								if no_bloque2 == -1 {
									break
								}
								if nuevo_inodo.I_block[j+1] == -1 {
									if len(contenido) > 64 {
										bloque := Bloque_Archivo{}
										copy(bloque.B_content[:], contenido[:64])
										contenido = contenido[64:]
										archivo.Seek(int64(super_bloque.S_first_blo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
											return
										}
										//Actualizo el inodo
										nuevo_inodo.I_block[j+1] = int32(super_bloque.S_first_blo)
										archivo.Seek(int64(posicion_nuevo_inodo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
											return
										}

										//Actualizar el super bloque
										super_bloque.S_first_blo += int32(binary.Size(bloque))
										super_bloque.S_free_blocks_count -= 1
										archivo.Seek(int64(pos_sb), 0)
										err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
											return
										}
										//Actualizar el bitmap de bloques
										pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
										archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
										var uno byte = 1
										err = binary.Write(archivo, binary.LittleEndian, &uno)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
											return
										}
									} else {
										bloque := Bloque_Archivo{}
										copy(bloque.B_content[:], contenido)
										archivo.Seek(int64(super_bloque.S_first_blo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
											return
										}
										//Actualizo el inodo
										nuevo_inodo.I_block[j+1] = int32(super_bloque.S_first_blo)
										archivo.Seek(int64(posicion_nuevo_inodo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
											return
										}

										//Actualizar el super bloque
										super_bloque.S_first_blo += int32(binary.Size(bloque))
										super_bloque.S_free_blocks_count -= 1
										archivo.Seek(int64(pos_sb), 0)
										err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
											return
										}
										//Actualizar el bitmap de bloques
										pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
										archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
										var uno byte = 1
										err = binary.Write(archivo, binary.LittleEndian, &uno)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
											return
										}
									}
								}
							}
						}
						consola_mkfile += "[*SUCCESS*] Se ha creado el archivo: " + nombre_archivo + " exitosamente\n"
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
							consola_mkfile += "[-ERROR-] Error al escribir el Bloque\n"
							return
						}
						//Actualizar el super bloque
						super_bloque.S_first_blo += int32(binary.Size(siguiente_bloque))
						super_bloque.S_free_blocks_count -= 1
						archivo.Seek(int64(pos_sb), 0)
						err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
							return
						}
						var uno byte = 1
						//Actualizar el bitmap de bloques
						pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
						archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
						err = binary.Write(archivo, binary.LittleEndian, &uno)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
							return
						}
						//Actualizar el inodo padre
						inodo_carpeta.I_block[i+1] = int32(pos_sig_bloque)
						archivo.Seek(int64(Bloque_prueba.B_content[0].B_inodo), 0)
						err = binary.Write(archivo, binary.LittleEndian, &inodo_carpeta)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
							return
						}

						//Crear el inodo de archivo
						nuevo_inodo := Inodo{}
						nuevo_inodo.I_uid = int32(Id_UserLogueado)
						nuevo_inodo.I_gid = int32(Id_GroupLogueado)
						nuevo_inodo.I_size = int32(len(contenido))
						copy(nuevo_inodo.I_atime[:], tiempoFormateado)
						copy(nuevo_inodo.I_ctime[:], tiempoFormateado)
						copy(nuevo_inodo.I_mtime[:], tiempoFormateado)
						nuevo_inodo.I_type = 1
						nuevo_inodo.I_perm = 664
						for j := 0; j < 16; j++ {
							nuevo_inodo.I_block[j] = -1
						}
						nuevo_inodo.I_block[0] = super_bloque.S_first_blo
						//Escribir el inodo en el archivo
						posicion_nuevo_inodo := super_bloque.S_first_ino
						archivo.Seek(int64(posicion_nuevo_inodo), 0)
						err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
							return
						}
						//Actualizar el super bloque
						super_bloque.S_first_ino += int32(binary.Size(nuevo_inodo))
						super_bloque.S_free_inodes_count -= 1
						archivo.Seek(int64(pos_sb), 0)
						err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
							return
						}
						//Actualizar el bitmap de inodos
						pos_bitmap = super_bloque.S_inodes_count - super_bloque.S_free_inodes_count
						archivo.Seek(int64(super_bloque.S_bm_inode_start+pos_bitmap), 0)

						err = binary.Write(archivo, binary.LittleEndian, &uno)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Inodos\n"
							return
						}
						//Actualizar el bloque de carpetas
						siguiente_bloque.B_content[2].B_inodo = int32(posicion_nuevo_inodo)
						copy(siguiente_bloque.B_content[2].B_name[:], nombre_archivo)
						archivo.Seek(int64(pos_sig_bloque), 0)
						err = binary.Write(archivo, binary.LittleEndian, &siguiente_bloque)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bloque\n"
							return
						}

						if mkfile.Size < 64 {
							//Creo el primer bloque_archivo
							bloque := Bloque_Archivo{}
							copy(bloque.B_content[:], contenido)
							archivo.Seek(int64(super_bloque.S_first_blo), 0)
							err = binary.Write(archivo, binary.LittleEndian, &bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
								return
							}
							//Actualizar el super bloque
							super_bloque.S_first_blo += int32(binary.Size(bloque))
							super_bloque.S_free_blocks_count -= 1
							archivo.Seek(int64(pos_sb), 0)
							err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
								return
							}
							//Actualizar el bitmap de bloques
							pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
							archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
							var uno byte = 1
							err = binary.Write(archivo, binary.LittleEndian, &uno)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
								return
							}
							consola_mkfile += "[*SUCCESS*] Se ha creado el archivo " + nombre_archivo + " correctamente\n"
							return
						} else {
							//Creo el primer bloque_archivo
							bloque := Bloque_Archivo{}
							copy(bloque.B_content[:], contenido[:64])
							contenido = contenido[64:]
							archivo.Seek(int64(super_bloque.S_first_blo), 0)
							err = binary.Write(archivo, binary.LittleEndian, &bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
								return
							}
							//Actualizar el super bloque
							super_bloque.S_first_blo += int32(binary.Size(bloque))
							super_bloque.S_free_blocks_count -= 1
							archivo.Seek(int64(pos_sb), 0)
							err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
								return
							}
							//Actualizar el bitmap de bloques
							pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
							archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
							var uno byte = 1
							err = binary.Write(archivo, binary.LittleEndian, &uno)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
								return
							}
							for j := 0; j < len(nuevo_inodo.I_block); j++ {
								no_bloque2 := int64(nuevo_inodo.I_block[j])
								if no_bloque2 == -1 {
									break
								}
								if nuevo_inodo.I_block[j+1] == -1 {
									if len(contenido) > 64 {
										bloque := Bloque_Archivo{}
										copy(bloque.B_content[:], contenido[:64])
										contenido = contenido[64:]
										archivo.Seek(int64(super_bloque.S_first_blo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
											return
										}
										//Actualizo el inodo
										nuevo_inodo.I_block[j+1] = int32(super_bloque.S_first_blo)
										archivo.Seek(int64(posicion_nuevo_inodo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
											return
										}

										//Actualizar el super bloque
										super_bloque.S_first_blo += int32(binary.Size(bloque))
										super_bloque.S_free_blocks_count -= 1
										archivo.Seek(int64(pos_sb), 0)
										err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
											return
										}
										//Actualizar el bitmap de bloques
										pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
										archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
										var uno byte = 1
										err = binary.Write(archivo, binary.LittleEndian, &uno)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
											return
										}
									} else {
										bloque := Bloque_Archivo{}
										copy(bloque.B_content[:], contenido)
										archivo.Seek(int64(super_bloque.S_first_blo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
											return
										}
										//Actualizo el inodo
										nuevo_inodo.I_block[j+1] = int32(super_bloque.S_first_blo)
										archivo.Seek(int64(posicion_nuevo_inodo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
											return
										}

										//Actualizar el super bloque
										super_bloque.S_first_blo += int32(binary.Size(bloque))
										super_bloque.S_free_blocks_count -= 1
										archivo.Seek(int64(pos_sb), 0)
										err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
											return
										}
										//Actualizar el bitmap de bloques
										pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
										archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
										var uno byte = 1
										err = binary.Write(archivo, binary.LittleEndian, &uno)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
											return
										}
									}
								}
							}
						}
						consola_mkfile += "[*SUCCESS*] Se ha creado el archivo: " + nombre_archivo + " exitosamente\n"
						return
					}
				}
			}
		}
	}

}

func (mkfile *Mkfile) CrearArchivoComputadora(pos_sb int, path string) {
	archivo, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		consola_mkfile += "[-ERROR-] Error al abrir el archivo\n"
		return
	}
	defer archivo.Close()
	name_carpetas := strings.Split(mkfile.Path, "/")
	if name_carpetas[0] == "" {
		name_carpetas[0] = "/"
	} else {
		consola_mkfile += "[-ERROR-] La ruta no es absoluta\n"
	}
	nombre_archivo := name_carpetas[len(name_carpetas)-1]
	name_carpetas = name_carpetas[:len(name_carpetas)-1]
	mkfile.CreacionRecursiva(name_carpetas, pos_sb, path)

	//Leer el SuperBloque
	archivo.Seek(int64(pos_sb), 0)
	super_bloque := SuperBloque{}
	err = binary.Read(archivo, binary.LittleEndian, &super_bloque)
	if err != nil {
		consola_mkfile += "[-ERROR-] Error al leer el SuperBloque\n"
		return
	}

	//Leer el Inodo de la carpeta
	posicion_carpeta := mkfile.PosCarpetaPadre(name_carpetas, int(super_bloque.S_inode_start), path)
	fmt.Println("Posicion padre: ", posicion_carpeta)
	archivo.Seek(int64(posicion_carpeta), 0)
	inodo_carpeta := Inodo{}
	err = binary.Read(archivo, binary.LittleEndian, &inodo_carpeta)
	if err != nil {
		consola_mkfile += "[-ERROR-] Error al leer el Inodo\n"
		return
	}
	//nombrare en que carpeta estoy
	archivo.Seek(int64(inodo_carpeta.I_block[0]), 0)
	Bloque_prueba := Bloque_Carpeta{}
	binary.Read(archivo, binary.LittleEndian, &Bloque_prueba)
	fmt.Println("Nombre de la carpeta actual: ", string(Bloque_prueba.B_content[0].B_name[:]))
	tiempo := time.Now()
	tiempoFormateado := tiempo.Format("2006-01-02 15:04:05")

	for i := 0; i < len(inodo_carpeta.I_block); i++ {
		if i != 15 {
			no_bloque := int(inodo_carpeta.I_block[i])
			if no_bloque != -1 {
				if inodo_carpeta.I_block[i+1] == -1 {
					bloque_c := Bloque_Carpeta{}
					archivo.Seek(int64(no_bloque), 0)
					err = binary.Read(archivo, binary.LittleEndian, &bloque_c)
					if err != nil {
						consola_mkfile += "[-ERROR-] Error al leer el Bloque\n"
						return
					}
					var hay_espacio bool = false
					var pos_b int = 0
					for j := 2; j < 4; j++ {
						if bloque_c.B_content[j].B_inodo == -1 {
							hay_espacio = true
							pos_b = j
							break
						}
					}
					//Crea el nuevo bloque Archivo
					contenido := mkfile.RetornarContenidoArchivoComputadora()
					if len(contenido) > 1024 {
						consola_mkfile += "[-ERROR-] El contenido del archivo es mayor a 1Kb\n"
						return
					}
					fmt.Println("CONTENIDO DEL ARCHIVO")
					fmt.Println(contenido)
					if contenido == "" {
						return
					}
					if hay_espacio {
						//Crear el inodo de archivo
						nuevo_inodo := Inodo{}
						nuevo_inodo.I_uid = int32(Id_UserLogueado)
						nuevo_inodo.I_gid = int32(Id_GroupLogueado)
						nuevo_inodo.I_size = int32(len(contenido))
						copy(nuevo_inodo.I_atime[:], tiempoFormateado)
						copy(nuevo_inodo.I_ctime[:], tiempoFormateado)
						copy(nuevo_inodo.I_mtime[:], tiempoFormateado)
						nuevo_inodo.I_type = 1
						nuevo_inodo.I_perm = 664
						for j := 0; j < 16; j++ {
							nuevo_inodo.I_block[j] = -1
						}
						nuevo_inodo.I_block[0] = super_bloque.S_first_blo
						//Escribir el inodo en el archivo
						posicion_nuevo_inodo := super_bloque.S_first_ino
						archivo.Seek(int64(posicion_nuevo_inodo), 0)
						err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
							return
						}
						//Actualizar el super bloque
						super_bloque.S_first_ino += int32(binary.Size(nuevo_inodo))
						super_bloque.S_free_inodes_count -= 1
						archivo.Seek(int64(pos_sb), 0)
						err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
							return
						}
						//Actualizar el bitmap de inodos
						pos_bitmap := super_bloque.S_inodes_count - super_bloque.S_free_inodes_count
						archivo.Seek(int64(super_bloque.S_bm_inode_start+pos_bitmap), 0)
						var uno byte = 1
						err = binary.Write(archivo, binary.LittleEndian, &uno)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Inodos\n"
							return
						}
						//Actualizar el bloque de carpetas
						bloque_c.B_content[pos_b].B_inodo = int32(posicion_nuevo_inodo)
						copy(bloque_c.B_content[pos_b].B_name[:], nombre_archivo)
						archivo.Seek(int64(no_bloque), 0)
						err = binary.Write(archivo, binary.LittleEndian, &bloque_c)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bloque\n"
							return
						}

						if len(contenido) < 64 {
							//Creo el primer bloque_archivo
							bloque := Bloque_Archivo{}
							copy(bloque.B_content[:], contenido)
							archivo.Seek(int64(super_bloque.S_first_blo), 0)
							err = binary.Write(archivo, binary.LittleEndian, &bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
								return
							}
							//Actualizar el super bloque
							super_bloque.S_first_blo += int32(binary.Size(bloque))
							super_bloque.S_free_blocks_count -= 1
							archivo.Seek(int64(pos_sb), 0)
							err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
								return
							}
							//Actualizar el bitmap de bloques
							pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
							archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
							var uno byte = 1
							err = binary.Write(archivo, binary.LittleEndian, &uno)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
								return
							}
							consola_mkfile += "[*SUCCESS*] Se ha creado el archivo " + nombre_archivo + " correctamente\n"
							return
						} else {
							//Creo el primer bloque_archivo
							bloque := Bloque_Archivo{}
							copy(bloque.B_content[:], contenido[:64])
							contenido = contenido[64:]
							archivo.Seek(int64(super_bloque.S_first_blo), 0)
							err = binary.Write(archivo, binary.LittleEndian, &bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
								return
							}
							//Actualizar el super bloque
							super_bloque.S_first_blo += int32(binary.Size(bloque))
							super_bloque.S_free_blocks_count -= 1
							archivo.Seek(int64(pos_sb), 0)
							err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
								return
							}
							//Actualizar el bitmap de bloques
							pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
							archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
							var uno byte = 1
							err = binary.Write(archivo, binary.LittleEndian, &uno)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
								return
							}
							for j := 0; j < len(nuevo_inodo.I_block); j++ {
								no_bloque2 := int64(nuevo_inodo.I_block[j])
								if no_bloque2 == -1 {
									break
								}
								if nuevo_inodo.I_block[j+1] == -1 {
									if len(contenido) > 64 {
										bloque := Bloque_Archivo{}
										copy(bloque.B_content[:], contenido[:64])
										contenido = contenido[64:]
										archivo.Seek(int64(super_bloque.S_first_blo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
											return
										}
										//Actualizo el inodo
										nuevo_inodo.I_block[j+1] = int32(super_bloque.S_first_blo)
										archivo.Seek(int64(posicion_nuevo_inodo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
											return
										}

										//Actualizar el super bloque
										super_bloque.S_first_blo += int32(binary.Size(bloque))
										super_bloque.S_free_blocks_count -= 1
										archivo.Seek(int64(pos_sb), 0)
										err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
											return
										}
										//Actualizar el bitmap de bloques
										pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
										archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
										var uno byte = 1
										err = binary.Write(archivo, binary.LittleEndian, &uno)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
											return
										}
									} else {
										bloque := Bloque_Archivo{}
										copy(bloque.B_content[:], contenido)
										archivo.Seek(int64(super_bloque.S_first_blo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
											return
										}
										//Actualizo el inodo
										nuevo_inodo.I_block[j+1] = int32(super_bloque.S_first_blo)
										archivo.Seek(int64(posicion_nuevo_inodo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
											return
										}

										//Actualizar el super bloque
										super_bloque.S_first_blo += int32(binary.Size(bloque))
										super_bloque.S_free_blocks_count -= 1
										archivo.Seek(int64(pos_sb), 0)
										err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
											return
										}
										//Actualizar el bitmap de bloques
										pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
										archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
										var uno byte = 1
										err = binary.Write(archivo, binary.LittleEndian, &uno)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
											return
										}
									}
								}
							}
						}
						consola_mkfile += "[*SUCCESS*] Se ha creado el archivo: " + nombre_archivo + " exitosamente\n"
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
							consola_mkfile += "[-ERROR-] Error al escribir el Bloque\n"
							return
						}
						//Actualizar el super bloque
						super_bloque.S_first_blo += int32(binary.Size(siguiente_bloque))
						super_bloque.S_free_blocks_count -= 1
						archivo.Seek(int64(pos_sb), 0)
						err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
							return
						}
						var uno byte = 1
						//Actualizar el bitmap de bloques
						pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
						archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
						err = binary.Write(archivo, binary.LittleEndian, &uno)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
							return
						}
						//Actualizar el inodo padre
						inodo_carpeta.I_block[i+1] = int32(pos_sig_bloque)
						archivo.Seek(int64(Bloque_prueba.B_content[0].B_inodo), 0)
						err = binary.Write(archivo, binary.LittleEndian, &inodo_carpeta)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
							return
						}

						//Crear el inodo de archivo
						nuevo_inodo := Inodo{}
						nuevo_inodo.I_uid = int32(Id_UserLogueado)
						nuevo_inodo.I_gid = int32(Id_GroupLogueado)
						nuevo_inodo.I_size = int32(len(contenido))
						copy(nuevo_inodo.I_atime[:], tiempoFormateado)
						copy(nuevo_inodo.I_ctime[:], tiempoFormateado)
						copy(nuevo_inodo.I_mtime[:], tiempoFormateado)
						nuevo_inodo.I_type = 1
						nuevo_inodo.I_perm = 664
						for j := 0; j < 16; j++ {
							nuevo_inodo.I_block[j] = -1
						}
						nuevo_inodo.I_block[0] = super_bloque.S_first_blo
						//Escribir el inodo en el archivo
						posicion_nuevo_inodo := super_bloque.S_first_ino
						archivo.Seek(int64(posicion_nuevo_inodo), 0)
						err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
							return
						}
						//Actualizar el super bloque
						super_bloque.S_first_ino += int32(binary.Size(nuevo_inodo))
						super_bloque.S_free_inodes_count -= 1
						archivo.Seek(int64(pos_sb), 0)
						err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
							return
						}
						//Actualizar el bitmap de inodos
						pos_bitmap = super_bloque.S_inodes_count - super_bloque.S_free_inodes_count
						archivo.Seek(int64(super_bloque.S_bm_inode_start+pos_bitmap), 0)

						err = binary.Write(archivo, binary.LittleEndian, &uno)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Inodos\n"
							return
						}
						//Actualizar el bloque de carpetas
						siguiente_bloque.B_content[2].B_inodo = int32(posicion_nuevo_inodo)
						copy(siguiente_bloque.B_content[2].B_name[:], nombre_archivo)
						archivo.Seek(int64(pos_sig_bloque), 0)
						err = binary.Write(archivo, binary.LittleEndian, &siguiente_bloque)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bloque\n"
							return
						}

						if mkfile.Size < 64 {
							//Creo el primer bloque_archivo
							bloque := Bloque_Archivo{}
							copy(bloque.B_content[:], contenido)
							archivo.Seek(int64(super_bloque.S_first_blo), 0)
							err = binary.Write(archivo, binary.LittleEndian, &bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
								return
							}
							//Actualizar el super bloque
							super_bloque.S_first_blo += int32(binary.Size(bloque))
							super_bloque.S_free_blocks_count -= 1
							archivo.Seek(int64(pos_sb), 0)
							err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
								return
							}
							//Actualizar el bitmap de bloques
							pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
							archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
							var uno byte = 1
							err = binary.Write(archivo, binary.LittleEndian, &uno)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
								return
							}
							consola_mkfile += "[*SUCCESS*] Se ha creado el archivo " + nombre_archivo + " correctamente\n"
							return
						} else {
							//Creo el primer bloque_archivo
							bloque := Bloque_Archivo{}
							copy(bloque.B_content[:], contenido[:64])
							contenido = contenido[64:]
							archivo.Seek(int64(super_bloque.S_first_blo), 0)
							err = binary.Write(archivo, binary.LittleEndian, &bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
								return
							}
							//Actualizar el super bloque
							super_bloque.S_first_blo += int32(binary.Size(bloque))
							super_bloque.S_free_blocks_count -= 1
							archivo.Seek(int64(pos_sb), 0)
							err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
								return
							}
							//Actualizar el bitmap de bloques
							pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
							archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
							var uno byte = 1
							err = binary.Write(archivo, binary.LittleEndian, &uno)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
								return
							}
							for j := 0; j < len(nuevo_inodo.I_block); j++ {
								no_bloque2 := int64(nuevo_inodo.I_block[j])
								if no_bloque2 == -1 {
									break
								}
								if nuevo_inodo.I_block[j+1] == -1 {
									if len(contenido) > 64 {
										bloque := Bloque_Archivo{}
										copy(bloque.B_content[:], contenido[:64])
										contenido = contenido[64:]
										archivo.Seek(int64(super_bloque.S_first_blo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
											return
										}
										//Actualizo el inodo
										nuevo_inodo.I_block[j+1] = int32(super_bloque.S_first_blo)
										archivo.Seek(int64(posicion_nuevo_inodo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
											return
										}

										//Actualizar el super bloque
										super_bloque.S_first_blo += int32(binary.Size(bloque))
										super_bloque.S_free_blocks_count -= 1
										archivo.Seek(int64(pos_sb), 0)
										err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
											return
										}
										//Actualizar el bitmap de bloques
										pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
										archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
										var uno byte = 1
										err = binary.Write(archivo, binary.LittleEndian, &uno)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
											return
										}
									} else {
										bloque := Bloque_Archivo{}
										copy(bloque.B_content[:], contenido)
										archivo.Seek(int64(super_bloque.S_first_blo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
											return
										}
										//Actualizo el inodo
										nuevo_inodo.I_block[j+1] = int32(super_bloque.S_first_blo)
										archivo.Seek(int64(posicion_nuevo_inodo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
											return
										}

										//Actualizar el super bloque
										super_bloque.S_first_blo += int32(binary.Size(bloque))
										super_bloque.S_free_blocks_count -= 1
										archivo.Seek(int64(pos_sb), 0)
										err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
											return
										}
										//Actualizar el bitmap de bloques
										pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
										archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
										var uno byte = 1
										err = binary.Write(archivo, binary.LittleEndian, &uno)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
											return
										}
									}
								}
							}
						}
						consola_mkfile += "[*SUCCESS*] Se ha creado el archivo: " + nombre_archivo + " exitosamente\n"
						return
					}
				}
			}
		}
	}
}

func (mkfile *Mkfile) CrearArchivoConTamañoRecursivo(pos_sb int, path string) {
	archivo, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		consola_mkfile += "[-ERROR-] Error al abrir el archivo\n"
		return
	}
	defer archivo.Close()
	name_carpetas := strings.Split(mkfile.Path, "/")
	if name_carpetas[0] == "" {
		name_carpetas[0] = "/"
	} else {
		consola_mkfile += "[-ERROR-] La ruta no es absoluta\n"
	}
	nombre_archivo := name_carpetas[len(name_carpetas)-1]
	name_carpetas = name_carpetas[:len(name_carpetas)-1]
	mkfile.CreacionRecursiva(name_carpetas, pos_sb, path)

	//Leer el SuperBloque
	archivo.Seek(int64(pos_sb), 0)
	super_bloque := SuperBloque{}
	err = binary.Read(archivo, binary.LittleEndian, &super_bloque)
	if err != nil {
		consola_mkfile += "[-ERROR-] Error al leer el SuperBloque\n"
		return
	}

	//Leer el Inodo de la carpeta
	posicion_carpeta := mkfile.PosCarpetaPadre(name_carpetas, int(super_bloque.S_inode_start), path)
	fmt.Println("Posicion padre: ", posicion_carpeta)
	archivo.Seek(int64(posicion_carpeta), 0)
	inodo_carpeta := Inodo{}
	err = binary.Read(archivo, binary.LittleEndian, &inodo_carpeta)
	if err != nil {
		consola_mkfile += "[-ERROR-] Error al leer el Inodo\n"
		return
	}
	//nombrare en que carpeta estoy
	archivo.Seek(int64(inodo_carpeta.I_block[0]), 0)
	Bloque_prueba := Bloque_Carpeta{}
	binary.Read(archivo, binary.LittleEndian, &Bloque_prueba)
	fmt.Println("Nombre de la carpeta actual: ", string(Bloque_prueba.B_content[0].B_name[:]))
	tiempo := time.Now()
	tiempoFormateado := tiempo.Format("2006-01-02 15:04:05")

	for i := 0; i < len(inodo_carpeta.I_block); i++ {
		if i != 15 {
			no_bloque := int(inodo_carpeta.I_block[i])
			if no_bloque != -1 {
				if inodo_carpeta.I_block[i+1] == -1 {
					bloque_c := Bloque_Carpeta{}
					archivo.Seek(int64(no_bloque), 0)
					err = binary.Read(archivo, binary.LittleEndian, &bloque_c)
					if err != nil {
						consola_mkfile += "[-ERROR-] Error al leer el Bloque\n"
						return
					}
					var hay_espacio bool = false
					var pos_b int = 0
					for j := 2; j < 4; j++ {
						if bloque_c.B_content[j].B_inodo == -1 {
							hay_espacio = true
							pos_b = j
							break
						}
					}
					if hay_espacio {
						//Crear el inodo de archivo
						nuevo_inodo := Inodo{}
						nuevo_inodo.I_uid = int32(Id_UserLogueado)
						nuevo_inodo.I_gid = int32(Id_GroupLogueado)
						nuevo_inodo.I_size = int32(mkfile.Size)
						copy(nuevo_inodo.I_atime[:], tiempoFormateado)
						copy(nuevo_inodo.I_ctime[:], tiempoFormateado)
						copy(nuevo_inodo.I_mtime[:], tiempoFormateado)
						nuevo_inodo.I_type = 1
						nuevo_inodo.I_perm = 664
						for j := 0; j < 16; j++ {
							nuevo_inodo.I_block[j] = -1
						}
						nuevo_inodo.I_block[0] = super_bloque.S_first_blo
						//Escribir el inodo en el archivo
						posicion_nuevo_inodo := super_bloque.S_first_ino
						archivo.Seek(int64(posicion_nuevo_inodo), 0)
						err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
							return
						}
						//Actualizar el super bloque
						super_bloque.S_first_ino += int32(binary.Size(nuevo_inodo))
						super_bloque.S_free_inodes_count -= 1
						archivo.Seek(int64(pos_sb), 0)
						err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
							return
						}
						//Actualizar el bitmap de inodos
						pos_bitmap := super_bloque.S_inodes_count - super_bloque.S_free_inodes_count
						archivo.Seek(int64(super_bloque.S_bm_inode_start+pos_bitmap), 0)
						var uno byte = 1
						err = binary.Write(archivo, binary.LittleEndian, &uno)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Inodos\n"
							return
						}
						//Actualizar el bloque de carpetas
						bloque_c.B_content[pos_b].B_inodo = int32(posicion_nuevo_inodo)
						copy(bloque_c.B_content[pos_b].B_name[:], nombre_archivo)
						archivo.Seek(int64(no_bloque), 0)
						err = binary.Write(archivo, binary.LittleEndian, &bloque_c)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bloque\n"
							return
						}
						//Crea el nuevo bloque Archivo
						contenido := mkfile.RetornaContenidoSize()
						if mkfile.Size < 64 {
							//Creo el primer bloque_archivo
							bloque := Bloque_Archivo{}
							copy(bloque.B_content[:], contenido)
							archivo.Seek(int64(super_bloque.S_first_blo), 0)
							err = binary.Write(archivo, binary.LittleEndian, &bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
								return
							}
							//Actualizar el super bloque
							super_bloque.S_first_blo += int32(binary.Size(bloque))
							super_bloque.S_free_blocks_count -= 1
							archivo.Seek(int64(pos_sb), 0)
							err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
								return
							}
							//Actualizar el bitmap de bloques
							pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
							archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
							var uno byte = 1
							err = binary.Write(archivo, binary.LittleEndian, &uno)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
								return
							}
							consola_mkfile += "[*SUCCESS*] Se ha creado el archivo " + nombre_archivo + " correctamente\n"
							return
						} else {
							//Creo el primer bloque_archivo
							bloque := Bloque_Archivo{}
							copy(bloque.B_content[:], contenido[:64])
							contenido = contenido[64:]
							archivo.Seek(int64(super_bloque.S_first_blo), 0)
							err = binary.Write(archivo, binary.LittleEndian, &bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
								return
							}
							//Actualizar el super bloque
							super_bloque.S_first_blo += int32(binary.Size(bloque))
							super_bloque.S_free_blocks_count -= 1
							archivo.Seek(int64(pos_sb), 0)
							err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
								return
							}
							//Actualizar el bitmap de bloques
							pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
							archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
							var uno byte = 1
							err = binary.Write(archivo, binary.LittleEndian, &uno)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
								return
							}
							for j := 0; j < len(nuevo_inodo.I_block); j++ {
								no_bloque2 := int64(nuevo_inodo.I_block[j])
								if no_bloque2 == -1 {
									break
								}
								if nuevo_inodo.I_block[j+1] == -1 {
									if len(contenido) > 64 {
										bloque := Bloque_Archivo{}
										copy(bloque.B_content[:], contenido[:64])
										contenido = contenido[64:]
										archivo.Seek(int64(super_bloque.S_first_blo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
											return
										}
										//Actualizo el inodo
										nuevo_inodo.I_block[j+1] = int32(super_bloque.S_first_blo)
										archivo.Seek(int64(posicion_nuevo_inodo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
											return
										}

										//Actualizar el super bloque
										super_bloque.S_first_blo += int32(binary.Size(bloque))
										super_bloque.S_free_blocks_count -= 1
										archivo.Seek(int64(pos_sb), 0)
										err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
											return
										}
										//Actualizar el bitmap de bloques
										pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
										archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
										var uno byte = 1
										err = binary.Write(archivo, binary.LittleEndian, &uno)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
											return
										}
									} else {
										bloque := Bloque_Archivo{}
										copy(bloque.B_content[:], contenido)
										archivo.Seek(int64(super_bloque.S_first_blo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
											return
										}
										//Actualizo el inodo
										nuevo_inodo.I_block[j+1] = int32(super_bloque.S_first_blo)
										archivo.Seek(int64(posicion_nuevo_inodo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
											return
										}

										//Actualizar el super bloque
										super_bloque.S_first_blo += int32(binary.Size(bloque))
										super_bloque.S_free_blocks_count -= 1
										archivo.Seek(int64(pos_sb), 0)
										err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
											return
										}
										//Actualizar el bitmap de bloques
										pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
										archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
										var uno byte = 1
										err = binary.Write(archivo, binary.LittleEndian, &uno)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
											return
										}
									}
								}
							}
						}
						consola_mkfile += "[*SUCCESS*] Se ha creado el archivo: " + nombre_archivo + " exitosamente\n"
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
							consola_mkfile += "[-ERROR-] Error al escribir el Bloque\n"
							return
						}
						//Actualizar el super bloque
						super_bloque.S_first_blo += int32(binary.Size(siguiente_bloque))
						super_bloque.S_free_blocks_count -= 1
						archivo.Seek(int64(pos_sb), 0)
						err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
							return
						}
						var uno byte = 1
						//Actualizar el bitmap de bloques
						pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
						archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
						err = binary.Write(archivo, binary.LittleEndian, &uno)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
							return
						}
						//Actualizar el inodo padre
						inodo_carpeta.I_block[i+1] = int32(pos_sig_bloque)
						archivo.Seek(int64(Bloque_prueba.B_content[0].B_inodo), 0)
						err = binary.Write(archivo, binary.LittleEndian, &inodo_carpeta)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
							return
						}

						//Crear el inodo de archivo
						nuevo_inodo := Inodo{}
						nuevo_inodo.I_uid = int32(Id_UserLogueado)
						nuevo_inodo.I_gid = int32(Id_GroupLogueado)
						nuevo_inodo.I_size = int32(mkfile.Size)
						copy(nuevo_inodo.I_atime[:], tiempoFormateado)
						copy(nuevo_inodo.I_ctime[:], tiempoFormateado)
						copy(nuevo_inodo.I_mtime[:], tiempoFormateado)
						nuevo_inodo.I_type = 1
						nuevo_inodo.I_perm = 664
						for j := 0; j < 16; j++ {
							nuevo_inodo.I_block[j] = -1
						}
						nuevo_inodo.I_block[0] = super_bloque.S_first_blo
						//Escribir el inodo en el archivo
						posicion_nuevo_inodo := super_bloque.S_first_ino
						archivo.Seek(int64(posicion_nuevo_inodo), 0)
						err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
							return
						}
						//Actualizar el super bloque
						super_bloque.S_first_ino += int32(binary.Size(nuevo_inodo))
						super_bloque.S_free_inodes_count -= 1
						archivo.Seek(int64(pos_sb), 0)
						err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
							return
						}
						//Actualizar el bitmap de inodos
						pos_bitmap = super_bloque.S_inodes_count - super_bloque.S_free_inodes_count
						archivo.Seek(int64(super_bloque.S_bm_inode_start+pos_bitmap), 0)

						err = binary.Write(archivo, binary.LittleEndian, &uno)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Inodos\n"
							return
						}
						//Actualizar el bloque de carpetas
						siguiente_bloque.B_content[2].B_inodo = int32(posicion_nuevo_inodo)
						copy(siguiente_bloque.B_content[2].B_name[:], nombre_archivo)
						archivo.Seek(int64(pos_sig_bloque), 0)
						err = binary.Write(archivo, binary.LittleEndian, &siguiente_bloque)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bloque\n"
							return
						}
						//Crea el nuevo bloque Archivo
						contenido := mkfile.RetornaContenidoSize()
						if mkfile.Size < 64 {
							//Creo el primer bloque_archivo
							bloque := Bloque_Archivo{}
							copy(bloque.B_content[:], contenido)
							archivo.Seek(int64(super_bloque.S_first_blo), 0)
							err = binary.Write(archivo, binary.LittleEndian, &bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
								return
							}
							//Actualizar el super bloque
							super_bloque.S_first_blo += int32(binary.Size(bloque))
							super_bloque.S_free_blocks_count -= 1
							archivo.Seek(int64(pos_sb), 0)
							err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
								return
							}
							//Actualizar el bitmap de bloques
							pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
							archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
							var uno byte = 1
							err = binary.Write(archivo, binary.LittleEndian, &uno)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
								return
							}
							consola_mkfile += "[*SUCCESS*] Se ha creado el archivo " + nombre_archivo + " correctamente\n"
							return
						} else {
							//Creo el primer bloque_archivo
							bloque := Bloque_Archivo{}
							copy(bloque.B_content[:], contenido[:64])
							contenido = contenido[64:]
							archivo.Seek(int64(super_bloque.S_first_blo), 0)
							err = binary.Write(archivo, binary.LittleEndian, &bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
								return
							}
							//Actualizar el super bloque
							super_bloque.S_first_blo += int32(binary.Size(bloque))
							super_bloque.S_free_blocks_count -= 1
							archivo.Seek(int64(pos_sb), 0)
							err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
								return
							}
							//Actualizar el bitmap de bloques
							pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
							archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
							var uno byte = 1
							err = binary.Write(archivo, binary.LittleEndian, &uno)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
								return
							}
							for j := 0; j < len(nuevo_inodo.I_block); j++ {
								no_bloque2 := int64(nuevo_inodo.I_block[j])
								if no_bloque2 == -1 {
									break
								}
								if nuevo_inodo.I_block[j+1] == -1 {
									if len(contenido) > 64 {
										bloque := Bloque_Archivo{}
										copy(bloque.B_content[:], contenido[:64])
										contenido = contenido[64:]
										archivo.Seek(int64(super_bloque.S_first_blo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
											return
										}
										//Actualizo el inodo
										nuevo_inodo.I_block[j+1] = int32(super_bloque.S_first_blo)
										archivo.Seek(int64(posicion_nuevo_inodo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
											return
										}

										//Actualizar el super bloque
										super_bloque.S_first_blo += int32(binary.Size(bloque))
										super_bloque.S_free_blocks_count -= 1
										archivo.Seek(int64(pos_sb), 0)
										err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
											return
										}
										//Actualizar el bitmap de bloques
										pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
										archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
										var uno byte = 1
										err = binary.Write(archivo, binary.LittleEndian, &uno)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
											return
										}
									} else {
										bloque := Bloque_Archivo{}
										copy(bloque.B_content[:], contenido)
										archivo.Seek(int64(super_bloque.S_first_blo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
											return
										}
										//Actualizo el inodo
										nuevo_inodo.I_block[j+1] = int32(super_bloque.S_first_blo)
										archivo.Seek(int64(posicion_nuevo_inodo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
											return
										}

										//Actualizar el super bloque
										super_bloque.S_first_blo += int32(binary.Size(bloque))
										super_bloque.S_free_blocks_count -= 1
										archivo.Seek(int64(pos_sb), 0)
										err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
											return
										}
										//Actualizar el bitmap de bloques
										pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
										archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
										var uno byte = 1
										err = binary.Write(archivo, binary.LittleEndian, &uno)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
											return
										}
									}
								}
							}
						}
						consola_mkfile += "[*SUCCESS*] Se ha creado el archivo: " + nombre_archivo + " exitosamente\n"
						return
					}
				}
			}
		}
	}

}

func (mkfile *Mkfile) CrearArchivoConTamaño(pos_sb int, path string) {
	archivo, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		consola_mkfile += "[-ERROR-] Error al abrir el archivo\n"
		return
	}
	defer archivo.Close()
	name_carpetas := strings.Split(mkfile.Path, "/")
	if name_carpetas[0] == "" {
		name_carpetas[0] = "/"
	} else {
		consola_mkfile += "[-ERROR-] La ruta no es absoluta\n"
	}
	nombre_archivo := name_carpetas[len(name_carpetas)-1]
	name_carpetas = name_carpetas[:len(name_carpetas)-1]
	//Leer el SuperBloque
	archivo.Seek(int64(pos_sb), 0)
	super_bloque := SuperBloque{}
	err = binary.Read(archivo, binary.LittleEndian, &super_bloque)
	if err != nil {
		consola_mkfile += "[-ERROR-] Error al leer el SuperBloque\n"
		return
	}

	if !mkfile.ExisteCarpetaPadre(name_carpetas, int(super_bloque.S_inode_start), path) {
		consola_mkfile += "[-ERROR-] No existe la carpeta donde quieres agregar una nueva carpeta\n"
		return
	}

	//Leer el Inodo de la carpeta
	posicion_carpeta := mkfile.PosCarpetaPadre(name_carpetas, int(super_bloque.S_inode_start), path)
	fmt.Println("Posicion padre: ", posicion_carpeta)
	archivo.Seek(int64(posicion_carpeta), 0)
	inodo_carpeta := Inodo{}
	err = binary.Read(archivo, binary.LittleEndian, &inodo_carpeta)
	if err != nil {
		consola_mkfile += "[-ERROR-] Error al leer el Inodo\n"
		return
	}
	//nombrare en que carpeta estoy
	archivo.Seek(int64(inodo_carpeta.I_block[0]), 0)
	Bloque_prueba := Bloque_Carpeta{}
	binary.Read(archivo, binary.LittleEndian, &Bloque_prueba)
	fmt.Println("Nombre de la carpeta actual: ", string(Bloque_prueba.B_content[0].B_name[:]))
	tiempo := time.Now()
	tiempoFormateado := tiempo.Format("2006-01-02 15:04:05")

	for i := 0; i < len(inodo_carpeta.I_block); i++ {
		if i != 15 {
			no_bloque := int(inodo_carpeta.I_block[i])
			if no_bloque != -1 {
				if inodo_carpeta.I_block[i+1] == -1 {
					bloque_c := Bloque_Carpeta{}
					archivo.Seek(int64(no_bloque), 0)
					err = binary.Read(archivo, binary.LittleEndian, &bloque_c)
					if err != nil {
						consola_mkfile += "[-ERROR-] Error al leer el Bloque\n"
						return
					}
					var hay_espacio bool = false
					var pos_b int = 0
					for j := 2; j < 4; j++ {
						if bloque_c.B_content[j].B_inodo == -1 {
							hay_espacio = true
							pos_b = j
							break
						}
					}
					if hay_espacio {
						//Crear el inodo de archivo
						nuevo_inodo := Inodo{}
						nuevo_inodo.I_uid = int32(Id_UserLogueado)
						nuevo_inodo.I_gid = int32(Id_GroupLogueado)
						nuevo_inodo.I_size = int32(mkfile.Size)
						copy(nuevo_inodo.I_atime[:], tiempoFormateado)
						copy(nuevo_inodo.I_ctime[:], tiempoFormateado)
						copy(nuevo_inodo.I_mtime[:], tiempoFormateado)
						nuevo_inodo.I_type = 1
						nuevo_inodo.I_perm = 664
						for j := 0; j < 16; j++ {
							nuevo_inodo.I_block[j] = -1
						}
						nuevo_inodo.I_block[0] = super_bloque.S_first_blo
						//Escribir el inodo en el archivo
						posicion_nuevo_inodo := super_bloque.S_first_ino
						archivo.Seek(int64(posicion_nuevo_inodo), 0)
						err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
							return
						}
						//Actualizar el super bloque
						super_bloque.S_first_ino += int32(binary.Size(nuevo_inodo))
						super_bloque.S_free_inodes_count -= 1
						archivo.Seek(int64(pos_sb), 0)
						err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
							return
						}
						//Actualizar el bitmap de inodos
						pos_bitmap := super_bloque.S_inodes_count - super_bloque.S_free_inodes_count
						archivo.Seek(int64(super_bloque.S_bm_inode_start+pos_bitmap), 0)
						var uno byte = 1
						err = binary.Write(archivo, binary.LittleEndian, &uno)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Inodos\n"
							return
						}
						//Actualizar el bloque de carpetas
						bloque_c.B_content[pos_b].B_inodo = int32(posicion_nuevo_inodo)
						copy(bloque_c.B_content[pos_b].B_name[:], nombre_archivo)
						archivo.Seek(int64(no_bloque), 0)
						err = binary.Write(archivo, binary.LittleEndian, &bloque_c)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bloque\n"
							return
						}
						//Crea el nuevo bloque Archivo
						contenido := mkfile.RetornaContenidoSize()
						if mkfile.Size < 64 {
							//Creo el primer bloque_archivo
							bloque := Bloque_Archivo{}
							copy(bloque.B_content[:], contenido)
							archivo.Seek(int64(super_bloque.S_first_blo), 0)
							err = binary.Write(archivo, binary.LittleEndian, &bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
								return
							}
							//Actualizar el super bloque
							super_bloque.S_first_blo += int32(binary.Size(bloque))
							super_bloque.S_free_blocks_count -= 1
							archivo.Seek(int64(pos_sb), 0)
							err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
								return
							}
							//Actualizar el bitmap de bloques
							pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
							archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
							var uno byte = 1
							err = binary.Write(archivo, binary.LittleEndian, &uno)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
								return
							}
							consola_mkfile += "[*SUCCESS*] Se ha creado el archivo " + nombre_archivo + " correctamente\n"
							return
						} else {
							//Creo el primer bloque_archivo
							bloque := Bloque_Archivo{}
							copy(bloque.B_content[:], contenido[:64])
							contenido = contenido[64:]
							archivo.Seek(int64(super_bloque.S_first_blo), 0)
							err = binary.Write(archivo, binary.LittleEndian, &bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
								return
							}
							//Actualizar el super bloque
							super_bloque.S_first_blo += int32(binary.Size(bloque))
							super_bloque.S_free_blocks_count -= 1
							archivo.Seek(int64(pos_sb), 0)
							err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
								return
							}
							//Actualizar el bitmap de bloques
							pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
							archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
							var uno byte = 1
							err = binary.Write(archivo, binary.LittleEndian, &uno)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
								return
							}
							for j := 0; j < len(nuevo_inodo.I_block); j++ {
								no_bloque2 := int64(nuevo_inodo.I_block[j])
								if no_bloque2 == -1 {
									break
								}
								if nuevo_inodo.I_block[j+1] == -1 {
									if len(contenido) > 64 {
										bloque := Bloque_Archivo{}
										copy(bloque.B_content[:], contenido[:64])
										contenido = contenido[64:]
										archivo.Seek(int64(super_bloque.S_first_blo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
											return
										}
										//Actualizo el inodo
										nuevo_inodo.I_block[j+1] = int32(super_bloque.S_first_blo)
										archivo.Seek(int64(posicion_nuevo_inodo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
											return
										}

										//Actualizar el super bloque
										super_bloque.S_first_blo += int32(binary.Size(bloque))
										super_bloque.S_free_blocks_count -= 1
										archivo.Seek(int64(pos_sb), 0)
										err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
											return
										}
										//Actualizar el bitmap de bloques
										pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
										archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
										var uno byte = 1
										err = binary.Write(archivo, binary.LittleEndian, &uno)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
											return
										}
									} else {
										bloque := Bloque_Archivo{}
										copy(bloque.B_content[:], contenido)
										archivo.Seek(int64(super_bloque.S_first_blo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
											return
										}
										//Actualizo el inodo
										nuevo_inodo.I_block[j+1] = int32(super_bloque.S_first_blo)
										archivo.Seek(int64(posicion_nuevo_inodo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
											return
										}

										//Actualizar el super bloque
										super_bloque.S_first_blo += int32(binary.Size(bloque))
										super_bloque.S_free_blocks_count -= 1
										archivo.Seek(int64(pos_sb), 0)
										err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
											return
										}
										//Actualizar el bitmap de bloques
										pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
										archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
										var uno byte = 1
										err = binary.Write(archivo, binary.LittleEndian, &uno)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
											return
										}
									}
								}
							}
						}
						consola_mkfile += "[*SUCCESS*] Se ha creado el archivo: " + nombre_archivo + " exitosamente\n"
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
							consola_mkfile += "[-ERROR-] Error al escribir el Bloque\n"
							return
						}
						//Actualizar el super bloque
						super_bloque.S_first_blo += int32(binary.Size(siguiente_bloque))
						super_bloque.S_free_blocks_count -= 1
						archivo.Seek(int64(pos_sb), 0)
						err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
							return
						}
						var uno byte = 1
						//Actualizar el bitmap de bloques
						pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
						archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
						err = binary.Write(archivo, binary.LittleEndian, &uno)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
							return
						}
						//Actualizar el inodo padre
						inodo_carpeta.I_block[i+1] = int32(pos_sig_bloque)
						archivo.Seek(int64(Bloque_prueba.B_content[0].B_inodo), 0)
						err = binary.Write(archivo, binary.LittleEndian, &inodo_carpeta)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
							return
						}

						//Crear el inodo de archivo
						nuevo_inodo := Inodo{}
						nuevo_inodo.I_uid = int32(Id_UserLogueado)
						nuevo_inodo.I_gid = int32(Id_GroupLogueado)
						nuevo_inodo.I_size = int32(mkfile.Size)
						copy(nuevo_inodo.I_atime[:], tiempoFormateado)
						copy(nuevo_inodo.I_ctime[:], tiempoFormateado)
						copy(nuevo_inodo.I_mtime[:], tiempoFormateado)
						nuevo_inodo.I_type = 1
						nuevo_inodo.I_perm = 664
						for j := 0; j < 16; j++ {
							nuevo_inodo.I_block[j] = -1
						}
						nuevo_inodo.I_block[0] = super_bloque.S_first_blo
						//Escribir el inodo en el archivo
						posicion_nuevo_inodo := super_bloque.S_first_ino
						archivo.Seek(int64(posicion_nuevo_inodo), 0)
						err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
							return
						}
						//Actualizar el super bloque
						super_bloque.S_first_ino += int32(binary.Size(nuevo_inodo))
						super_bloque.S_free_inodes_count -= 1
						archivo.Seek(int64(pos_sb), 0)
						err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
							return
						}
						//Actualizar el bitmap de inodos
						pos_bitmap = super_bloque.S_inodes_count - super_bloque.S_free_inodes_count
						archivo.Seek(int64(super_bloque.S_bm_inode_start+pos_bitmap), 0)

						err = binary.Write(archivo, binary.LittleEndian, &uno)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Inodos\n"
							return
						}
						//Actualizar el bloque de carpetas
						siguiente_bloque.B_content[2].B_inodo = int32(posicion_nuevo_inodo)
						copy(siguiente_bloque.B_content[2].B_name[:], nombre_archivo)
						archivo.Seek(int64(pos_sig_bloque), 0)
						err = binary.Write(archivo, binary.LittleEndian, &siguiente_bloque)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bloque\n"
							return
						}
						//Crea el nuevo bloque Archivo
						contenido := mkfile.RetornaContenidoSize()
						if mkfile.Size < 64 {
							//Creo el primer bloque_archivo
							bloque := Bloque_Archivo{}
							copy(bloque.B_content[:], contenido)
							archivo.Seek(int64(super_bloque.S_first_blo), 0)
							err = binary.Write(archivo, binary.LittleEndian, &bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
								return
							}
							//Actualizar el super bloque
							super_bloque.S_first_blo += int32(binary.Size(bloque))
							super_bloque.S_free_blocks_count -= 1
							archivo.Seek(int64(pos_sb), 0)
							err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
								return
							}
							//Actualizar el bitmap de bloques
							pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
							archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
							var uno byte = 1
							err = binary.Write(archivo, binary.LittleEndian, &uno)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
								return
							}
							consola_mkfile += "[*SUCCESS*] Se ha creado el archivo " + nombre_archivo + " correctamente\n"
							return
						} else {
							//Creo el primer bloque_archivo
							bloque := Bloque_Archivo{}
							copy(bloque.B_content[:], contenido[:64])
							contenido = contenido[64:]
							archivo.Seek(int64(super_bloque.S_first_blo), 0)
							err = binary.Write(archivo, binary.LittleEndian, &bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
								return
							}
							//Actualizar el super bloque
							super_bloque.S_first_blo += int32(binary.Size(bloque))
							super_bloque.S_free_blocks_count -= 1
							archivo.Seek(int64(pos_sb), 0)
							err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
								return
							}
							//Actualizar el bitmap de bloques
							pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
							archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
							var uno byte = 1
							err = binary.Write(archivo, binary.LittleEndian, &uno)
							if err != nil {
								consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
								return
							}
							for j := 0; j < len(nuevo_inodo.I_block); j++ {
								no_bloque2 := int64(nuevo_inodo.I_block[j])
								if no_bloque2 == -1 {
									break
								}
								if nuevo_inodo.I_block[j+1] == -1 {
									if len(contenido) > 64 {
										bloque := Bloque_Archivo{}
										copy(bloque.B_content[:], contenido[:64])
										contenido = contenido[64:]
										archivo.Seek(int64(super_bloque.S_first_blo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
											return
										}
										//Actualizo el inodo
										nuevo_inodo.I_block[j+1] = int32(super_bloque.S_first_blo)
										archivo.Seek(int64(posicion_nuevo_inodo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
											return
										}

										//Actualizar el super bloque
										super_bloque.S_first_blo += int32(binary.Size(bloque))
										super_bloque.S_free_blocks_count -= 1
										archivo.Seek(int64(pos_sb), 0)
										err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
											return
										}
										//Actualizar el bitmap de bloques
										pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
										archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
										var uno byte = 1
										err = binary.Write(archivo, binary.LittleEndian, &uno)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
											return
										}
									} else {
										bloque := Bloque_Archivo{}
										copy(bloque.B_content[:], contenido)
										archivo.Seek(int64(super_bloque.S_first_blo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bloque Archivo\n"
											return
										}
										//Actualizo el inodo
										nuevo_inodo.I_block[j+1] = int32(super_bloque.S_first_blo)
										archivo.Seek(int64(posicion_nuevo_inodo), 0)
										err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
											return
										}

										//Actualizar el super bloque
										super_bloque.S_first_blo += int32(binary.Size(bloque))
										super_bloque.S_free_blocks_count -= 1
										archivo.Seek(int64(pos_sb), 0)
										err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
											return
										}
										//Actualizar el bitmap de bloques
										pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
										archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
										var uno byte = 1
										err = binary.Write(archivo, binary.LittleEndian, &uno)
										if err != nil {
											consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
											return
										}
									}
								}
							}
						}
						consola_mkfile += "[*SUCCESS*] Se ha creado el archivo: " + nombre_archivo + " exitosamente\n"
						return
					}
				}
			}
		}
	}

}

func (mkfile *Mkfile) CrearArchivoRecursivo(pos_sb int, path string) {
	archivo, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		consola_mkfile += "[-ERROR-] Error al abrir el archivo\n"
		return
	}
	defer archivo.Close()
	name_carpetas := strings.Split(mkfile.Path, "/")
	if name_carpetas[0] == "" {
		name_carpetas[0] = "/"
	} else {
		consola_mkfile += "[-ERROR-] La ruta no es absoluta\n"
	}
	nombre_archivo := name_carpetas[len(name_carpetas)-1]
	name_carpetas = name_carpetas[:len(name_carpetas)-1]
	mkfile.CreacionRecursiva(name_carpetas, pos_sb, path)

	//Leer el SuperBloque
	archivo.Seek(int64(pos_sb), 0)
	super_bloque := SuperBloque{}
	err = binary.Read(archivo, binary.LittleEndian, &super_bloque)
	if err != nil {
		consola_mkfile += "[-ERROR-] Error al leer el SuperBloque\n"
		return
	}

	//Leer el Inodo de la carpeta
	posicion_carpeta := mkfile.PosCarpetaPadre(name_carpetas, int(super_bloque.S_inode_start), path)
	fmt.Println("Posicion padre: ", posicion_carpeta)
	archivo.Seek(int64(posicion_carpeta), 0)
	inodo_carpeta := Inodo{}
	err = binary.Read(archivo, binary.LittleEndian, &inodo_carpeta)
	if err != nil {
		consola_mkfile += "[-ERROR-] Error al leer el Inodo\n"
		return
	}
	//nombrare en que carpeta estoy
	archivo.Seek(int64(inodo_carpeta.I_block[0]), 0)
	Bloque_prueba := Bloque_Carpeta{}
	binary.Read(archivo, binary.LittleEndian, &Bloque_prueba)
	fmt.Println("Nombre de la carpeta actual: ", string(Bloque_prueba.B_content[0].B_name[:]))
	tiempo := time.Now()
	tiempoFormateado := tiempo.Format("2006-01-02 15:04:05")

	for i := 0; i < len(inodo_carpeta.I_block); i++ {
		if i != 15 {
			no_bloque := int(inodo_carpeta.I_block[i])
			if no_bloque != -1 {
				if inodo_carpeta.I_block[i+1] == -1 {
					bloque_c := Bloque_Carpeta{}
					archivo.Seek(int64(no_bloque), 0)
					err = binary.Read(archivo, binary.LittleEndian, &bloque_c)
					if err != nil {
						consola_mkfile += "[-ERROR-] Error al leer el Bloque\n"
						return
					}
					var hay_espacio bool = false
					var pos_b int = 0
					for j := 2; j < 4; j++ {
						if bloque_c.B_content[j].B_inodo == -1 {
							hay_espacio = true
							pos_b = j
							break
						}
					}
					if hay_espacio {
						//Crear el inodo de archivo
						nuevo_inodo := Inodo{}
						nuevo_inodo.I_uid = int32(Id_UserLogueado)
						nuevo_inodo.I_gid = int32(Id_GroupLogueado)
						nuevo_inodo.I_size = 0
						copy(nuevo_inodo.I_atime[:], tiempoFormateado)
						copy(nuevo_inodo.I_ctime[:], tiempoFormateado)
						copy(nuevo_inodo.I_mtime[:], tiempoFormateado)
						nuevo_inodo.I_type = 1
						nuevo_inodo.I_perm = 664
						for j := 0; j < 16; j++ {
							nuevo_inodo.I_block[j] = -1
						}
						nuevo_inodo.I_block[0] = super_bloque.S_first_blo
						//Escribir el inodo en el archivo
						posicion_nuevo_inodo := super_bloque.S_first_ino
						archivo.Seek(int64(posicion_nuevo_inodo), 0)
						err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
							return
						}
						//Actualizar el super bloque
						super_bloque.S_first_ino += int32(binary.Size(nuevo_inodo))
						super_bloque.S_free_inodes_count -= 1
						archivo.Seek(int64(pos_sb), 0)
						err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
							return
						}
						//Actualizar el bitmap de inodos
						pos_bitmap := super_bloque.S_inodes_count - super_bloque.S_free_inodes_count
						archivo.Seek(int64(super_bloque.S_bm_inode_start+pos_bitmap), 0)
						var uno byte = 1
						err = binary.Write(archivo, binary.LittleEndian, &uno)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Inodos\n"
							return
						}
						//Actualizar el bloque de carpetas
						bloque_c.B_content[pos_b].B_inodo = int32(posicion_nuevo_inodo)
						copy(bloque_c.B_content[pos_b].B_name[:], nombre_archivo)
						archivo.Seek(int64(no_bloque), 0)
						err = binary.Write(archivo, binary.LittleEndian, &bloque_c)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bloque\n"
							return
						}
						consola_mkfile += "[*SUCCESS*] Se ha creado el archivo " + nombre_archivo + " correctamente\n"
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
							consola_mkfile += "[-ERROR-] Error al escribir el Bloque\n"
							return
						}
						//Actualizar el super bloque
						super_bloque.S_first_blo += int32(binary.Size(siguiente_bloque))
						super_bloque.S_free_blocks_count -= 1
						archivo.Seek(int64(pos_sb), 0)
						err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
							return
						}
						var uno byte = 1
						//Actualizar el bitmap de bloques
						pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
						archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
						err = binary.Write(archivo, binary.LittleEndian, &uno)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
							return
						}
						//Actualizar el inodo padre
						inodo_carpeta.I_block[i+1] = int32(pos_sig_bloque)
						archivo.Seek(int64(Bloque_prueba.B_content[0].B_inodo), 0)
						err = binary.Write(archivo, binary.LittleEndian, &inodo_carpeta)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
							return
						}

						//Crear el inodo de archivo
						nuevo_inodo := Inodo{}
						nuevo_inodo.I_uid = int32(Id_UserLogueado)
						nuevo_inodo.I_gid = int32(Id_GroupLogueado)
						nuevo_inodo.I_size = 0
						copy(nuevo_inodo.I_atime[:], tiempoFormateado)
						copy(nuevo_inodo.I_ctime[:], tiempoFormateado)
						copy(nuevo_inodo.I_mtime[:], tiempoFormateado)
						nuevo_inodo.I_type = 1
						nuevo_inodo.I_perm = 664
						for j := 0; j < 16; j++ {
							nuevo_inodo.I_block[j] = -1
						}
						nuevo_inodo.I_block[0] = super_bloque.S_first_blo
						//Escribir el inodo en el archivo
						posicion_nuevo_inodo := super_bloque.S_first_ino
						archivo.Seek(int64(posicion_nuevo_inodo), 0)
						err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
							return
						}
						//Actualizar el super bloque
						super_bloque.S_first_ino += int32(binary.Size(nuevo_inodo))
						super_bloque.S_free_inodes_count -= 1
						archivo.Seek(int64(pos_sb), 0)
						err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
							return
						}
						//Actualizar el bitmap de inodos
						pos_bitmap = super_bloque.S_inodes_count - super_bloque.S_free_inodes_count
						archivo.Seek(int64(super_bloque.S_bm_inode_start+pos_bitmap), 0)

						err = binary.Write(archivo, binary.LittleEndian, &uno)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Inodos\n"
							return
						}
						//Actualizar el bloque de carpetas
						siguiente_bloque.B_content[2].B_inodo = int32(posicion_nuevo_inodo)
						copy(siguiente_bloque.B_content[2].B_name[:], nombre_archivo)
						archivo.Seek(int64(pos_sig_bloque), 0)
						err = binary.Write(archivo, binary.LittleEndian, &siguiente_bloque)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bloque\n"
							return
						}
						consola_mkfile += "[*SUCCESS*] Se ha creado el archivo " + nombre_archivo + " correctamente\n"
						return
					}
				}
			}
		}
	}
	consola_mkfile += "[-ERROR-] No se pudo crear el archivo\n"
}

func (mkfile *Mkfile) CrearArchivoNoRecursivo(pos_sb int, path string) {
	archivo, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		consola_mkfile += "[-ERROR-] Error al abrir el archivo\n"
		return
	}
	defer archivo.Close()
	name_carpetas := strings.Split(mkfile.Path, "/")
	if name_carpetas[0] == "" {
		name_carpetas[0] = "/"
	} else {
		consola_mkfile += "[-ERROR-] La ruta no es absoluta\n"
	}
	nombre_archivo := name_carpetas[len(name_carpetas)-1]
	name_carpetas = name_carpetas[:len(name_carpetas)-1]

	//Leer el SuperBloque
	archivo.Seek(int64(pos_sb), 0)
	super_bloque := SuperBloque{}
	err = binary.Read(archivo, binary.LittleEndian, &super_bloque)
	if err != nil {
		consola_mkfile += "[-ERROR-] Error al leer el SuperBloque\n"
		return
	}

	if !mkfile.ExisteCarpetaPadre(name_carpetas, int(super_bloque.S_inode_start), path) {
		consola_mkfile += "[-ERROR-] No existe la carpeta donde quieres agregar una nueva carpeta\n"
		return
	}

	//Leer el Inodo de la carpeta
	posicion_carpeta := mkfile.PosCarpetaPadre(name_carpetas, int(super_bloque.S_inode_start), path)
	fmt.Println("Posicion padre: ", posicion_carpeta)
	archivo.Seek(int64(posicion_carpeta), 0)
	inodo_carpeta := Inodo{}
	err = binary.Read(archivo, binary.LittleEndian, &inodo_carpeta)
	if err != nil {
		consola_mkfile += "[-ERROR-] Error al leer el Inodo\n"
		return
	}
	//nombrare en que carpeta estoy
	archivo.Seek(int64(inodo_carpeta.I_block[0]), 0)
	Bloque_prueba := Bloque_Carpeta{}
	binary.Read(archivo, binary.LittleEndian, &Bloque_prueba)
	fmt.Println("Nombre de la carpeta actual: ", string(Bloque_prueba.B_content[0].B_name[:]))
	tiempo := time.Now()
	tiempoFormateado := tiempo.Format("2006-01-02 15:04:05")

	for i := 0; i < len(inodo_carpeta.I_block); i++ {
		if i != 15 {
			no_bloque := int(inodo_carpeta.I_block[i])
			if no_bloque != -1 {
				if inodo_carpeta.I_block[i+1] == -1 {
					bloque_c := Bloque_Carpeta{}
					archivo.Seek(int64(no_bloque), 0)
					err = binary.Read(archivo, binary.LittleEndian, &bloque_c)
					if err != nil {
						consola_mkfile += "[-ERROR-] Error al leer el Bloque\n"
						return
					}
					var hay_espacio bool = false
					var pos_b int = 0
					for j := 2; j < 4; j++ {
						if bloque_c.B_content[j].B_inodo == -1 {
							hay_espacio = true
							pos_b = j
							break
						}
					}
					if hay_espacio {
						//Crear el inodo de archivo
						nuevo_inodo := Inodo{}
						nuevo_inodo.I_uid = int32(Id_UserLogueado)
						nuevo_inodo.I_gid = int32(Id_GroupLogueado)
						nuevo_inodo.I_size = 0
						copy(nuevo_inodo.I_atime[:], tiempoFormateado)
						copy(nuevo_inodo.I_ctime[:], tiempoFormateado)
						copy(nuevo_inodo.I_mtime[:], tiempoFormateado)
						nuevo_inodo.I_type = 1
						nuevo_inodo.I_perm = 664
						for j := 0; j < 16; j++ {
							nuevo_inodo.I_block[j] = -1
						}
						nuevo_inodo.I_block[0] = super_bloque.S_first_blo
						//Escribir el inodo en el archivo
						posicion_nuevo_inodo := super_bloque.S_first_ino
						archivo.Seek(int64(posicion_nuevo_inodo), 0)
						err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
							return
						}
						//Actualizar el super bloque
						super_bloque.S_first_ino += int32(binary.Size(nuevo_inodo))
						super_bloque.S_free_inodes_count -= 1
						archivo.Seek(int64(pos_sb), 0)
						err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
							return
						}
						//Actualizar el bitmap de inodos
						pos_bitmap := super_bloque.S_inodes_count - super_bloque.S_free_inodes_count
						archivo.Seek(int64(super_bloque.S_bm_inode_start+pos_bitmap), 0)
						var uno byte = 1
						err = binary.Write(archivo, binary.LittleEndian, &uno)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Inodos\n"
							return
						}
						//Actualizar el bloque de carpetas
						bloque_c.B_content[pos_b].B_inodo = int32(posicion_nuevo_inodo)
						copy(bloque_c.B_content[pos_b].B_name[:], nombre_archivo)
						archivo.Seek(int64(no_bloque), 0)
						err = binary.Write(archivo, binary.LittleEndian, &bloque_c)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bloque\n"
							return
						}
						consola_mkfile += "[*SUCCESS*] Se ha creado el archivo " + nombre_archivo + " correctamente\n"
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
							consola_mkfile += "[-ERROR-] Error al escribir el Bloque\n"
							return
						}
						//Actualizar el super bloque
						super_bloque.S_first_blo += int32(binary.Size(siguiente_bloque))
						super_bloque.S_free_blocks_count -= 1
						archivo.Seek(int64(pos_sb), 0)
						err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
							return
						}
						var uno byte = 1
						//Actualizar el bitmap de bloques
						pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
						archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
						err = binary.Write(archivo, binary.LittleEndian, &uno)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
							return
						}
						//Actualizar el inodo padre
						inodo_carpeta.I_block[i+1] = int32(pos_sig_bloque)
						archivo.Seek(int64(Bloque_prueba.B_content[0].B_inodo), 0)
						err = binary.Write(archivo, binary.LittleEndian, &inodo_carpeta)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
							return
						}

						//Crear el inodo de archivo
						nuevo_inodo := Inodo{}
						nuevo_inodo.I_uid = int32(Id_UserLogueado)
						nuevo_inodo.I_gid = int32(Id_GroupLogueado)
						nuevo_inodo.I_size = 0
						copy(nuevo_inodo.I_atime[:], tiempoFormateado)
						copy(nuevo_inodo.I_ctime[:], tiempoFormateado)
						copy(nuevo_inodo.I_mtime[:], tiempoFormateado)
						nuevo_inodo.I_type = 1
						nuevo_inodo.I_perm = 664
						for j := 0; j < 16; j++ {
							nuevo_inodo.I_block[j] = -1
						}
						nuevo_inodo.I_block[0] = super_bloque.S_first_blo
						//Escribir el inodo en el archivo
						posicion_nuevo_inodo := super_bloque.S_first_ino
						archivo.Seek(int64(posicion_nuevo_inodo), 0)
						err = binary.Write(archivo, binary.LittleEndian, &nuevo_inodo)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
							return
						}
						//Actualizar el super bloque
						super_bloque.S_first_ino += int32(binary.Size(nuevo_inodo))
						super_bloque.S_free_inodes_count -= 1
						archivo.Seek(int64(pos_sb), 0)
						err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
							return
						}
						//Actualizar el bitmap de inodos
						pos_bitmap = super_bloque.S_inodes_count - super_bloque.S_free_inodes_count
						archivo.Seek(int64(super_bloque.S_bm_inode_start+pos_bitmap), 0)

						err = binary.Write(archivo, binary.LittleEndian, &uno)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Inodos\n"
							return
						}
						//Actualizar el bloque de carpetas
						siguiente_bloque.B_content[2].B_inodo = int32(posicion_nuevo_inodo)
						copy(siguiente_bloque.B_content[2].B_name[:], nombre_archivo)
						archivo.Seek(int64(pos_sig_bloque), 0)
						err = binary.Write(archivo, binary.LittleEndian, &siguiente_bloque)
						if err != nil {
							consola_mkfile += "[-ERROR-] Error al escribir el Bloque\n"
							return
						}
						consola_mkfile += "[*SUCCESS*] Se ha creado el archivo " + nombre_archivo + " correctamente\n"
						return
					}
				}
			}
		}
	}
	consola_mkfile += "[-ERROR-] No se pudo crear el archivo\n"
}

func (mkfile *Mkfile) CreacionRecursiva(nombres_carpetas []string, pos_sb int, path string) {
	fmt.Println("CARPETAS")
	fmt.Println(nombres_carpetas)
	archivo, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		consola_mkfile += "[-ERROR-] Error al abrir el archivo\n"
		return
	}
	defer archivo.Close()

	//LEE EL SUPERBLOQUE

	archivo.Seek(int64(pos_sb), 0)
	super_bloque := SuperBloque{}
	fmt.Println("POS SB: ", pos_sb)
	err = binary.Read(archivo, binary.LittleEndian, &super_bloque)
	if err != nil {
		consola_mkfile += "[-ERROR-] Error al leer el SuperBloque\n"
		return
	}
	fmt.Println("SUPERBLOQUE")
	fmt.Println(super_bloque)

	for i := 0; i < len(nombres_carpetas)-1; i++ {
		fmt.Println(i)
		if mkfile.ExisteCarpetaPadre(nombres_carpetas[:i+1], int(super_bloque.S_inode_start), path) && !mkfile.ExisteCarpetaPadre(nombres_carpetas[:i+2], int(super_bloque.S_inode_start), path) {
			posicion_padre := mkfile.PosCarpetaPadre(nombres_carpetas[:i+1], int(super_bloque.S_inode_start), path)
			fmt.Println("POSICION PADRE: ", posicion_padre)
			nueva_carpeta := nombres_carpetas[i+1]
			nombre_padre := nombres_carpetas[i]
			archivo.Seek(int64(posicion_padre), 0)
			//LEE EL INODO PADRE
			inodo_padre := Inodo{}
			err = binary.Read(archivo, binary.LittleEndian, &inodo_padre)
			if err != nil {
				consola_mkfile += "[-ERROR-] Error al leer el Inodo\n"
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
								consola_mkfile += "[-ERROR-] Error al leer el Bloque\n"
								return
							}
							var hay_espacio bool = false
							var pos_b int = 0
							for j := 2; j < 4; j++ {
								name_comp := string(bloque.B_content[j].B_name[:])
								name_comp = strings.Replace(name_comp, "\u0000", "", -1)
								if name_comp == nueva_carpeta {
									consola_mkfile += "[-ERROR-] Ya existe una carpeta con ese nombre\n"
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
									consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
									return
								}
								//Actualizar el super bloque
								super_bloque.S_first_ino += int32(binary.Size(nuevo_inodo))
								super_bloque.S_free_inodes_count -= 1
								archivo.Seek(int64(pos_sb), 0)
								err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
								if err != nil {
									consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
									return
								}
								//Actualizar el bitmap de inodos
								pos_bitmap := super_bloque.S_inodes_count - super_bloque.S_free_inodes_count
								archivo.Seek(int64(super_bloque.S_bm_inode_start+pos_bitmap), 0)
								var uno byte = 1
								err = binary.Write(archivo, binary.LittleEndian, &uno)
								if err != nil {
									consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Inodos\n"
									return
								}
								//Actualizar el bloque de carpetas
								bloque.B_content[pos_b].B_inodo = int32(posicion_nuevo_inodo)
								copy(bloque.B_content[pos_b].B_name[:], nueva_carpeta)
								archivo.Seek(int64(no_bloque), 0)
								err = binary.Write(archivo, binary.LittleEndian, &bloque)
								if err != nil {
									consola_mkfile += "[-ERROR-] Error al escribir el Bloque\n"
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
									consola_mkfile += "[-ERROR-] Error al escribir el Bloque\n"
									return
								}
								//Actualizar el super bloque
								super_bloque.S_first_blo += int32(binary.Size(nuevo_bloque))
								super_bloque.S_free_blocks_count -= 1
								archivo.Seek(int64(pos_sb), 0)
								err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
								if err != nil {
									consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
									return
								}
								//Actualizar el bitmap de bloques
								pos_bitmap = super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
								archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
								err = binary.Write(archivo, binary.LittleEndian, &uno)
								if err != nil {
									consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
									return
								}

								mkfile.CreacionRecursiva(nombres_carpetas, pos_sb, path)
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
									consola_mkfile += "[-ERROR-] Error al escribir el Bloque\n"
									return
								}
								//Actualizar el super bloque
								super_bloque.S_first_blo += int32(binary.Size(siguiente_bloque))
								super_bloque.S_free_blocks_count -= 1
								archivo.Seek(int64(pos_sb), 0)
								err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
								if err != nil {
									consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
									return
								}
								var uno byte = 1
								//Actualizar el bitmap de bloques
								pos_bitmap := super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
								archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
								err = binary.Write(archivo, binary.LittleEndian, &uno)
								if err != nil {
									consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
									return
								}
								//Actualizar el inodo padre
								inodo_padre.I_block[i+1] = int32(pos_sig_bloque)
								archivo.Seek(int64(Bloque_prueba.B_content[0].B_inodo), 0)
								err = binary.Write(archivo, binary.LittleEndian, &inodo_padre)
								if err != nil {
									consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
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
									consola_mkfile += "[-ERROR-] Error al escribir el Inodo\n"
									return
								}
								//Actualizar el super bloque
								super_bloque.S_first_ino += int32(binary.Size(nuevo_inodo))
								super_bloque.S_free_inodes_count -= 1
								archivo.Seek(int64(pos_sb), 0)
								err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
								if err != nil {
									consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
									return
								}
								//Actualizar el bitmap de inodos
								pos_bitmap = super_bloque.S_inodes_count - super_bloque.S_free_inodes_count
								archivo.Seek(int64(super_bloque.S_bm_inode_start+pos_bitmap), 0)
								err = binary.Write(archivo, binary.LittleEndian, &uno)
								if err != nil {
									consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Inodos\n"
									return
								}
								//Actualizar el bloque de carpeta
								siguiente_bloque.B_content[2].B_inodo = int32(pos_nuevo_inodo)
								copy(siguiente_bloque.B_content[2].B_name[:], nueva_carpeta)
								archivo.Seek(int64(pos_sig_bloque), 0)
								err = binary.Write(archivo, binary.LittleEndian, &siguiente_bloque)
								if err != nil {
									consola_mkfile += "[-ERROR-] Error al escribir el Bloque\n"
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
									consola_mkfile += "[-ERROR-] Error al escribir el Bloque\n"
									return
								}
								//Actualizar el super bloque
								super_bloque.S_first_blo += int32(binary.Size(nuevo_bloque))
								super_bloque.S_free_blocks_count -= 1
								archivo.Seek(int64(pos_sb), 0)
								err = binary.Write(archivo, binary.LittleEndian, &super_bloque)
								if err != nil {
									consola_mkfile += "[-ERROR-] Error al escribir el Super Bloque\n"
									return
								}
								//Actualizar el bitmap de bloques
								pos_bitmap = super_bloque.S_blocks_count - super_bloque.S_free_blocks_count
								archivo.Seek(int64(super_bloque.S_bm_block_start+pos_bitmap), 0)
								err = binary.Write(archivo, binary.LittleEndian, &uno)
								if err != nil {
									consola_mkfile += "[-ERROR-] Error al escribir el Bitmap de Bloques\n"
									return
								}
								mkfile.CreacionRecursiva(nombres_carpetas, pos_sb, path)
							}
						}
					}
				}
			}
		}
	}

}

func (mkfile *Mkfile) ExisteCarpetaPadre(names []string, pos int, path string) bool {
	fmt.Println("NAMES: ", names)
	fmt.Println("POS: ", pos)
	archivo, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		consola_mkfile += "[-ERROR-] Error al abrir el archivo\n"
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
		consola_mkfile += "[-ERROR-] Error al leer el Inodo\n"
		return false
	}
	for i := 0; i < len(inodo.I_block); i++ {
		if inodo.I_block[i] != -1 {
			fmt.Println("Bloque: ", inodo.I_block[i])
			bloque := Bloque_Carpeta{}
			archivo.Seek(int64(inodo.I_block[i]), 0)
			err = binary.Read(archivo, binary.LittleEndian, &bloque)
			if err != nil {
				consola_mkfile += "[-ERROR-] Error al leer el Bloque Carpeta\n"
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
							retornar := mkfile.ExisteCarpetaPadre(names[1:], int(bloque.B_content[j].B_inodo), path)
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

func (mkfile *Mkfile) PosCarpetaPadre(names []string, pos int, path string) int32 {
	archivo, err := os.OpenFile(path, os.O_RDWR, 0666)
	if err != nil {
		consola_mkfile += "[-ERROR-] Error al abrir el archivo\n"
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
		consola_mkfile += "[-ERROR-] Error al leer el Inodo\n"
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
				consola_mkfile += "[-ERROR-] Error al leer el Bloque Carpeta\n"
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
							return mkfile.PosCarpetaPadre(names[1:], int(bloque.B_content[j].B_inodo), path)
						}
					}
				}
			}
		}
	}
	return 0
}

func (mkfile *Mkfile) RetornaContenidoSize() string {
	//Retorna el contenido del size
	contenido := ""
	contador := 0
	aux := 0
	for contador < mkfile.Size {
		if aux == 10 {
			aux = 0
		}
		aux2 := strconv.Itoa(aux)
		contenido += aux2
		aux++
		contador++
	}
	return contenido
}

func (mkfile *Mkfile) RetornarContenidoArchivoComputadora() string {
	data, err := ioutil.ReadFile(mkfile.Cont)
	if err != nil {
		consola_mkfile += "[-ERROR-] Error al leer el archivo CONT\n"
		return ""
	}
	return string(data)

}

func (mkfile *Mkfile) CadenaVacia(cadena [16]byte) bool {

	for _, v := range cadena {
		if v != 0 {
			return false
		}
	}
	return true

}

func (mkfile *Mkfile) VerificarID() bool {
	//Verificando si existe el id
	for i := 0; i < len(ParticionesMontadasList); i++ {
		if strings.Contains(strings.ToLower(ParticionesMontadasList[i].Id), strings.ToLower(Idlogueado)) {
			return true
		}
	}
	return false
}

func (mkfile *Mkfile) RetornarStrictMontada(id string) ParticionMontada {
	for i := 0; i < len(ParticionesMontadasList); i++ {
		if strings.Contains(strings.ToLower(ParticionesMontadasList[i].Id), strings.ToLower(id)) {
			return ParticionesMontadasList[i]
		}
	}
	return ParticionMontada{}
}

func (mkfile *Mkfile) IsParticionMontadaVacia(p ParticionMontada) bool {
	return !p.Sistema_archivos && p.Id == "" && p.Letra == "" && p.Numero == 0 && p.Path == "" && p.Type == "" && p.Name == ""
}

func (mkfile *Mkfile) ObtenerParticiones(Mbr MBR) []Partition {
	var particiones []Partition
	particiones = append(particiones, Mbr.Mbr_partition_1)
	particiones = append(particiones, Mbr.Mbr_partition_2)
	particiones = append(particiones, Mbr.Mbr_partition_3)
	particiones = append(particiones, Mbr.Mbr_partition_4)
	return particiones
}

func (mkfile *Mkfile) ListadoEBR(Extendida Partition, path string) []EBR {
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
		if !mkfile.CadenaVacia(ebr.Part_name) {
			ebrs = append(ebrs, ebr)
		} else if mkfile.CadenaVacia(ebr.Part_name) && ebr.Part_size != 0 {
			ebrs = append(ebrs, ebr)
		} else {
			break
		}
		temp = ebr.Part_next
	}
	return ebrs
}

func (mkfile *Mkfile) ExisteArchivo() bool {
	_, err := os.Stat(mkfile.Cont)
	if os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func RetornarConsolamkfile() string {
	return consola_mkfile
}
