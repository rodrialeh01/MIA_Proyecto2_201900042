package analizador

import (
	"encoding/binary"
	"os"
	"strconv"
	"strings"
)

type Mount struct {
	Path string
	Name string
}

var consola_mount string
var letras = [26]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}

func (mount *Mount) VerificarParams(parametros map[string]string) {
	consola_mount = ""
	//Verificando parametros obligatorios
	if mount.Path == "" {
		consola_mount += "[-ERROR-] Falta el parametro path\n"
		return
	}
	if mount.Name == "" {
		consola_mount += "[-ERROR-] Falta el parametro name\n"
		return
	}

	//Cambiando el path
	if mount.Path[0] == '"' && mount.Path[len(mount.Path)-1] == '"' {
		mount.Path = mount.Path[1 : len(mount.Path)-1]
	}

	if !mount.ExisteDisco() {
		consola_mount += "[-ERROR-] El disco no existe\n"
		return
	}

	mount.MontarParticion()
}

func (mount *Mount) MontarParticion() {
	archivo, err := os.Open(mount.Path)
	if err != nil {
		consola_mount += "[-ERROR-] No se pudo leer el disco\n"
		return
	}
	defer archivo.Close()

	//Lee el MBR
	var Mbr MBR
	archivo.Seek(0, 0)
	err = binary.Read(archivo, binary.BigEndian, &Mbr)
	if err != nil {
		consola_mount += "[-ERROR-] No se pudo leer el disco\n"
		return
	}

	//Obtiene las particiones
	particiones := mount.ObtenerParticiones(Mbr)
	extendid := false

	for i := 0; i < 4; i++ {
		if strings.Contains(string(particiones[i].Part_type[0]), "P") {
			if strings.Contains(string(particiones[i].Part_name[:]), mount.Name) {
				if mount.ExisteParticionMontada(mount.Name, mount.Path) {
					consola_mount += "[-ERROR-] La particion ya esta montada\n"
					return
				}
				no_disk := strconv.Itoa(mount.CrearNoDisco(mount.Path))
				id := "42" + no_disk + mount.CrearLetraParticion(mount.Path)
				nueva_montada := ParticionMontada{mount.Path, mount.Name, id, "P", mount.CrearLetraParticion(mount.Path), false, mount.CrearNoDisco(mount.Path)}
				ParticionesMontadasList = append(ParticionesMontadasList, nueva_montada)
				break
			}
		} else if strings.Contains(string(particiones[i].Part_type[0]), "E") {
			if strings.Contains(string(particiones[i].Part_name[:]), mount.Name) {
				if mount.ExisteParticionMontada(mount.Name, mount.Path) {
					consola_mount += "[-ERROR-] La particion ya esta montada\n"
					return
				}
				no_disk := strconv.Itoa(mount.CrearNoDisco(mount.Path))
				id := "42" + no_disk + mount.CrearLetraParticion(mount.Path)
				nueva_montada := ParticionMontada{mount.Path, mount.Name, id, "E", mount.CrearLetraParticion(mount.Path), false, mount.CrearNoDisco(mount.Path)}
				ParticionesMontadasList = append(ParticionesMontadasList, nueva_montada)
				extendid = true
				break
			} else {
				ebrs := mount.ListadoEBR(particiones[i], mount.Path)
				for j := 0; j < len(ebrs); j++ {
					if strings.Contains(string(ebrs[j].Part_name[:]), mount.Name) {
						if mount.ExisteParticionMontada(mount.Name, mount.Path) {
							consola_mount += "[-ERROR-] La particion ya esta montada\n"
							return
						}
						no_disk := strconv.Itoa(mount.CrearNoDisco(mount.Path))
						id := "42" + no_disk + mount.CrearLetraParticion(mount.Path)
						nueva_montada := ParticionMontada{mount.Path, mount.Name, id, "L", mount.CrearLetraParticion(mount.Path), false, mount.CrearNoDisco(mount.Path)}
						ParticionesMontadasList = append(ParticionesMontadasList, nueva_montada)
						break
					}
				}
			}
		}

	}

	if extendid {
		consola_mount += "[/\\WARNING/\\] Se montó una partición extendida, al hacer esto unicamente podra operar al visualizar los reportes\n"
	}

	consola_mount += "[*SUCCESS*] La partición se ha montado correctamente\n"
	mount.MostrarParticionesMontadas()
}

func (mount *Mount) ExisteDisco() bool {
	_, err := os.Stat(mount.Path)
	if os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func (mount *Mount) ObtenerParticiones(Mbr MBR) []Partition {
	var particiones []Partition
	particiones = append(particiones, Mbr.Mbr_partition_1)
	particiones = append(particiones, Mbr.Mbr_partition_2)
	particiones = append(particiones, Mbr.Mbr_partition_3)
	particiones = append(particiones, Mbr.Mbr_partition_4)
	return particiones
}

func (mount *Mount) ListadoEBR(Extendida Partition, path string) []EBR {
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
		if !mount.CadenaVacia(ebr.Part_name) {
			ebrs = append(ebrs, ebr)
		} else if mount.CadenaVacia(ebr.Part_name) && ebr.Part_size != 0 {
			ebrs = append(ebrs, ebr)
		} else {
			break
		}
		temp = ebr.Part_next
	}
	return ebrs
}

func (mount *Mount) ExisteParticionMontada(name string, path string) bool {
	if len(ParticionesMontadasList) == 0 {
		return false
	} else {
		for i := 0; i < len(ParticionesMontadasList); i++ {
			if strings.Contains(ParticionesMontadasList[i].Path, path) && strings.Contains(ParticionesMontadasList[i].Name, name) {
				return true
			}
		}
		return false
	}
}

func (mount *Mount) CrearLetraParticion(path string) string {
	posicion := 0
	for i := 0; i < len(ParticionesMontadasList); i++ {
		if strings.Contains(ParticionesMontadasList[i].Path, path) {
			posicion++
		}
	}
	return letras[posicion]
}

func (mount *Mount) CrearNoDisco(path string) int {
	if len(ParticionesMontadasList) == 0 {
		return 1
	} else {
		for i := 0; i < len(ParticionesMontadasList); i++ {
			if strings.Contains(ParticionesMontadasList[i].Path, path) {
				return ParticionesMontadasList[i].Numero
			}
		}
		return mount.ValorMaxDisk() + 1
	}
}

func (mount *Mount) MostrarParticionesMontadas() {
	consola_mount += "\n"
	consola_mount += "!------------PARTICIONES MONTADAS------------!\n"
	for i := 0; i < len(ParticionesMontadasList); i++ {
		consola_mount += "\t\tID: " + ParticionesMontadasList[i].Id + "\n"
	}
	consola_mount += "!--------------------------------------------!\n\n"
}

func (mount *Mount) ValorMaxDisk() int {
	valor := ParticionesMontadasList[0].Numero
	for _, num := range ParticionesMontadasList {
		if num.Numero > valor {
			valor = num.Numero
		}
	}
	return valor
}

func (mount *Mount) CadenaVacia(cadena [16]byte) bool {

	for _, v := range cadena {
		if v != 0 {
			return false
		}
	}
	return true

}

func RetornarConsolamount() string {
	return consola_mount
}
