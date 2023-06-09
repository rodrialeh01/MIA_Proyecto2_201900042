package analizador

import (
	"fmt"
	"strconv"
	"strings"
)

var consola_response string

func Analizador_Comandos(entrada string) {
	//Limpia el array de reportes
	Reportes = Reportes[:0]
	consola_response = ""
	lista_comandos := strings.Split(entrada, "\n")
	for i := 0; i < len(lista_comandos); i++ {
		lista_comandos[i] = strings.TrimSpace(lista_comandos[i])
		lista_comandos[i] = strings.Replace(lista_comandos[i], "\t", "", -1)
		lista_comandos[i] = strings.Replace(lista_comandos[i], "\r", "", -1)
		if lista_comandos[i] != "" {
			Analizar_Comando(lista_comandos[i])
		}
	}
}

func Analizar_Comando(comando string) {
	//Verificar si es comentario
	if comando[0] == '#' {
		consola_response += comando + "\n"
	} else {
		//Verifica los demas comandos
		verificar_comando := strings.Split(comando, " ")
		verificador := strings.ToLower(verificar_comando[0])
		switch verificador {
		case "mkdisk":
			consola_response += "\n\nEJECUTANDO: mkdisk\n"
			//Obtener parametros y se almacenan en un map
			params := getParams(comando)
			//Pasa al análisis del MKDISK
			AnalizarMkdisk(params)
		case "rmdisk":
			consola_response += "\n\nEJECUTANDO: rmdisk\n"
			params := getParams(comando)
			AnalizarRmdisk(params)
		case "fdisk":
			consola_response += "\n\nEJECUTANDO: fdisk\n"
			params := getParams(comando)
			AnalizarFdisk(params)
		case "mount":
			consola_response += "\n\nEJECUTANDO: mount\n"
			params := getParams(comando)
			AnalizarMount(params)
		case "mkfs":
			consola_response += "\n\nEJECUTANDO: mkfs\n"
			params := getParams(comando)
			AnalizarMkfs(params)
		case "login":
			consola_response += "\n\nEJECUTANDO: login\n"
			params := getParams(comando)
			AnalizarLogin(params)
		case "logout":
			consola_response += "\n\nEJECUTANDO: logout\n"
			AnalizarLogout()
		case "mkgrp":
			consola_response += "\n\nEJECUTANDO: mkgrp\n"
			params := getParams(comando)
			AnalizarMkgrp(params)
		case "rmgrp":
			consola_response += "\n\nEJECUTANDO: rmgrp\n"
			params := getParams(comando)
			AnalizarRmgrp(params)
		case "mkusr":
			consola_response += "\n\nEJECUTANDO: mkusr\n"
			params := getParams(comando)
			AnalizarMkuser(params)
		case "rmusr":
			consola_response += "\n\nEJECUTANDO: rmusr\n"
			params := getParams(comando)
			AnalizarRmusr(params)
		case "mkfile":
			consola_response += "\n\nEJECUTANDO: mkfile\n"
			params := getParams(comando)
			fmt.Println(params)
			AnalizarMkfile(params)
		case "mkdir":
			consola_response += "\n\nEJECUTANDO: mkdir\n"
			params := getParams(comando)
			AnalizarMkdir(params)
		case "rep":
			consola_response += "\n\nEJECUTANDO: rep\n"
			params := getParams(comando)
			AnalizarRep(params)
		case "pause":
			consola_response += "Se detectó el comando Pause\n\n"
		default:
			consola_response += "[-ERROR-] Comando no reconocido\n\n"
		}
	}
}

func getParams(comando string) map[string]string {
	lista_params := strings.Split(comando, ">")
	parametros := make(map[string]string)
	for i := 1; i < len(lista_params); i++ {
		lista_params[i] = strings.TrimSpace(lista_params[i])
		tipo_params := strings.Split(lista_params[i], "=")
		tipo_params[0] = strings.TrimSpace(tipo_params[0])
		tipo_params[0] = strings.ToLower(tipo_params[0])
		if strings.TrimSpace(strings.ToLower(lista_params[i])) != "r" {
			tipo_params[1] = strings.TrimSpace(tipo_params[1])
			parametros[tipo_params[0]] = tipo_params[1]
		} else if strings.TrimSpace(strings.ToLower(lista_params[i])) == "r" {
			parametros["r"] = "r"
		}
	}
	for key, value := range parametros {
		fmt.Println(key, ":", value)
	}
	return parametros
}

func AnalizarMkdisk(params map[string]string) {
	var mkdisk MkDisk
	for key, value := range params {
		switch key {
		case "size":
			s, err := strconv.Atoi(value)
			if err != nil {
				fmt.Println("Error al convertir el valor")
			}
			mkdisk.Size = s
			fmt.Println(value)
		case "unit":
			mkdisk.Unit = value
			fmt.Println(value)
		case "path":
			mkdisk.Path = value
			fmt.Println(value)
		case "fit":
			mkdisk.Fit = value
			fmt.Println(value)
		default:
			consola_response += "[-ERROR-] Parametro no reconocido"
			return
		}
	}
	mkdisk.VerificarParams(params)
	consola_response += RetornarConsolamkdisk()
}

func AnalizarRmdisk(params map[string]string) {
	var rmdisk Rmdisk
	for key, value := range params {
		switch key {
		case "path":
			rmdisk.Path = value
			fmt.Println(value)
		default:
			consola_response += "[-ERROR-] Parametro no reconocido"
			return
		}
	}
	rmdisk.VerificarParams(params)
	consola_response += RetornarConsolarmdisk()
}

func AnalizarFdisk(params map[string]string) {
	var fdisk Fdisk
	for key, value := range params {
		switch key {
		case "size":
			s, err := strconv.Atoi(value)
			if err != nil {
				fmt.Println("Error al convertir el valor")
			}
			fdisk.Size = s
			fmt.Println(value)
		case "unit":
			fdisk.Unit = value
			fmt.Println(value)
		case "path":
			fdisk.Path = value
			fmt.Println(value)
		case "type":
			fdisk.Type = value
			fmt.Println(value)
		case "fit":
			fdisk.Fit = value
			fmt.Println(value)
		case "name":
			fdisk.Name = value
			fmt.Println(value)
		default:
			consola_response += "[-ERROR-] Parametro no reconocido"
			return
		}
	}
	fdisk.VerificarParams(params)
	consola_response += RetornarConsolafdisk()
}

func AnalizarMount(params map[string]string) {
	var mount Mount
	for key, value := range params {
		switch key {
		case "path":
			mount.Path = value
			fmt.Println(value)
		case "name":
			mount.Name = value
			fmt.Println(value)
		default:
			consola_response += "[-ERROR-] Parametro no reconocido"
			return
		}
	}
	mount.VerificarParams(params)
	consola_response += RetornarConsolamount()
}

func AnalizarMkfs(params map[string]string) {
	var mkfs Mkfs
	for key, value := range params {
		switch key {
		case "id":
			mkfs.Id = value
			fmt.Println(value)
		case "type":
			mkfs.Type = value
			fmt.Println(value)
		default:
			consola_response += "[-ERROR-] Parametro no reconocido"
			return
		}
	}
	mkfs.VerificarParams(params)
	consola_response += RetornarConsolamkfs()
}

func AnalizarLogin(params map[string]string) {
	var login Login
	for key, value := range params {
		switch key {
		case "user":
			login.User = value
			fmt.Println(value)
		case "pwd":
			login.Pwd = value
			fmt.Println(value)
		case "id":
			login.Id = value
			fmt.Println(value)
		default:
			consola_response += "[-ERROR-] Parametro no reconocido"
			return
		}
	}
	login.VerificarParams(params)
	consola_response += RetornarConsolalogin()
}

func AnalizarLogout() {
	var logout Logout
	logout.VerificarParams()
	consola_response += RetornarConsolalogout()
}

func AnalizarMkgrp(params map[string]string) {
	var mkgrp Mkgrp
	for key, value := range params {
		switch key {
		case "name":
			mkgrp.Name = value
			fmt.Println(value)
		default:
			consola_response += "[-ERROR-] Parametro no reconocido"
			return
		}
	}
	mkgrp.VerificarParams(params)
	consola_response += RetornarConsolamkgrp()
}

func AnalizarRmgrp(params map[string]string) {
	var rmgrp Rmgrp
	for key, value := range params {
		switch key {
		case "name":
			rmgrp.Name = value
			fmt.Println(value)
		default:
			consola_response += "[-ERROR-] Parametro no reconocido"
			return
		}
	}
	rmgrp.VerificarParams(params)
	consola_response += RetornarConsolarmgrp()
}

func AnalizarMkuser(params map[string]string) {
	var mkuser Mkuser
	for key, value := range params {
		switch key {
		case "user":
			mkuser.User = value
			fmt.Println(value)
		case "pwd":
			mkuser.Pwd = value
			fmt.Println(value)
		case "grp":
			mkuser.Grp = value
			fmt.Println(value)
		default:
			consola_response += "[-ERROR-] Parametro no reconocido"
			return
		}
	}
	mkuser.VerificarParams(params)
	consola_response += RetornarConsolamkuser()
}

func AnalizarRmusr(params map[string]string) {
	var rmusr Rmusr
	for key, value := range params {
		switch key {
		case "user":
			rmusr.User = value
			fmt.Println(value)
		default:
			consola_response += "[-ERROR-] Parametro no reconocido"
			return
		}
	}
	rmusr.VerificarParams(params)
	consola_response += RetornarConsolarmusr()
}

func AnalizarMkdir(params map[string]string) {
	var mkdir Mkdir
	for key, value := range params {
		switch key {
		case "path":
			mkdir.Path = value
			fmt.Println(value)
		case "r":
			if value == "r" {
				mkdir.R = true
			} else {
				mkdir.R = false
			}
		default:
			consola_response += "[-ERROR-] Parametro no reconocido"
			return
		}
	}
	mkdir.VerificarParams(params)
	consola_response += RetornarConsolamkdir()
}

func AnalizarMkfile(params map[string]string) {
	var mkfile Mkfile
	for key, value := range params {
		switch key {
		case "path":
			mkfile.Path = value
			fmt.Println(value)
		case "size":
			s, err := strconv.Atoi(value)
			if err != nil {
				fmt.Println("Error al convertir el valor")
			}
			mkfile.Size = s
			fmt.Println(value)
		case "cont":
			mkfile.Cont = value
			fmt.Println(value)
		case "r":
			if value == "r" {
				mkfile.R = true
			} else {
				mkfile.R = false
			}
		default:
			consola_response += "[-ERROR-] Parametro no reconocido"
			return
		}
	}
	mkfile.VerificarParams(params)
	consola_response += RetornarConsolamkfile()
}

func AnalizarRep(params map[string]string) {
	var rep Rep
	for key, value := range params {
		switch key {
		case "name":
			rep.Name = value
			fmt.Println(value)
		case "path":
			rep.Path = value
			fmt.Println(value)
		case "id":
			rep.Id = value
			fmt.Println(value)
		case "ruta":
			rep.Ruta = value
		default:
			fmt.Println("Parametro no reconocido")
		}
	}
	rep.VerificarParams(params)
	consola_response += RetornarConsolarep()
}

func Devolver_consola() string {
	return consola_response
}
