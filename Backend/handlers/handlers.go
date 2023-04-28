package handlers

import (
	"analizador/analizador"
	"encoding/json"
	"fmt"
	"net/http"
)

type Comando struct {
	Comando string `json:"comando"`
}

type Respuesta struct {
	Respuesta string `json:"respuesta"`
}

type ReportesResponse struct {
	Id   int    `json:"id"`
	Path string `json:"path"`
	Type string `json:"type"`
	Dot  string `json:"dot"`
}

type Login_Reportes struct {
	Id_particion string `json:"id_particion"`
	Usuario      string `json:"usuario"`
	Password     string `json:"password"`
}

type Respuesta2 struct {
	Autenticado bool `json:"autenticado"`
}

type mio struct {
	Carnet int    `json:"carnet"`
	Nombre string `json:"nombre"`
}

func GetAPI(rw http.ResponseWriter, r *http.Request) {
	//rw.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(rw, "API Funcionando correctamente")
}

func AnalizarComandos(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	//Obtener registro
	consola := Comando{}

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&consola); err != nil {
		fmt.Fprintln(rw, http.StatusUnprocessableEntity)
	} else {

		//Analizar comando
		analizador.Analizador_Comandos(consola.Comando)
		fmt.Println("Ya termino xd")
		fmt.Println("Respuesta:")
		fmt.Println(analizador.Devolver_consola())
		//Respuesta
		res := Respuesta{Respuesta: analizador.Devolver_consola()}
		output, _ := json.Marshal(res)
		fmt.Fprintln(rw, string(output))
	}

}

func Login(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	//Obtener registro
	autenticar := Login_Reportes{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&autenticar); err != nil {
		fmt.Fprintln(rw, http.StatusUnprocessableEntity)
	} else {
		loguear := analizador.LoginR{Id: autenticar.Id_particion, User: autenticar.Usuario, Pwd: autenticar.Password}
		valido := loguear.IniciarSesion()
		fmt.Println("Verificando")
		fmt.Println("Respuesta:")
		fmt.Println(valido)
		//Respuesta
		res := Respuesta2{Autenticado: valido}
		output, _ := json.Marshal(res)
		fmt.Fprintln(rw, string(output))
	}
}

func Reportes(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	//Obtener registro
	retorno := []ReportesResponse{}
	for i := 0; i < len(analizador.Reportes); i++ {
		retorno = append(retorno, ReportesResponse{Id: i + 1, Path: analizador.Reportes[i].Path, Type: analizador.Reportes[i].Type, Dot: analizador.Reportes[i].Dot})
	}

	output, _ := json.Marshal(retorno)
	fmt.Fprintln(rw, string(output))
}

func GetEstudiante(rw http.ResponseWriter, r *http.Request) {
	rodri := mio{Carnet: 201900042, Nombre: "Rodrigo Alejandro Hernández de León"}
	output, _ := json.Marshal(rodri)
	fmt.Fprintln(rw, string(output))
}
