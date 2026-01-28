package querie_test

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rijum8906/go-micro-service/packages/common/database/postgres"
	"github.com/rijum8906/go-micro-service/packages/common/env"
	db "github.com/rijum8906/go-micro-service/services/user-service/internal/db/generated"
)

var (
	pgxPool *pgxpool.Pool
	queries *db.Queries
	email   string
	pass    string
	id      pgtype.UUID
)

func FormatUUID(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	// Formats the 16-byte array into the standard UUID string format
	return fmt.Sprintf("%x-%x-%x-%x-%x", u.Bytes[0:4], u.Bytes[4:6], u.Bytes[6:8], u.Bytes[8:10], u.Bytes[10:16])
}

func GenerateTestCredentials() (string, string) {
	// Generate a random ID for the email
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	email := fmt.Sprintf("test-%d@gmail.com", n.Int64())

	// Generate a secure random password (16 characters)
	b := make([]byte, 8) // 8 bytes = 16 hex characters
	if _, err := rand.Read(b); err != nil {
		return email, "defaultPassword123" // Fallback
	}
	password := hex.EncodeToString(b)

	return email, password
}

func IsDuplicateEntryError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" // Unique Violation
	}
	return false
}

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
	email, pass = GenerateTestCredentials()
	// 4. Run the tests
	code := m.Run()

	// 5. Cleanup (optional)
	pool.Close()

	os.Exit(code)
}

func TestAccountWorkflow(t *testing.T) {
	// 1. Create the Account (sets the global 'id')
	t.Run("CreateAccount", testCreateAccount)

	// 2. Get the Account by ID
	t.Run("GetAccount", testGetAccount)

	// 3. Get the Account by Email
	t.Run("GetAccountByEmail", testGetAccountByEmail)

	// 4. Update and finally Delete
	t.Run("UpdateAccount", testUpdateAccount)
	t.Run("DeleteAccount", testDeleteAccount)
}

func testCreateAccount(t *testing.T) {
	accountParams := db.CreateAccountParams{
		Email:        email,
		PasswordHash: pass,
	}
	ctx := context.Background()

	account, err := queries.CreateAccount(ctx, accountParams)
	if err != nil {
		t.Fatal(err)
	}

	id = account.ID
	_, err = queries.CreateAccount(ctx, accountParams)
	if isDup := IsDuplicateEntryError(err); isDup != true {
		t.Fatal("this should return error for duplicate entry")
	}
	// TODO : Write test
}

func testGetAccount(t *testing.T) {
	ctx := context.Background()
	account, err := queries.GetAccount(ctx, id)
	if err != nil {
		t.Fatal(err)
	}

	if account.Email != email {
		t.Fatal("account email not matching")
	}

	if account.PasswordHash != pass {
		t.Fatal("account password not matching")
	}
	// TODO : Write test
}

func testGetAccountByEmail(t *testing.T) {
	ctx := context.Background()
	account, err := queries.GetAccountByEmail(ctx, email)
	if err != nil {
		t.Fatal(err)
	}

	if account.Email != email {
		t.Fatal("account email not matching")
	}

	if account.PasswordHash != pass {
		t.Fatal("account password not matching")
	}
	// TODO : Write test
}

func testUpdateAccount(t *testing.T) {
	ctx := context.Background()
	email2, pass2 := GenerateTestCredentials()
	updateParams := db.UpdateAccountParams{
		ID:           id,
		Email:        email2,
		PasswordHash: pass2,
	}
	account, err := queries.UpdateAccount(ctx, updateParams)
	if err != nil {
		t.Fatal(err)
	}
	if account.Email != email2 {
		t.Fatal("account email not matching")
	}
	if account.PasswordHash != pass2 {
		t.Fatal("account password not matching")
	}
	// TODO : Write test
}

func testDeleteAccount(t *testing.T) {
	ctx := context.Background()
	err := queries.DeleteAccount(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	_, err = queries.GetAccount(ctx, id)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("account not deleted")
	}
	// TODO : Write test
}
