package analizador

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

type Login struct {
	User string
	Pwd  string
	Id   string
}

var consola_login string

func (login *Login) VerificarParams(parametros map[string]string) {
	//Verificando parametros obligatorios
	consola_login = ""
	if login.User == "" {
		consola_login += "[-ERROR-] Falta el parametro user\n"
		return
	}
	if login.Pwd == "" {
		consola_login += "[-ERROR-] Falta el parametro pwd\n"
		return
	}
	if login.Id == "" {
		consola_login += "[-ERROR-] Falta el parametro id\n"
		return
	}

	if login.User[0] == '"' && login.User[len(login.User)-1] == '"' {
		login.User = login.User[1 : len(login.User)-1]
	}

	if login.Pwd[0] == '"' && login.Pwd[len(login.Pwd)-1] == '"' {
		login.Pwd = login.Pwd[1 : len(login.Pwd)-1]
	}

	if login.Id[0] == '"' && login.Id[len(login.Id)-1] == '"' {
		login.Id = login.Id[1 : len(login.Id)-1]
	}

	if !login.VerificarID() {
		consola_login += "[-ERROR-] No existe la particion con el id: " + login.Id + "\n"
		return
	}

	login.IniciarSesion()
}

func (login *Login) IniciarSesion() {
	montada := login.RetornarStrictMontada(login.Id)
	if login.IsParticionMontadaVacia(montada) {
		consola_login += "[-ERROR-] La partición con id: " + login.Id + " no está montada\n"
		return
	}

	if !montada.Sistema_archivos {
		consola_login += "[-ERROR-] La partición con id: " + login.Id + " no tiene un sistema de archivos\n"
	}

	//Abrir el archivo binario
	archivo, err := os.OpenFile(montada.Path, os.O_RDWR, 0666)
	if err != nil {
		consola_login += "[-ERROR-] Error al abrir el archivo\n"
		return
	}
	defer archivo.Close()

	//Leer el MBR
	mbr := MBR{}
	archivo.Seek(int64(0), 0)
	err = binary.Read(archivo, binary.LittleEndian, &mbr)
	if err != nil {
		consola_login += "[-ERROR-] Error al leer el MBR\n"
		return
	}
	fmt.Println("MBR DESDE REP")
	fmt.Println(mbr)
	particiones := login.ObtenerParticiones(mbr)
	var ebrs []EBR
	logica_existe := false
	var particion_logica EBR
	for i := 0; i < len(particiones); i++ {
		if strings.ToLower(string(particiones[i].Part_type[0])) == "e" {
			ebrs = login.ListadoEBR(particiones[i], montada.Path)
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
		consola_login += "[-ERROR-] Error al leer el SuperBloque\n"
		return
	}

	//Leer el Inodo del archivo users.txt
	inodo_users := Inodo{}
	pos_inodo := super_bloque.S_inode_start + int32(binary.Size(Inodo{}))
	archivo.Seek(int64(pos_inodo), 0)
	err = binary.Read(archivo, binary.LittleEndian, &inodo_users)
	if err != nil {
		consola_login += "[-ERROR-] Error al leer el Inodo del archivo users.txt\n"
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
			consola_login += "[-ERROR-] Error al leer el Bloque de datos del archivo users.txt\n"
			return
		}
		usuariostxt += string(bloque.B_content[:])
	}

	usuariostxt = strings.Replace(usuariostxt, "\u0000", "", -1)

	usuarios_grupos := strings.Split(usuariostxt, "\n")
	for i := 0; i < len(usuarios_grupos); i++ {
		datos := strings.Split(usuarios_grupos[i], ",")
		if len(datos) > 1 {
			if strings.Contains(datos[1], "U") {
				if strings.Contains(datos[3], login.User) {
					if strings.Contains(datos[4], login.Pwd) {
						if Idlogueado == "" {
							if !montada.Logueado {
								montada.Logueado = true
								montada.User = login.User
								montada.Password = login.Pwd
								Idlogueado = montada.Id
								for i := 0; i < len(ParticionesMontadasList); i++ {
									if ParticionesMontadasList[i].Id == montada.Id {
										ParticionesMontadasList[i] = montada
										break
									}
								}
								consola_login += "[*SUCCESS*] Se ha iniciado sesión correctamente con el usuario: " + login.User + " (Esto significa que unicamente has iniciado sesión en el sistema de comandos)\n"
								return
							} else {
								consola_login += "[-ERROR-] Ya existe una sesión iniciada en el sistema de comandos\n"
								return
							}
						} else {
							consola_login += "[-ERROR-] Ya existe una sesión iniciada en el sistema de comandos\n"
							return
						}
					} else {
						consola_login += "[-ERROR-] Contraseña incorrecta\n"
						return
					}
				}
			}
		}
	}

	consola_login += "[-ERROR-] No se encontró al usuario\n"
}

func (login *Login) CadenaVacia(cadena [16]byte) bool {

	for _, v := range cadena {
		if v != 0 {
			return false
		}
	}
	return true

}

func (login *Login) VerificarID() bool {
	//Verificando si existe el id
	for i := 0; i < len(ParticionesMontadasList); i++ {
		if strings.Contains(strings.ToLower(ParticionesMontadasList[i].Id), strings.ToLower(login.Id)) {
			return true
		}
	}
	return false
}

func (login *Login) RetornarStrictMontada(id string) ParticionMontada {
	for i := 0; i < len(ParticionesMontadasList); i++ {
		if strings.Contains(strings.ToLower(ParticionesMontadasList[i].Id), strings.ToLower(id)) {
			return ParticionesMontadasList[i]
		}
	}
	return ParticionMontada{}
}

func (login *Login) IsParticionMontadaVacia(p ParticionMontada) bool {
	return !p.Sistema_archivos && p.Id == "" && p.Letra == "" && p.Numero == 0 && p.Path == "" && p.Type == "" && p.Name == ""
}

func (login *Login) ObtenerParticiones(Mbr MBR) []Partition {
	var particiones []Partition
	particiones = append(particiones, Mbr.Mbr_partition_1)
	particiones = append(particiones, Mbr.Mbr_partition_2)
	particiones = append(particiones, Mbr.Mbr_partition_3)
	particiones = append(particiones, Mbr.Mbr_partition_4)
	return particiones
}

func (login *Login) ListadoEBR(Extendida Partition, path string) []EBR {
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
		if !login.CadenaVacia(ebr.Part_name) {
			ebrs = append(ebrs, ebr)
		} else if login.CadenaVacia(ebr.Part_name) && ebr.Part_size != 0 {
			ebrs = append(ebrs, ebr)
		} else {
			break
		}
		temp = ebr.Part_next
	}
	return ebrs
}

func RetornarConsolalogin() string {
	return consola_login
}
