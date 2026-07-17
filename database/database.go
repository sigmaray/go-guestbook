package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"go-guestbook/config"

	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	gooseLock "github.com/pressly/goose/v3/lock"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// LoadEnv loads variables from a .env file when present.
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Debug().Err(err).Msg("no .env file loaded")
	}
}

// Connect opens a GORM connection to PostgreSQL using the provided configuration.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}
	return db, nil
}

// RunMigrations applies embedded Goose migrations to the configured database.
func RunMigrations(cfg *config.Config, migrations embed.FS) error {
	db, err := Connect(cfg)
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get database handle: %w", err)
	}
	defer sqlDB.Close()

	return applyMigrations(migrations, sqlDB)
}

// applyMigrations runs Goose migrations using a PostgreSQL session locker.
func applyMigrations(migrations embed.FS, sqlDB *sql.DB) error {
	migrationFS, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("open migrations directory: %w", err)
	}

	sessionLocker, err := gooseLock.NewPostgresSessionLocker()
	if err != nil {
		return fmt.Errorf("create migration session locker: %w", err)
	}

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		sqlDB,
		migrationFS,
		goose.WithSessionLocker(sessionLocker),
	)
	if err != nil {
		return fmt.Errorf("create migration provider: %w", err)
	}

	if _, err := provider.Up(context.Background()); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}

// ListTables returns user table names in the public schema.
// db is the open GORM connection used to query PostgreSQL catalog tables.
func ListTables(db *gorm.DB) ([]string, error) {
	var tables []string
	err := db.Raw(`
		SELECT tablename FROM pg_tables
		WHERE schemaname = 'public'
		ORDER BY tablename
	`).Scan(&tables).Error
	return tables, err
}

// ClearTable truncates the named table and restarts its identity columns.
// db is the open GORM connection; table is the PostgreSQL table name to clear.
func ClearTable(db *gorm.DB, table string) error {
	safeTable, err := sanitizeIdentifier(table)
	if err != nil {
		return err
	}
	return db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", safeTable)).Error
}

// ExecuteSQL runs arbitrary SQL and returns result rows for SELECT or WITH queries.
// db is the open GORM connection; query is the SQL statement to execute.
// For SELECT/WITH it returns columns and string rows; otherwise it returns rowsAffected.
func ExecuteSQL(db *gorm.DB, query string) (columns []string, rows [][]string, rowsAffected int64, err error) {
	trimmed := strings.TrimSpace(strings.ToUpper(query))
	if strings.HasPrefix(trimmed, "SELECT") || strings.HasPrefix(trimmed, "WITH") {
		sqlRows, qerr := db.Raw(query).Rows()
		if qerr != nil {
			return nil, nil, 0, qerr
		}
		defer func() { _ = sqlRows.Close() }()

		columns, err = sqlRows.Columns()
		if err != nil {
			return nil, nil, 0, err
		}

		for sqlRows.Next() {
			values := make([]interface{}, len(columns))
			ptrs := make([]interface{}, len(columns))
			for i := range values {
				ptrs[i] = &values[i]
			}
			if err := sqlRows.Scan(ptrs...); err != nil {
				return nil, nil, 0, err
			}
			row := make([]string, len(columns))
			for i, v := range values {
				if v == nil {
					row[i] = "NULL"
				} else {
					row[i] = fmt.Sprint(v)
				}
			}
			rows = append(rows, row)
		}
		return columns, rows, int64(len(rows)), sqlRows.Err()
	}

	result := db.Exec(query)
	return nil, nil, result.RowsAffected, result.Error
}

// sanitizeIdentifier validates a SQL identifier so it can be interpolated safely.
// name is the candidate table or identifier string from caller input.
func sanitizeIdentifier(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("empty table name")
	}
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			continue
		}
		return "", fmt.Errorf("invalid table name: %s", name)
	}
	return name, nil
}
