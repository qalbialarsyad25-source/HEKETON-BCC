package suite

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"main/internal/app"
	"main/internal/config"
	"main/internal/infrastructure"
	"main/internal/shared/ports"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// ownedTables are emptied before every test. Add each new table here.
var ownedTables = []string{
	"posts",
	"refresh_tokens",
	"announcements",
	"invitations",
	"registrations",
	"committees",
	"sessions",
	"events",
	"users",
}

// lockKey names the Postgres advisory lock that keeps e2e tests from running at the same time.
// `go test ./tests/e2e/...` runs the packages in parallel, and they all share one database.
const lockKey int64 = 7_301_001

type Suite struct {
	Router   *gin.Engine
	Adapters *ports.Adapters
	DB       *gorm.DB
	Config   *config.Config
}

type E2ETest[T any] struct {
	Name               string
	Data               T
	ExpectedStatusCode int
	ExpectedResponse   any
}

// New starts the application against the services in .env, with empty tables.
// Everything is closed again when the test ends.
func New(t *testing.T) *Suite {
	t.Helper()
	ctx := context.Background()

	cfg, err := config.Load()
	require.NoError(t, err, "load config")

	if cfg.App.Mode == config.ModeProduction {
		t.Fatal("refusing to run e2e tests with APP_MODE=PRODUCTION: they truncate the database")
	}

	// Quiet Gin output, and keep test logs out of the repo's logs/ folder.
	cfg.App.Mode = config.ModeTest
	cfg.Logger.FilePath = filepath.Join(t.TempDir(), "app.log")

	adapters, db, err := infrastructure.Initialize(ctx, cfg)
	require.NoError(t, err, "initialize adapters")
	t.Cleanup(func() {
		_ = infrastructure.Shutdown(adapters, db)
		_ = infrastructure.CloseLogger(adapters.Logger)
	})

	lockDatabase(t, db)
	truncateTables(t, db)

	return &Suite{
		Router:   app.SetupRouter(cfg, db),
		Adapters: adapters,
		DB:       db,
		Config:   cfg,
	}
}

// lockDatabase waits until no other e2e test holds the lock, and holds it until this test ends.
func lockDatabase(t *testing.T, db *gorm.DB) {
	t.Helper()
	ctx := context.Background()

	sqlDB, err := db.DB()
	require.NoError(t, err)

	// A session lock belongs to one connection, so keep a dedicated one.
	conn, err := sqlDB.Conn(ctx)
	require.NoError(t, err)

	_, err = conn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", lockKey)
	require.NoError(t, err, "take the e2e database lock")

	t.Cleanup(func() {
		_, _ = conn.ExecContext(context.Background(), "SELECT pg_advisory_unlock($1)", lockKey)
		_ = conn.Close()
	})
}

func truncateTables(t *testing.T, db *gorm.DB) {
	t.Helper()

	err := db.Exec("TRUNCATE TABLE " + strings.Join(ownedTables, ", ") + " RESTART IDENTITY CASCADE").Error
	require.NoError(t, err, "truncate tables")
}
