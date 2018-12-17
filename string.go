package migrate

import (
	"database/sql"
)

type stringMigration struct {
	version int64
	sql     string
}

// The Version is extracted from the filename
// It implements the Migration interface
func (sm stringMigration) Version() int64 {
	return sm.version
}

// Run executes the migration
// It implements the Migration interface
func (sm stringMigration) Run(tx *sql.Tx) error {
	_, err := tx.Exec(sm.sql)
	return err
}

// NewStringMigrator creates a new set of migrations from a slice of strings
// Each one is run in a transaction.
func NewStringMigrator(db *sql.DB, src []string, opts ...Option) (*Migrator, error) {
	migrations := make([]Migration, 0)

	for i, s := range src {
		migrations = append(migrations, stringMigration{
			version: int64(i + 1),
			sql:     s,
		})
	}

	m := Migrator{db: db, migrations: migrations}

	for _, opt := range opts {
		if err := opt(&m); err != nil {
			return nil, err
		}
	}

	return &m, nil
}
