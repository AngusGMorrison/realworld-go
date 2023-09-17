package postgres

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/angusgmorrison/realworld-go/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("connects to the DB and runs migrations", func(t *testing.T) {
		t.Parallel()

		cfg, err := config.New()
		require.NoError(t, err)

		client, err := New(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.db)
		assert.NotNil(t, client.queries)

		err = client.db.Ping()
		assert.NoError(t, err)

		expectedVersion := latestMigrationVersion(t)
		migrator, err := newMigrator(client.db)
		gotVersion, dirty, err := migrator.Version()
		require.NoError(t, err)
		assert.Equal(t, expectedVersion, gotVersion)
		assert.False(t, dirty, "Latest migration is dirty")

		_ = client.Close()
	})

	t.Run("returns an error if the DB can't be opened", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{}

		client, err := New(cfg)
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}

func latestMigrationVersion(t *testing.T) uint {
	t.Helper()

	migrations, err := os.ReadDir(migrationsPath)
	require.NoError(t, err)

	latestMigrationName := migrations[len(migrations)-1].Name()
	timestamp, _, found := strings.Cut(latestMigrationName, "_")
	require.True(t, found, "Failed to parse migration timestamp: no underscore found")

	version, err := strconv.ParseUint(timestamp, 10, 64)
	require.NoError(t, err)

	return uint(version)
}
