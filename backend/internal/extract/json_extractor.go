package extract

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Usuario mapea la estructura del JSON (debe coincidir con el generador)
type Usuario struct {
	ID        int       `json:"id"`
	Nombre    string    `json:"nombre"`
	Email     string    `json:"email"`
	Empresa   string    `json:"empresa"`
	CreadoEn  time.Time `json:"creado_en"`
}

// ProcesarArchivoJSON lee el archivo de forma eficiente usando streaming
func ProcesarArchivoJSON(rutaArchivo string, callback func(Usuario) error) error {
	archivo, err := os.Open(rutaArchivo)
	if err != nil {
		return fmt.Errorf("no se pudo abrir el archivo: %w", err)
	}
	defer archivo.Close()

	decoder := json.NewDecoder(archivo)

	// Leer el corchete de apertura del array '['
	_, err = decoder.Token()
	if err != nil {
		return fmt.Errorf("error al leer inicio del JSON: %w", err)
	}

	contador := 0
	// decoder.More() verifica si hay más elementos dentro del array
	for decoder.More() {
		var u Usuario
		// Decodifica únicamente el objeto JSON actual en la variable 'u'
		err := decoder.Decode(&u)
		if err != nil {
			return fmt.Errorf("error al decodificar registro en la posicion %d: %w", contador, err)
		}

		// Enviamos el usuario al callback para hacer "algo" con él (en el futuro, enviarlo a la BD)
		err = callback(u)
		if err != nil {
			return fmt.Errorf("error en el procesamiento del registro: %w", err)
		}

		contador++
	}

	fmt.Printf("\n[Extractor] Stream finalizado con éxito. Registros procesados: %d\n", contador)
	return nil
}