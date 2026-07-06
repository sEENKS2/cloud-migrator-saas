package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"cloud-migrator/internal/extract"
	"cloud-migrator/internal/load"
)

type RespuestaEstandar struct {
	Mensaje   string    `json:"mensaje"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type EstadoProgreso struct {
	TotalRegistros int     `json:"total_registros"`
	Procesados     int     `json:"procesados"`
	LotesFallidos  int     `json:"lotes_fallidos"`
	Porcentaje     float64 `json:"porcentaje"`
	Status         string  `json:"status"`
}

type Servidor struct {
	addr       string
	connString string
	mu         sync.Mutex
	progreso   EstadoProgreso
}

func NuevoServidor(addr, connString string) *Servidor {
	return &Servidor{
		addr:       addr,
		connString: connString,
		progreso: EstadoProgreso{
			Status: "IDLE",
		},
	}
}

func (s *Servidor) Iniciar() error {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/salud", s.handleSalud)
	mux.HandleFunc("POST /api/migrar", s.handleMigrar)
	mux.HandleFunc("GET /api/progreso", s.handleProgreso)
	mux.HandleFunc("GET /api/errores/descargar", s.handleDescargarErrores) // Endpoint registrado

	fmt.Printf("[API] Servidor corriendo en %s\n", s.addr)
	return http.ListenAndServe(s.addr, mux)
}

func (s *Servidor) handleSalud(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(RespuestaEstandar{
		Mensaje:   "SaaS en línea y motor ETL listo.",
		Status:    "OK",
		Timestamp: time.Now(),
	})
}

func (s *Servidor) handleProgreso(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	json.NewEncoder(w).Encode(s.progreso)
}

func (s *Servidor) handleMigrar(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	s.mu.Lock()
	if s.progreso.Status == "PROCESSING" {
		s.mu.Unlock()
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(RespuestaEstandar{
			Mensaje: "Ya hay una migración en curso.",
			Status:  "CONFLICT",
		})
		return
	}
	
	s.progreso = EstadoProgreso{
		TotalRegistros: 500000,
		Procesados:     0,
		LotesFallidos:  0,
		Porcentaje:     0.0,
		Status:         "PROCESSING",
	}
	s.mu.Unlock()

	w.WriteHeader(http.StatusAccepted)
	go s.ejecutarMigracion()

	json.NewEncoder(w).Encode(RespuestaEstandar{
		Mensaje:   "Migración masiva iniciación en segundo plano.",
		Status:    "PROCESSING",
		Timestamp: time.Now(),
	})
}

func (s *Servidor) handleDescargarErrores(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	nombreArchivo := "errores_migracion.log"

	if _, err := os.Stat(nombreArchivo); os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(RespuestaEstandar{
			Mensaje: "No se encontró ningún reporte de errores reciente.",
			Status:  "NOT_FOUND",
		})
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=errores_migracion.log")
	w.Header().Set("Content-Type", "text/plain")

	http.ServeFile(w, r, nombreArchivo)
}

func (s *Servidor) ejecutarMigracion() {
	ctx := context.Background()
	rutaArchivo := "usuarios_masivos.json"
	tamanoLote := 5000
	numWorkers := 4

	logErrores, _ := os.Create("errores_migracion.log")
	defer logErrores.Close()

	loader, err := load.NuevoLoader(ctx, s.connString)
	if err != nil {
		s.actualizarStatusFinal("FAILED")
		return
	}
	defer loader.Cerrar(ctx)

	_ = loader.PrepararTabla(ctx)

	canalUsuarios := make(chan extract.Usuario, 10000)
	var wg sync.WaitGroup

	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			var lote []extract.Usuario

			for u := range canalUsuarios {
				if u.ID == 250005 {
					u.Email = "este_email_es_demasiado_largo_para_la_columna_y_va_a_romper_el_batch..."
				}
				lote = append(lote, u)

				if len(lote) >= tamanoLote {
					err := loader.CargarEnLotes(ctx, lote)
					
					s.mu.Lock()
					s.progreso.Procesados += len(lote)
					if err != nil {
						s.progreso.LotesFallidos++
						msgError := fmt.Sprintf("Error lote ID %d: %v\n", u.ID, err)
						logErrores.WriteString(msgError)
					}
					s.progreso.Porcentaje = (float64(s.progreso.Procesados) / float64(s.progreso.TotalRegistros)) * 100
					s.mu.Unlock()

					lote = lote[:0]
				}
			}
			
			if len(lote) > 0 {
				err := loader.CargarEnLotes(ctx, lote)
				s.mu.Lock()
				s.progreso.Procesados += len(lote)
				if err != nil {
					s.progreso.LotesFallidos++
				}
				s.progreso.Porcentaje = (float64(s.progreso.Procesados) / float64(s.progreso.TotalRegistros)) * 100
				s.mu.Unlock()
			}
		}(i)
	}

	_ = extract.ProcesarArchivoJSON(rutaArchivo, func(u extract.Usuario) error {
		canalUsuarios <- u
		return nil
	})

	close(canalUsuarios)
	wg.Wait()

	s.actualizarStatusFinal("COMPLETED")
	fmt.Println("[Motor] ✅ Telemetría finalizada.")
}

func (s *Servidor) actualizarStatusFinal(status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.progreso.Status = status
	if status == "COMPLETED" {
		s.progreso.Porcentaje = 100.0
		s.progreso.Procesados = s.progreso.TotalRegistros
	}
}