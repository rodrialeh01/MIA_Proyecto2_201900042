package handlers

import (
	"apirest/analizador"
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

		//Respuesta
		res := Respuesta{Respuesta: "Comando recibido"}
		output, _ := json.Marshal(res)
		fmt.Fprintln(rw, string(output))
	}

}

func Login(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(rw, "Login")
}

func Reportes(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(rw, "Reportes")
}
