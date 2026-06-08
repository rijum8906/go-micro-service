package db_test

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/google/uuid"
	coretest "github.com/rijum8906/relay/packages/core/testutils"
	taskdb "github.com/rijum8906/relay/services/task-service/internal/db"
)

var (
	testPool *pgxpool.Pool
	poolOnce sync.Once
)

func TestMain(m *testing.M) {
	poolOnce.Do(func() {
		testPool = coretest.MustConnectDB(
			coretest.WithDBName(coretest.GetTestDBName("task-service")),
		)
	})

	code := m.Run()

	if testPool != nil {
		testPool.Close()
	}

	os.Exit(code)
}

func newTestQueries(t *testing.T) (context.Context, *taskdb.Queries) {
	t.Helper()

	ctx := context.Background()

	tx, err := testPool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	t.Cleanup(func() {
		_ = tx.Rollback(ctx)
	})

	return ctx, taskdb.New(tx)
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: id,
		Valid: true,
	}
}
