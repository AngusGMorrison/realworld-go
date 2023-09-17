package sqlite

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDBBasename = "realworld-test.db"

var testDBPath = filepath.Join(os.TempDir(), testDBBasename)

func TestMain(m *testing.M) {
	// Create test DB if it doesn't exist and run migrations.
	db, err := New(testDBPath)
	if err != nil {
		log.Fatalf("initialize test DB: %v", err)
	}

	if err := db.Close(); err != nil {
		log.Fatalf("close test DB: %v", err)
	}

	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("creates a new DB and runs migrations if it doesn't exist", func(t *testing.T) {
		t.Parallel()

		db, err := New(filepath.Join(os.TempDir(), fmt.Sprintf("%s.db", uuid.New().String())))
		assert.NoError(t, err)
		assert.NotNil(t, db)
		assert.NotNil(t, db.innerDB)
		assert.NotNil(t, db.queries)

		migrator, err := newMigrator(db.innerDB)
		require.NoError(t, err)

		version, dirty, err := migrator.Version()
		require.NoError(t, err)
		assert.Equal(t, uint(20230723103656), version)
		assert.False(t, dirty)

		_ = db.Close()
	})

	t.Run("opens an existing DB", func(t *testing.T) {
		t.Parallel()

		db, err := New(testDBPath)
		assert.NoError(t, err)
		assert.NotNil(t, db)
		assert.NotNil(t, db.innerDB)
		assert.NotNil(t, db.queries)

		_ = db.Close()
	})
}
