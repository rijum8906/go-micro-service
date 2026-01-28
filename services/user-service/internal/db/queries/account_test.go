package querie_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rijum8906/go-micro-service/packages/common/database/postgres"
	"github.com/rijum8906/go-micro-service/packages/common/env"
	db "github.com/rijum8906/go-micro-service/services/user-service/internal/db/generated"
)

var (
	pgxPool *pgxpool.Pool
	queries *db.Queries
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// 2. Load environment (ensure you're hitting your local Arch test_db)
	cfg, err := env.Load()
	if err != nil {
		panic("cannot load config: " + err.Error())
	}

	// 3. Setup: One-time connection for the whole package
	pool := postgres.Connect(ctx, postgres.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		Database: "postgres", // Hardcode or env-gate the test database name!
	})

	// Assign to global variable
	pgxPool = pool
	queries = db.New(pgxPool)
	// 4. Run the tests
	code := m.Run()

	// 5. Cleanup (optional)
	pool.Close()

	os.Exit(code)
}

func TestCreateAccount(t *testing.T) {
	accountParams := db.CreateAccountParams{
		Email:        "test",
		PasswordHash: "test",
	}
	ctx := context.Background()

	_, err := queries.CreateAccount(ctx, accountParams)
	if err != nil {
		t.Fatal(err)
	}
	// TODO : Write test
}

func GetAccount(t *testing.T) {
	// TODO : Write test
}

func GetAccountByEmail(t *testing.T) {
	// TODO : Write test
}

func UpdateAccount(t *testing.T) {
	// TODO : Write test
}

func DeleteAccount(t *testing.T) {
	// TODO : Write test
}
