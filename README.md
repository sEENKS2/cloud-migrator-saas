# Cloud Migrator SaaS Engine 🚀

Un motor de Extracción, Transformación y Carga (ETL) concurrente diseñado en **Go (Golang)** y conectado a un Dashboard en **React**, capaz de migrar volúmenes masivos de datos corporativos a la nube con consumo de memoria controlado y tolerancia a fallos en tiempo real.

---

## 📊 Métricas de Rendimiento Clave

* **Volumen de datos:** 500,000 registros persistidos en base de datos real.
* **Tiempo total del proceso:** ~2.01 segundos (Frente a los 13.3 segundos del modelo sincrónico).
* **Optimización del rendimiento:** **85% de reducción** en tiempos de procesamiento.
* **Consumo de RAM Máximo:** Estabilizado entre **6 MB y 8 MB** continuos, gracias a una arquitectura basada íntegramente en *Streaming Parsers*.

---

## 🛠️ Decisiones de Arquitectura y Complejidad Técnica

### 1. Procesamiento Eficiente de Memoria ($O(1)$ RAM)
En lugar de cargar el set de datos masivo por completo en la memoria del servidor (causando errores *Out Of Memory*), el backend implementa `json.Decoder` nativo de Go. Esto actúa como un cursor/stream que procesa el JSON token por token, manteniendo una huella de memoria constante e independiente del tamaño del archivo original.

### 2. Concurrencia de Alto Rendimiento (Patrón Productor-Consumidor)
Para eliminar el cuello de botella sincrónico de las escrituras en disco, se rediseñó el flujo principal utilizando **Goroutines** y **Channels** de Go:
* **1 Productor:** Realiza el streaming del archivo a velocidad de lectura de disco e inyecta elementos en un canal compartido buffered.
* **4 Workers en Paralelo (Consumidores):** Consumen del canal compartiendo la carga de forma balanceada y realizando escrituras masivas (*Bulk Inserts*) concurrentes hacia la base de datos.

### 3. Transacciones Atómicas y Tolerancia a Fallos (DLQ local)
Las escrituras se agrupan en lotes (*chunks*) de 5,000 registros usando `pgx.Batch` envueltos en **Transacciones SQL (`Begin/Commit/Rollback`)**. Si una fila de un lote está corrupta o viola restricciones de tipo de datos:
1.  Se aplica un `ROLLBACK` total e inmediato para ese bloque, garantizando la integridad de la base de datos.
2.  El motor captura la excepción, desvía el reporte detallado del error nativo de PostgreSQL hacia una cola de descarte (*Dead Letter Queue* simulada en un archivo de logs), y continúa procesando el siguiente lote sin interrumpir el SaaS.

### 4. Telemetría y Dashboard en Tiempo Real
El servidor expone estados protegidos contra condiciones de carrera (*Race Conditions*) utilizando un cerrojo de exclusión mutua (**`sync.Mutex`**). El cliente en React realiza consultas periódicas (*Polling de alta frecuencia*) consumiendo métricas precisas (porcentaje exacto, cantidad de procesados y lotes fallidos) mientras el motor asíncrono vuela en background.

---

## 🚀 Tecnologías Utilizadas

* **Backend:** Go (Golang 1.22+), `net/http` nativo, `pgx/v5` (Driver nativo de alto rendimiento).
* **Base de Datos:** PostgreSQL 15 ejecutado bajo contenedores **Docker / Docker Compose** en entornos Linux.
* **Frontend:** React, Vite, Tailwind CSS v4.

---

## 💻 Instrucciones para Levantar el Entorno Local

### Requisitos previos
Tener instalado Linux, Go, Node.js y Docker.

### 1. Levantar la Base de Datos
```bash
cd backend
docker compose up -d

### 2. Generar el set de datos masivos

go run cmd/generator/main.go

### 