package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Usuario representa la estructura de datos que vamos a migrar
type Usuario struct {
	ID        int       `json:"id"`
	Nombre    string    `json:"nombre"`
	Email     string    `json:"email"`
	Empresa   string    `json:"empresa"`
	CreadoEn  time.Time `json:"creado_en"`
}

func main() {
	ContadorRegistros := 500000
	NombreArchivo := "usuarios_masivos.json"

	archivo, err := os.Create(NombreArchivo)
	if err != nil {
		fmt.Printf("Error al crear el archivo: %v\n", err)
		return
	}
	defer archivo.Close()

	// Usamos un buffer de escritura para que sea veloz
	writer := bufio.NewWriter(archivo)
	defer writer.Flush()

	fmt.Printf("Generando %d registros en %s...\n", ContadorRegistros, NombreArchivo)
	start := time.Now()

	// Abrimos el array JSON
	writer.WriteString("[\n")

	for i := 1; i <= ContadorRegistros; i++ {
		u := Usuario{
			ID:        i,
			Nombre:    fmt.Sprintf("Usuario %d", i),
			Email:     fmt.Sprintf("usuario%d@empresa.com", i),
			Empresa:   "Empresa Ficticia S.A.",
			CreadoEn:  time.Now().AddDate(0, 0, -i),
		}

		data, err := json.Marshal(u)
		if err != nil {
			fmt.Printf("Error al serializar: %v\n", err)
			return
		}

		writer.Write(data)

		// Si no es el último, agregamos una coma
		if i < ContadorRegistros {
			writer.WriteString(",\n")
		} else {
			writer.WriteString("\n")
		}
	}

	// Cerramos el array JSON
	writer.WriteString("]")

	fmt.Printf("¡Listo! Archivo generado en %v\n", time.Since(start))
}