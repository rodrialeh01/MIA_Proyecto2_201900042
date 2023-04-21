package analizador

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
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

	//Verificando si la particion ya existe
	/*if fdisk.ExisteParticion() {
		consola_fdisk += "[-ERROR-] La particion ya existe\n"
		return
	}*/

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

	if fdisk.Size > int(mbr.Mbr_tamano) {
		consola_fdisk += "[-ERROR-] El tamaño de la particion es mayor al tamaño del disco\n"
		return
	}

	particiones := fdisk.ObtenerParticiones(mbr)

	if fdisk.Type == "p" || fdisk.Type == "e" {
		switch fdisk.Fit {
		case "B":
			fdisk.BestFitPyE(particiones, mbr)
		case "F":
			fdisk.FirstFitPyE(particiones, mbr)
		case "W":
			fdisk.WorstFitPyE(particiones, mbr)
		}
	} else if fdisk.Type == "l" {
		fdisk.CrearParticionesLogicas()
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
			name := string(particiones[i].Part_name[:])
			if name != "" {
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

}

func (fdisk *Fdisk) WorstFitPyE(particiones []Partition, Mbr MBR) {

}

func (fdisk *Fdisk) CrearParticionesLogicas() {

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

func RetornarConsolafdisk() string {
	return consola_fdisk
}
