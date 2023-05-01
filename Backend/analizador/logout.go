package analizador

import (
	"strings"
)

type Logout struct {
}

var consola_logout string

func (logout *Logout) VerificarParams() {
	consola_logout = ""
	logout.CerrarSesion()
}

func (logout *Logout) CerrarSesion() {
	if Idlogueado == "" {
		consola_logout += "[-ERROR-] No hay ninguna sesión iniciada\n"
		return
	}

	//Verificando si existe el id
	if !logout.VerificarID() {
		consola_logout += "[-ERROR-] No existe la particion con el id: " + Idlogueado + "\n"
		return
	}

	//Obteniendo particion montada
	for i := 0; i < len(ParticionesMontadasList); i++ {
		if strings.Contains(strings.ToLower(ParticionesMontadasList[i].Id), strings.ToLower(Idlogueado)) {
			if ParticionesMontadasList[i].Sistema_archivos {
				if ParticionesMontadasList[i].Logueado {
					userlogout := ParticionesMontadasList[i].User
					ParticionesMontadasList[i].Logueado = false
					ParticionesMontadasList[i].User = ""
					ParticionesMontadasList[i].Password = ""
					Idlogueado = ""
					Id_UserLogueado = 0
					Id_GroupLogueado = 0
					consola_logout += "[*SUCCESS*] Ha Cerrado Sesión correctamente usuario " + userlogout + "\n"
					return
				} else {
					consola_logout += "[-ERROR-] No hay ninguna sesión iniciada\n"
					return
				}
			} else {
				consola_logout += "[-ERROR-] La particion no tiene un sistema de archivos montado\n"
				return
			}
		}
	}
	consola_logout += "[-ERROR-] No existe la particion con el id: " + Idlogueado + "\n"
}

func (logout *Logout) CadenaVacia(cadena [16]byte) bool {

	for _, v := range cadena {
		if v != 0 {
			return false
		}
	}
	return true

}

func (logout *Logout) VerificarID() bool {
	//Verificando si existe el id
	for i := 0; i < len(ParticionesMontadasList); i++ {
		if strings.Contains(strings.ToLower(ParticionesMontadasList[i].Id), strings.ToLower(Idlogueado)) {
			return true
		}
	}
	return false
}

func (logout *Logout) IsParticionMontadaVacia(p ParticionMontada) bool {
	return !p.Sistema_archivos && p.Id == "" && p.Letra == "" && p.Numero == 0 && p.Path == "" && p.Type == "" && p.Name == ""
}

func RetornarConsolalogout() string {
	return consola_logout
}
