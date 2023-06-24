package testutils

import (
	"database/sql"
	"fmt"
	"math/rand"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/ergomake/ergomake/internal/database"
)

func CreateRandomDB(t *testing.T) *database.DB {
	oldConnStr := "postgres://ergomake:ergomake@localhost/ergomake?sslmode=disable"

	sqlDB, err := sql.Open("postgres", oldConnStr)
	require.NoError(t, err)
	defer sqlDB.Close()

	rand.Seed(time.Now().UnixNano())
	dbName := fmt.Sprintf("ergomake_testdb_%d", rand.Intn(10000))

	_, err = sqlDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	require.NoError(t, err)

	err = sqlDB.Close()
	require.NoError(t, err)

	t.Cleanup(func() {
		db, err := sql.Open("postgres", oldConnStr)
		require.NoError(t, err)
		defer db.Close()

		_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
		require.NoError(t, err)
	})

	connectionString := fmt.Sprintf("postgres://ergomake:ergomake@localhost/%s?sslmode=disable", dbName)
	db, err := database.Connect(connectionString)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := db.Close()
		require.NoError(t, err)
	})

	_, err = db.Migrate()
	require.NoError(t, err)

	return db
}
