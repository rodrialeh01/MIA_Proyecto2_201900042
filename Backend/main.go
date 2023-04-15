package main

import (
	myhandlers "analizador/handlers"
	"fmt"
	"log"
	"net/http"

	gorillahandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	//Rutas
	mux := mux.NewRouter()

	//Endpoints
	mux.HandleFunc("/api", myhandlers.GetAPI).Methods("GET")
	mux.HandleFunc("/api/consola", myhandlers.AnalizarComandos).Methods("POST")
	mux.HandleFunc("/api/login", myhandlers.Login).Methods("POST")
	mux.HandleFunc("/api/reportes", myhandlers.Reportes).Methods("GET")
	mux.HandleFunc("/estudiante/201900042", myhandlers.GetEstudiante).Methods("GET")

	// Habilitar CORS para todas las solicitudes
	headers := gorillahandlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methods := gorillahandlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"})
	origins := gorillahandlers.AllowedOrigins([]string{"*"})

	//Servidor
	fmt.Println("Servidor corriendo en el puerto 3000")
	fmt.Println("http://localhost:3000/api")
	log.Fatal(http.ListenAndServe(":3000", gorillahandlers.CORS(headers, methods, origins)(mux)))
}
