package analizador

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"unsafe"
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

	//Cambiando el path
	if fdisk.Path[0] == '"' && fdisk.Path[len(fdisk.Path)-1] == '"' {
		fdisk.Path = fdisk.Path[1 : len(fdisk.Path)-1]
	}

	//Verificando si el disco existe
	if !fdisk.ExisteDisco() {
		consola_fdisk += "[-ERROR-] El disco no existe\n"
		return
	}

	if fdisk.Unit == "k" {
		fdisk.Size = fdisk.Size * 1024
	} else if fdisk.Unit == "m" {
		fdisk.Size = fdisk.Size * 1024 * 1024
	}

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
	fmt.Println("Tamaño del MBR: ", binary.Size(MBR{}))
	archivo.Seek(0, 0)
	err = binary.Read(archivo, binary.LittleEndian, &mbr)
	if err != nil {
		fmt.Println("Error al leer el MBR: ", err)
		consola_fdisk += "[-ERROR-] No se pudo leer el MBR\n"
		return
	}

	//Verificando si se leyo bien el MBR
	fmt.Println("DESDE EL FDISK")
	fmt.Println("Fecha de creacion: ", string(mbr.Mbr_fecha_creacion[:]))
	fmt.Println("Tamaño del disco: ", mbr.Mbr_tamano)
	fmt.Println("Signature: ", mbr.Mbr_dsk_signature)
	fmt.Println("Fit: ", string(mbr.Mbr_fit[0]))

	//Verificando si la particion existe
	particiones_verif := fdisk.ObtenerParticiones(mbr)
	for i := 0; i < len(particiones_verif); i++ {
		if string(particiones_verif[i].Part_name[:]) == fdisk.Name {
			consola_fdisk += "[-ERROR-] La particion ya existe\n"
			return
		}
	}
	cantidad_primarias := 0
	cantidad_extendidas := 0

	for i := 0; i < len(particiones_verif); i++ {
		if string(particiones_verif[i].Part_type[0]) == "P" {
			cantidad_primarias++
		} else if string(particiones_verif[i].Part_type[0]) == "E" {
			cantidad_extendidas++
		}
	}

	if cantidad_primarias == 4 && fdisk.Type == "p" {
		consola_fdisk += "[-ERROR-] Ya existen 4 particiones primarias\n"
		return
	} else if cantidad_extendidas == 1 && fdisk.Type == "e" {
		consola_fdisk += "[-ERROR-] Ya existe una particion extendida\n"
		return
	} else if cantidad_extendidas == 0 && fdisk.Type == "l" {
		consola_fdisk += "[-ERROR-] No existe una particion extendida\n"
		return
	}

	if fdisk.Size > int(mbr.Mbr_tamano) {
		consola_fdisk += "[-ERROR-] El tamaño de la particion es mayor al tamaño del disco\n"
		return
	}

	particiones := fdisk.ObtenerParticiones(mbr)

	if fdisk.Type == "p" || fdisk.Type == "e" {
		switch string(mbr.Mbr_fit[0]) {
		case "B":
			fdisk.BestFitPyE(particiones, mbr)
		case "F":
			fdisk.FirstFitPyE(particiones, mbr)
		case "W":
			fdisk.WorstFitPyE(particiones, mbr)
		}
	} else if fdisk.Type == "l" {
		fdisk.CrearParticionesLogicas(particiones, mbr)
	}
}

func (fdisk *Fdisk) FirstFitPyE(particiones []Partition, Mbr MBR) {
	for i := 0; i < len(particiones); i++ {
		if particiones[i].Part_start == -1 {
			if i == 0 {
				particiones[i].Part_start = int32(binary.Size(MBR{}))
				particiones[i].Part_size = int32(fdisk.Size)
				if fdisk.Fit == "bf" {
					tipo := []byte{byte('B')}
					copy(particiones[i].Part_fit[:], tipo)
				} else if fdisk.Fit == "ff" {
					tipo := []byte{byte('F')}
					copy(particiones[i].Part_fit[:], tipo)
				} else if fdisk.Fit == "wf" {
					tipo := []byte{byte('W')}
					copy(particiones[i].Part_fit[:], tipo)
				}
				if fdisk.Type == "p" {
					tipo := []byte{byte('P')}
					copy(particiones[i].Part_type[:], tipo)
				} else if fdisk.Type == "e" {
					tipo := []byte{byte('E')}
					copy(particiones[i].Part_type[:], tipo)
				}
				name := []byte(fdisk.Name)
				copy(particiones[i].Part_name[:], name)
				fdisk.ActualizarMBRDisco(particiones, Mbr)
				return
			} else {
				tamanio_disponible := int(Mbr.Mbr_tamano) - int(particiones[i-1].Part_start) - int(particiones[i-1].Part_size)
				if tamanio_disponible >= fdisk.Size {
					particiones[i].Part_start = particiones[i-1].Part_start + particiones[i-1].Part_size
					particiones[i].Part_size = int32(fdisk.Size)
					if fdisk.Fit == "bf" {
						tipo := []byte{byte('B')}
						copy(particiones[i].Part_fit[:], tipo)
					} else if fdisk.Fit == "ff" {
						tipo := []byte{byte('F')}
						copy(particiones[i].Part_fit[:], tipo)
					} else if fdisk.Fit == "wf" {
						tipo := []byte{byte('W')}
						copy(particiones[i].Part_fit[:], tipo)
					}
					if fdisk.Type == "p" {
						tipo := []byte{byte('P')}
						copy(particiones[i].Part_type[:], tipo)
					} else if fdisk.Type == "e" {
						tipo := []byte{byte('E')}
						copy(particiones[i].Part_type[:], tipo)
					}
					name := []byte(fdisk.Name)
					copy(particiones[i].Part_name[:], name)
					fdisk.ActualizarMBRDisco(particiones, Mbr)
					return
				} else {
					consola_fdisk += "[-ERROR-] No hay espacio suficiente para crear la particion\n"
					return
				}
			}
		}
	}

	error_fdisk := true

	for i := 0; i < len(particiones); i++ {
		if particiones[i].Part_start != -1 {
			if len(particiones[i].Part_name) == 0 {
				if i != 3 {
					espacio_disponible := int(particiones[i+1].Part_start) - int(particiones[i].Part_start)
					if espacio_disponible >= fdisk.Size {
						particiones[i].Part_size = int32(fdisk.Size)
						if fdisk.Fit == "bf" {
							tipo := []byte{byte('B')}
							copy(particiones[i].Part_fit[:], tipo)
						} else if fdisk.Fit == "ff" {
							tipo := []byte{byte('F')}
							copy(particiones[i].Part_fit[:], tipo)
						} else if fdisk.Fit == "wf" {
							tipo := []byte{byte('W')}
							copy(particiones[i].Part_fit[:], tipo)
						}
						if fdisk.Type == "p" {
							tipo := []byte{byte('P')}
							copy(particiones[i].Part_type[:], tipo)
						} else if fdisk.Type == "e" {
							tipo := []byte{byte('E')}
							copy(particiones[i].Part_type[:], tipo)
						}
						name2 := []byte(fdisk.Name)
						copy(particiones[i].Part_name[:], name2)
						error_fdisk = false
						fdisk.ActualizarMBRDisco(particiones, Mbr)
						break
					}
				} else if i == 3 {
					espacio_disponible := int(Mbr.Mbr_tamano) - int(particiones[i].Part_start)
					if espacio_disponible >= fdisk.Size {
						particiones[i].Part_size = int32(fdisk.Size)
						if fdisk.Fit == "bf" {
							tipo := []byte{byte('B')}
							copy(particiones[i].Part_fit[:], tipo)
						} else if fdisk.Fit == "ff" {
							tipo := []byte{byte('F')}
							copy(particiones[i].Part_fit[:], tipo)
						} else if fdisk.Fit == "wf" {
							tipo := []byte{byte('W')}
							copy(particiones[i].Part_fit[:], tipo)
						}
						if fdisk.Type == "p" {
							tipo := []byte{byte('P')}
							copy(particiones[i].Part_type[:], tipo)
						} else if fdisk.Type == "e" {
							tipo := []byte{byte('E')}
							copy(particiones[i].Part_type[:], tipo)
						}
						name2 := []byte(fdisk.Name)
						copy(particiones[i].Part_name[:], name2)
						error_fdisk = false
						fdisk.ActualizarMBRDisco(particiones, Mbr)
						break
					}
				}
			}
		}
	}

	if error_fdisk {
		consola_fdisk += "[-ERROR-] No hay espacio suficiente para crear la particion\n"
	}

}

func (fdisk *Fdisk) BestFitPyE(particiones []Partition, Mbr MBR) {
	for i := 0; i < len(particiones); i++ {
		if particiones[i].Part_start == -1 {
			if i == 0 {
				particiones[i].Part_start = int32(binary.Size(Mbr))
				particiones[i].Part_size = int32(fdisk.Size)
				if fdisk.Fit == "bf" {
					tipo := []byte{byte('B')}
					copy(particiones[i].Part_fit[:], tipo)
				} else if fdisk.Fit == "ff" {
					tipo := []byte{byte('F')}
					copy(particiones[i].Part_fit[:], tipo)
				} else if fdisk.Fit == "wf" {
					tipo := []byte{byte('W')}
					copy(particiones[i].Part_fit[:], tipo)
				}
				if fdisk.Type == "p" {
					tipo := []byte{byte('P')}
					copy(particiones[i].Part_type[:], tipo)
				} else if fdisk.Type == "e" {
					tipo := []byte{byte('E')}
					copy(particiones[i].Part_type[:], tipo)
				}
				name := []byte(fdisk.Name)
				copy(particiones[i].Part_name[:], name)
				fdisk.ActualizarMBRDisco(particiones, Mbr)
				return
			} else {
				tamanio_disponible := int(Mbr.Mbr_tamano) - int(particiones[i-1].Part_start) - int(particiones[i-1].Part_size)
				if tamanio_disponible >= fdisk.Size {
					particiones[i].Part_start = particiones[i-1].Part_start + particiones[i-1].Part_size
					particiones[i].Part_size = int32(fdisk.Size)
					if fdisk.Fit == "bf" {
						tipo := []byte{byte('B')}
						copy(particiones[i].Part_fit[:], tipo)
					} else if fdisk.Fit == "ff" {
						tipo := []byte{byte('F')}
						copy(particiones[i].Part_fit[:], tipo)
					} else if fdisk.Fit == "wf" {
						tipo := []byte{byte('W')}
						copy(particiones[i].Part_fit[:], tipo)
					}
					if fdisk.Type == "p" {
						tipo := []byte{byte('P')}
						copy(particiones[i].Part_type[:], tipo)
					} else if fdisk.Type == "e" {
						tipo := []byte{byte('E')}
						copy(particiones[i].Part_type[:], tipo)
					}
					name := []byte(fdisk.Name)
					copy(particiones[i].Part_name[:], name)
					fdisk.ActualizarMBRDisco(particiones, Mbr)
					return
				} else {
					consola_fdisk += "[-ERROR-] No hay espacio suficiente para crear la particion\n"
					return
				}
			}
		}
	}

	espacio_pequenio := 1 * 1024 * 1024 * 1024
	no_particion := 0
	for i := 0; i < len(particiones); i++ {
		if particiones[i].Part_start != -1 {
			if len(particiones[i].Part_name) == 0 {
				espacio_pequenio = int(particiones[i].Part_size)
				no_particion = i
			}
		}
	}

	if espacio_pequenio == 1*1024*1024*1024 {
		consola_fdisk += "[-ERROR-] No hay espacio suficiente para crear la particion\n"
		return
	} else {
		particiones[no_particion].Part_size = int32(fdisk.Size)
		if fdisk.Fit == "bf" {
			tipo := []byte{byte('B')}
			copy(particiones[no_particion].Part_fit[:], tipo)
		} else if fdisk.Fit == "ff" {
			tipo := []byte{byte('F')}
			copy(particiones[no_particion].Part_fit[:], tipo)
		} else if fdisk.Fit == "wf" {
			tipo := []byte{byte('W')}
			copy(particiones[no_particion].Part_fit[:], tipo)
		}
		if fdisk.Type == "p" {
			tipo := []byte{byte('P')}
			copy(particiones[no_particion].Part_type[:], tipo)
		} else if fdisk.Type == "e" {
			tipo := []byte{byte('E')}
			copy(particiones[no_particion].Part_type[:], tipo)
		}
		name := []byte(fdisk.Name)
		copy(particiones[no_particion].Part_name[:], name)
		fdisk.ActualizarMBRDisco(particiones, Mbr)
		return
	}
}

func (fdisk *Fdisk) WorstFitPyE(particiones []Partition, Mbr MBR) {
	for i := 0; i < len(particiones); i++ {
		if particiones[i].Part_start == -1 {
			if i == 0 {
				particiones[i].Part_start = int32(unsafe.Sizeof(Mbr))
				particiones[i].Part_size = int32(fdisk.Size)
				if fdisk.Fit == "bf" {
					tipo := []byte{byte('B')}
					copy(particiones[i].Part_fit[:], tipo)
				} else if fdisk.Fit == "ff" {
					tipo := []byte{byte('F')}
					copy(particiones[i].Part_fit[:], tipo)
				} else if fdisk.Fit == "wf" {
					tipo := []byte{byte('W')}
					copy(particiones[i].Part_fit[:], tipo)
				}
				if fdisk.Type == "p" {
					tipo := []byte{byte('P')}
					copy(particiones[i].Part_type[:], tipo)
				} else if fdisk.Type == "e" {
					tipo := []byte{byte('E')}
					copy(particiones[i].Part_type[:], tipo)
				}
				name := []byte(fdisk.Name)
				copy(particiones[i].Part_name[:], name)
				fdisk.ActualizarMBRDisco(particiones, Mbr)
				return
			} else {
				tamanio_disponible := int(Mbr.Mbr_tamano) - int(particiones[i-1].Part_start) - int(particiones[i-1].Part_size)
				if tamanio_disponible >= fdisk.Size {
					particiones[i].Part_start = particiones[i-1].Part_start + particiones[i-1].Part_size
					particiones[i].Part_size = int32(fdisk.Size)
					if fdisk.Fit == "bf" {
						tipo := []byte{byte('B')}
						copy(particiones[i].Part_fit[:], tipo)
					} else if fdisk.Fit == "ff" {
						tipo := []byte{byte('F')}
						copy(particiones[i].Part_fit[:], tipo)
					} else if fdisk.Fit == "wf" {
						tipo := []byte{byte('W')}
						copy(particiones[i].Part_fit[:], tipo)
					}
					if fdisk.Type == "p" {
						tipo := []byte{byte('P')}
						copy(particiones[i].Part_type[:], tipo)
					} else if fdisk.Type == "e" {
						tipo := []byte{byte('E')}
						copy(particiones[i].Part_type[:], tipo)
					}
					name := []byte(fdisk.Name)
					copy(particiones[i].Part_name[:], name)
					fdisk.ActualizarMBRDisco(particiones, Mbr)
					return
				} else {
					consola_fdisk += "[-ERROR-] No hay espacio suficiente para crear la particion\n"
					return
				}
			}
		}
	}

	espacio_grande := 0
	no_particion := 0
	for i := 0; i < len(particiones); i++ {
		if particiones[i].Part_start != -1 {
			if len(particiones[i].Part_name) == 0 {
				if particiones[i].Part_size > int32(espacio_grande) {
					espacio_grande = int(particiones[i].Part_size)
					no_particion = i
				}
			}
		}
	}

	if espacio_grande == 0 {
		consola_fdisk += "[-ERROR-] No hay espacio suficiente para crear la particion\n"
		return
	} else {
		particiones[no_particion].Part_size = int32(fdisk.Size)
		if fdisk.Fit == "bf" {
			tipo := []byte{byte('B')}
			copy(particiones[no_particion].Part_fit[:], tipo)
		} else if fdisk.Fit == "ff" {
			tipo := []byte{byte('F')}
			copy(particiones[no_particion].Part_fit[:], tipo)
		} else if fdisk.Fit == "wf" {
			tipo := []byte{byte('W')}
			copy(particiones[no_particion].Part_fit[:], tipo)
		}
		if fdisk.Type == "p" {
			tipo := []byte{byte('P')}
			copy(particiones[no_particion].Part_type[:], tipo)
		} else if fdisk.Type == "e" {
			tipo := []byte{byte('E')}
			copy(particiones[no_particion].Part_type[:], tipo)
		}
		name := []byte(fdisk.Name)
		copy(particiones[no_particion].Part_name[:], name)
		fdisk.ActualizarMBRDisco(particiones, Mbr)
		return
	}
}

func (fdisk *Fdisk) CrearParticionesLogicas(particiones []Partition, Mbr MBR) {
	error_particion := true
	particion_id := 0
	for i := 0; i < len(particiones); i++ {
		if particiones[i].Part_type == [1]byte{byte('E')} {
			error_particion = false
			particion_id = i
		}
	}

	if error_particion {
		consola_fdisk += "[-ERROR-] No existe una particion extendida para crear la particion logica\n"
		return
	}

	//POSICION INICIAL DE LA PARTICION EXTENDIDA
	inicio := particiones[particion_id].Part_start
	//POSICION FINAL DE LA PARTICION EXTENDIDA
	final := particiones[particion_id].Part_start + particiones[particion_id].Part_size
	//TAMAÑO DE LA PARTICION EXTENDIDA
	tamanio := int(particiones[particion_id].Part_size)

	if fdisk.Size > tamanio {
		consola_fdisk += "[-ERROR-] No hay espacio suficiente para crear la particion\n"
		return
	}

	logicas := fdisk.ListadoEBR(particiones[particion_id], fdisk.Path)
	//VERIFICA SI EXISTE UNA PARTICION LOGICA CON EL MISMO NOMBRE
	for i := 0; i < len(logicas); i++ {
		if string(logicas[i].Part_name[:]) == fdisk.Name {
			consola_fdisk += "[-ERROR-] Ya existe una particion con el mismo nombre\n"
			return
		}
	}

	//CREA LA PARTICION LOGICA
	if len(logicas) == 0 {
		ebr := EBR{}
		status := []byte{byte('0')}
		copy(ebr.Part_status[:], status)
		if fdisk.Fit == "bf" {
			tipo := []byte{byte('B')}
			copy(ebr.Part_fit[:], tipo)
		} else if fdisk.Fit == "ff" {
			tipo := []byte{byte('F')}
			copy(ebr.Part_fit[:], tipo)
		} else if fdisk.Fit == "wf" {
			tipo := []byte{byte('W')}
			copy(ebr.Part_fit[:], tipo)
		}
		namexd := []byte(fdisk.Name)
		copy(ebr.Part_name[:], namexd)
		ebr.Part_size = int32(fdisk.Size)
		ebr.Part_start = inicio
		ebr.Part_next = -1
		fdisk.AgregarEBR(ebr, fdisk.Path)
	} else {
		ebr := EBR{}
		status := []byte{byte('0')}
		copy(ebr.Part_status[:], status)
		if fdisk.Fit == "bf" {
			tipo := []byte{byte('B')}
			copy(ebr.Part_fit[:], tipo)
		} else if fdisk.Fit == "ff" {
			tipo := []byte{byte('F')}
			copy(ebr.Part_fit[:], tipo)
		} else if fdisk.Fit == "wf" {
			tipo := []byte{byte('W')}
			copy(ebr.Part_fit[:], tipo)
		}

		if particiones[particion_id].Part_fit == [1]byte{byte('F')} {
			ebr.Part_fit = [1]byte{byte('F')}
			if fdisk.FirstFit_Logicas(logicas, fdisk.Size, int(final)) != -1 {
				ebr.Part_start = int32(fdisk.FirstFit_Logicas(logicas, fdisk.Size, int(final)))
			} else {
				consola_fdisk += "[-ERROR-] No hay espacio suficiente para crear la particion\n"
				return
			}
		} else if particiones[particion_id].Part_fit == [1]byte{byte('B')} {
			ebr.Part_fit = [1]byte{byte('B')}
			if fdisk.BestFit_Logicas(logicas, fdisk.Size, int(final)) != -1 {
				ebr.Part_start = int32(fdisk.BestFit_Logicas(logicas, fdisk.Size, int(final)))
			} else {
				consola_fdisk += "[-ERROR-] No hay espacio suficiente para crear la particion\n"
				return
			}
		} else if particiones[particion_id].Part_fit == [1]byte{byte('W')} {
			ebr.Part_fit = [1]byte{byte('W')}
			if fdisk.WorstFit_Logicas(logicas, fdisk.Size, int(final)) != -1 {
				ebr.Part_start = int32(fdisk.WorstFit_Logicas(logicas, fdisk.Size, int(final)))
			} else {
				consola_fdisk += "[-ERROR-] No hay espacio suficiente para crear la particion\n"
				return
			}
		}
		namexd := []byte(fdisk.Name)
		copy(ebr.Part_name[:], namexd)
		ebr.Part_size = int32(fdisk.Size)
		for i := 0; i < len(logicas); i++ {
			if logicas[i].Part_next == -1 {
				ebr.Part_next = -1
				logicas[i].Part_next = int32(ebr.Part_start)
				fdisk.ActualizarEBR(ebr, fdisk.Path)
				break
			} else if logicas[i].Part_next != -1 {
				if logicas[i].Part_start == ebr.Part_start {
					ebr.Part_next = logicas[i].Part_next
					logicas[i].Part_next = int32(ebr.Part_start)
					break
				}
			}
		}
		ebr.Part_next = -1
		fdisk.AgregarEBR(ebr, fdisk.Path)
	}

}

func (fdisk *Fdisk) ListadoEBR(Extendida Partition, path string) []EBR {
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
		if !fdisk.CadenaVacia(ebr.Part_name) {
			ebrs = append(ebrs, ebr)
		} else if fdisk.CadenaVacia(ebr.Part_name) && ebr.Part_size != 0 {
			ebrs = append(ebrs, ebr)
		} else {
			break
		}
		temp = ebr.Part_next
	}
	return ebrs
}

func (fdisk *Fdisk) AgregarEBR(ebr EBR, path string) {
	fmt.Println("============= EBR===============")
	fmt.Println(string(ebr.Part_fit[0]))
	fmt.Println(string(ebr.Part_name[:]))
	fmt.Println(int32(ebr.Part_next))
	fmt.Println(ebr.Part_size)
	fmt.Println(ebr.Part_start)
	fmt.Println("=================================")

	archivo1, err1 := os.OpenFile(path, os.O_RDWR, 0666)
	if err1 != nil {
		consola_fdisk += "[-ERROR-] No se pudo abrir el disco\n"
		return
	}
	defer archivo1.Close()

	archivo1.Seek(int64(ebr.Part_start), 0)
	err := binary.Write(archivo1, binary.LittleEndian, &ebr)
	if err != nil {
		consola_fdisk += "[-ERROR-] No se pudo escribir el EBR\n"
		fmt.Println(err)
		return
	}
	consola_fdisk += "[*SUCCESS*] Partición Lógica creada con éxito\n"

}

func (fdisk *Fdisk) ActualizarEBR(ebr EBR, path string) {
	archivo2, _ := os.OpenFile(path, os.O_RDWR, 0666)
	defer archivo2.Close()

	archivo2.Seek(int64(ebr.Part_start), 0)
	err := binary.Write(archivo2, binary.LittleEndian, &ebr)
	if err != nil {
		consola_fdisk += "[-ERROR-] No se pudo actualizar el EBR\n"
		fmt.Println(err)
		return
	}
}

func (fdisk *Fdisk) FirstFit_Logicas(ebrs []EBR, tamanio int, final_pe int) int {
	inicio := -1
	for i := 0; i < len(ebrs); i++ {
		if i != len(ebrs)-1 {
			if fdisk.CadenaVacia(ebrs[i].Part_name) && (ebrs[i].Part_size-ebrs[i].Part_start) >= int32(tamanio) {
				return inicio
			}
		} else {
			if fdisk.CadenaVacia(ebrs[i].Part_name) && (int32(final_pe)-ebrs[i].Part_start) >= int32(tamanio) {
				return inicio
			} else if ebrs[i].Part_next == -1 {
				if (int32(final_pe) - (ebrs[i].Part_start + ebrs[i].Part_size)) >= int32(tamanio) {
					inicio = int(ebrs[i].Part_start + ebrs[i].Part_size)
					return inicio
				}
			}
		}
		inicio = int(ebrs[i].Part_next)
	}
	return inicio
}

func (fdisk *Fdisk) BestFit_Logicas(ebrs []EBR, tamanio int, final_pe int) int {
	mejor_ajuste := 999999999
	mejor_inicio := -1
	for i := 0; i < len(ebrs); i++ {
		if i != len(ebrs)-1 {
			if fdisk.CadenaVacia(ebrs[i].Part_name) && (ebrs[i].Part_size) >= int32(tamanio) {
				if ebrs[i].Part_size < int32(mejor_ajuste) {
					mejor_ajuste = int(ebrs[i].Part_size)
					mejor_inicio = int(ebrs[i].Part_start)
				}
			}
		} else {
			if fdisk.CadenaVacia(ebrs[i].Part_name) && (int32(final_pe)-ebrs[i].Part_start) >= int32(tamanio) {
				if (int32(final_pe) - ebrs[i].Part_start) < int32(mejor_ajuste) {
					mejor_ajuste = final_pe - int(ebrs[i].Part_start)
					mejor_inicio = int(ebrs[i].Part_start)
				}
			} else if ebrs[i].Part_next == -1 {
				if final_pe-(int(ebrs[i].Part_start)+int(ebrs[i].Part_size)) < mejor_ajuste {
					mejor_ajuste = final_pe - (int(ebrs[i].Part_start) + int(ebrs[i].Part_size))
					mejor_inicio = int(ebrs[i].Part_start) + int(ebrs[i].Part_size)
				}
			}
		}
	}
	return mejor_inicio
}

func (fdisk *Fdisk) WorstFit_Logicas(ebrs []EBR, tamanio int, final_pe int) int {
	peor_ajuste := 0
	peor_inicio := -1
	for i := 0; i < len(ebrs); i++ {
		if i != len(ebrs)-1 {
			if fdisk.CadenaVacia(ebrs[i].Part_name) && (ebrs[i].Part_size) >= int32(tamanio) {
				if ebrs[i].Part_size > int32(peor_ajuste) {
					peor_ajuste = int(ebrs[i].Part_size)
					peor_inicio = int(ebrs[i].Part_start)
				}
			}
		} else {
			if fdisk.CadenaVacia(ebrs[i].Part_name) && (int32(final_pe)-ebrs[i].Part_start) >= int32(tamanio) {
				if (final_pe - int(ebrs[i].Part_start)) > peor_ajuste {
					peor_ajuste = final_pe - int(ebrs[i].Part_start)
					peor_inicio = int(ebrs[i].Part_start)
					break
				}
			} else if ebrs[i].Part_next == -1 {
				if final_pe-(int(ebrs[i].Part_start)+int(ebrs[i].Part_size)) > peor_ajuste {
					peor_ajuste = final_pe - (int(ebrs[i].Part_start) + int(ebrs[i].Part_size))
					peor_inicio = int(ebrs[i].Part_start) + int(ebrs[i].Part_size)
					break
				}
			}
		}
	}
	return peor_inicio

}

func (fdisk *Fdisk) ActualizarMBRDisco(particiones []Partition, Mbr MBR) {
	Mbr.Mbr_partition_1 = particiones[0]
	Mbr.Mbr_partition_2 = particiones[1]
	Mbr.Mbr_partition_3 = particiones[2]
	Mbr.Mbr_partition_4 = particiones[3]

	archivo, err := os.OpenFile(fdisk.Path, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer archivo.Close()

	archivo.Seek(0, 0)
	err = binary.Write(archivo, binary.LittleEndian, &Mbr)
	if err != nil {
		consola_fdisk += "[-ERROR-] No se pudo crear el disco\n"
		return
	}
	consola_fdisk += "[*SUCCESS*] Partición creada con exito\n"
	return
}

func (fdisk *Fdisk) ObtenerParticiones(Mbr MBR) []Partition {
	var particiones []Partition
	particiones = append(particiones, Mbr.Mbr_partition_1)
	particiones = append(particiones, Mbr.Mbr_partition_2)
	particiones = append(particiones, Mbr.Mbr_partition_3)
	particiones = append(particiones, Mbr.Mbr_partition_4)
	return particiones
}

func (fdisk *Fdisk) ExisteDisco() bool {
	_, err := os.Stat(fdisk.Path)
	if os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func (fdisk *Fdisk) CadenaVacia(cadena [16]byte) bool {

	for _, v := range cadena {
		if v != 0 {
			return false
		}
	}
	return true

}

func RetornarConsolafdisk() string {
	return consola_fdisk
}
