package main

import (
	"cloud-migrator/internal/api"
	"log"
)

func main() {
	addr := ":8080"
	connString := "postgres://migrator_user:migrator_password@localhost:5432/cloud_migrator_db?sslmode=disable"

	// Inicializamos el servidor web del SaaS
	srv := api.NuevoServidor(addr, connString)

	// Encendemos el servidor (bloquea el hilo principal esperando peticiones)
	if err := srv.Iniciar(); err != nil {
		log.Fatalf("Error al arrancar el servidor: %v", err)
	}
}