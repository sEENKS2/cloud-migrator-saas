package load

import (
	"context"
	"fmt"

	"cloud-migrator/internal/extract"
	"github.com/jackc/pgx/v5"
)

// Loader gestiona la conexión a la base de datos de destino
type Loader struct {
	conn *pgx.Conn
}

// NuevoLoader inicializa la conexión a Postgres
func NuevoLoader(ctx context.Context, connString string) (*Loader, error) {
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("no se pudo conectar a la BD: %w", err)
	}
	return &Loader{conn: conn}, nil
}

// Cerrar cierra la conexión de la BD
func (l *Loader) Cerrar(ctx context.Context) {
	l.conn.Close(ctx)
}

// PrepararTabla genera la estructura limpia en el destino
func (l *Loader) PrepararTabla(ctx context.Context) error {
	query := `
	DROP TABLE IF EXISTS usuarios;
	CREATE TABLE usuarios (
		id INT PRIMARY KEY,
		nombre VARCHAR(100),
		email VARCHAR(150),
		empresa VARCHAR(100),
		creado_en TIMESTAMP
	);`
	_, err := l.conn.Exec(ctx, query)
	return err
}

// CargarEnLotes recibe un chunk de usuarios y los inserta usando un Batch de pgx
func (l *Loader) CargarEnLotes(ctx context.Context, usuarios []extract.Usuario) error {
	// Iniciamos la transacción para este lote
	tx, err := l.conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("no se pudo iniciar la transaccion: %w", err)
	}
	
	// Nos aseguramos de que si la función termina con error, se deshagan los cambios abortados
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}
	for _, u := range usuarios {
		query := `INSERT INTO usuarios (id, nombre, email, empresa, creado_en) VALUES ($1, $2, $3, $4, $5)`
		batch.Queue(query, u.ID, u.Nombre, u.Email, u.Empresa, u.CreadoEn)
	}

	// Enviamos el batch DENTRO de la transacción
	br := tx.SendBatch(ctx, batch)
	err = br.Close() // br.Close() evalúa errores generales del batch
	if err != nil {
		return fmt.Errorf("error al cerrar el batch de la transaccion: %w", err)
	}

	// Si todo el batch fue exitoso, consolidamos los cambios en el disco
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("error al hacer commit de la transaccion: %w", err)
	}

	return nil
}