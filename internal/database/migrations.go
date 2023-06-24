package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
)

func GetDatabaseConnection(dbUrl string) (*sql.DB, error) {
	return sql.Open("postgres", dbUrl)
}

func (db *DB) Migrate() (int, error) {
	pkgRoot, err := findPackageRoot()
	if err != nil {
		return 0, errors.Wrap(err, "failed to find package root")
	}

	migrations := &migrate.FileMigrationSource{
		Dir: pkgRoot + "/migrations",
	}

	sql, err := db.DB.DB()
	if err != nil {
		return 0, errors.Wrap(err, "fail to get sql.DB")
	}

	n, err := migrate.Exec(sql, "postgres", migrations, migrate.Up)

	return n, errors.Wrap(err, "failed to apply migrations")
}

func findPackageRoot() (string, error) {
	// Get the path to the current file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("failed to get path to current file")
	}

	// Traverse up the directory tree until we find a directory that contains a go.mod file
	dir := filepath.Dir(filename)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("failed to find package root directory")
		}

		dir = parent
	}

	return dir, nil
}
