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
			consola_response += "COMANDO: mkdisk\n"
			//Obtener parametros y se almacenan en un map
			params := getParams(comando)
			//Pasa al análisis del MKDISK
			AnalizarMkdisk(params)
		case "rmdisk":
			consola_response += "COMANDO: rmdisk\n"
			params := getParams(comando)
			AnalizarRmdisk(params)
		case "fdisk":
			consola_response += "COMANDO: fdisk\n"
			params := getParams(comando)
			AnalizarFdisk(params)
		case "mount":
			consola_response += "COMANDO: mount\n"
			params := getParams(comando)
			AnalizarMount(params)
		case "mkfs":
			consola_response += "COMANDO: mkfs\n"
			params := getParams(comando)
			AnalizarMkfs(params)
		case "login":
			consola_response += "COMANDO: login\n"
		case "logout":
			consola_response += "COMANDO: logout\n"
		case "mkgrp":
			consola_response += "COMANDO: mkgrp\n"
		case "rmgrp":
			consola_response += "COMANDO: rmgrp\n"
		case "mkusr":
			consola_response += "COMANDO: mkusr\n"
		case "rmusr":
			consola_response += "COMANDO: rmusr\n"
		case "mkfile":
			consola_response += "COMANDO: mkfile\n"
		case "mkdir":
			consola_response += "COMANDO: mkdir\n"
		case "rep":
			consola_response += "COMANDO: rep\n"
			params := getParams(comando)
			AnalizarRep(params)
		case "pause":
			consola_response += "COMANDO: pause\n"
		default:
			consola_response += "[-ERROR-] Comando no reconocido\n"
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
		tipo_params[1] = strings.TrimSpace(tipo_params[1])
		parametros[tipo_params[0]] = tipo_params[1]
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
			fmt.Println("Parametro no reconocido")
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
			fmt.Println("Parametro no reconocido")
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
			fmt.Println("Parametro no reconocido")
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
			fmt.Println("Parametro no reconocido")
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
			fmt.Println("Parametro no reconocido")
		}
	}
	mkfs.VerificarParams(params)
	consola_response += RetornarConsolamkfs()
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
