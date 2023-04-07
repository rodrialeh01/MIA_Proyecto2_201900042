package main

import (
	"apirest/handlers"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	//Rutas
	mux := mux.NewRouter()

	//Endpoints
	mux.HandleFunc("/api", handlers.GetAPI).Methods("GET")
	mux.HandleFunc("/api/consola", handlers.AnalizarComandos).Methods("POST")
	mux.HandleFunc("/api/login", handlers.Login).Methods("POST")
	mux.HandleFunc("/api/reportes", handlers.Reportes).Methods("GET")

	//Servidor
	fmt.Println("Servidor corriendo en el puerto 3000")
	fmt.Println("http://localhost:3000/api")
	log.Fatal(http.ListenAndServe(":3000", mux))
}
