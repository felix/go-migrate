package migrate

import (
	"database/sql"
	"fmt"
	"sort"
	"time"
)

const (
	// NilVersion is a Claytons version
	// "the version you are at when you are not at a version"
	NilVersion = -1
)

// A Migrator collates and runs migrations
type Migrator struct {
	db           *sql.DB
	migrations   []Migration
	versionTable *string
	stmts        map[string]*sql.Stmt
	prepared     bool
}

// Migration interface
type Migration interface {
	// The version of this migration
	Version() int64
	// Run the migration
	Run(*sql.Tx) error
}

// ResultFunc is the callback signature
type ResultFunc func(int64, int64, error)

// Sort those migrations
type sorted []Migration

func (s sorted) Len() int           { return len(s) }
func (s sorted) Less(i, j int) bool { return s[i].Version() < s[j].Version() }
func (s sorted) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// Version reports the current version of the database
func (m *Migrator) Version() (int64, error) {
	err := m.prepareForMigration()
	if err != nil {
		return NilVersion, err
	}

	rows, err := m.stmts["getVersion"].Query()
	if rows.Next() {
		var version int64
		err = rows.Scan(&version)
		if err == sql.ErrNoRows {
			return NilVersion, nil
		}
		if err == nil {
			return version, nil
		}
	}
	return 0, err
}

// Migrate migrates the database to the highest possible version
func (m *Migrator) Migrate(cb ResultFunc) error {
	err := m.prepareForMigration()
	if err != nil {
		return err
	}

	// Get the last available migration
	v := m.migrations[len(m.migrations)-1].Version()
	return m.MigrateTo(v, cb)
}

// MigrateTo migrates the database to the specified version
func (m *Migrator) MigrateTo(toVersion int64, cb ResultFunc) error {
	err := m.prepareForMigration()
	if err != nil {
		return err
	}

	maxVersion := m.migrations[len(m.migrations)-1].Version()

	currVersion, err := m.Version()
	if err != nil {
		return err
	}

	if currVersion >= toVersion {
		if cb != nil {
			go cb(maxVersion, currVersion, nil)
		}
		return nil
	}

	for _, mig := range m.migrations {
		nextVersion := mig.Version()

		// Skip old migrations
		if nextVersion <= currVersion {
			continue
		}

		// Ensure contiguous
		if currVersion != NilVersion && nextVersion != currVersion+1 {
			return fmt.Errorf("non-contiguous migration: %v -> %v", currVersion, nextVersion)
		}

		if currVersion < nextVersion && nextVersion <= toVersion {
			err = func() error {
				fmt.Println("running migration", nextVersion)
				// Start a transaction
				tx, err := m.db.Begin()
				if err != nil {
					if cb != nil {
						go cb(maxVersion, currVersion, err)
					}
					return err
				}
				defer tx.Rollback()

				// Run the migration
				if err = mig.Run(tx); err != nil {
					if cb != nil {
						go cb(maxVersion, currVersion, err)
					}
					return err
				}
				// Update the version entry
				fmt.Println("updating version")
				if err = m.setVersion(tx, nextVersion); err != nil {
					if cb != nil {
						go cb(maxVersion, currVersion, err)
					}
					return err
				}
				// Commit the transaction
				fmt.Println("committing version")
				return tx.Commit()
			}()
			if err != nil {
				if cb != nil {
					go cb(maxVersion, currVersion, err)
				}
				return err
			}
			if cb != nil {
				go cb(maxVersion, currVersion, nil)
			}
		}
		currVersion = nextVersion
	}

	return nil
}

func (m *Migrator) setVersion(tx *sql.Tx, version int64) (err error) {
	if version >= 0 {
		_, err = tx.Stmt(m.stmts["insertVersion"]).Exec(version, time.Now().Unix())
	}
	return err
}

func (m *Migrator) prepareForMigration() error {
	if m.prepared {
		return nil
	}

	if m.versionTable == nil {
		vt := "current_schema_version"
		m.versionTable = &vt
	}

	if _, err := m.db.Exec(fmt.Sprintf(createTableSQL, *m.versionTable)); err != nil {
		return err
	}

	if err := m.prepareStmts(); err != nil {
		return err
	}

	sort.Sort(sorted(m.migrations))

	m.prepared = true
	return nil
}

func (m *Migrator) prepareStmts() error {
	m.stmts = make(map[string]*sql.Stmt)
	s, err := m.db.Prepare(fmt.Sprintf(getVersionSQL, NilVersion, *m.versionTable))
	if err != nil {
		return err
	}
	m.stmts["getVersion"] = s

	s, err = m.db.Prepare(fmt.Sprintf(insertVersionSQL, *m.versionTable))
	if err != nil {
		return err
	}
	m.stmts["insertVersion"] = s

	return nil
}

const (
	getVersionSQL    = `select coalesce(max(version), %d) from %q`
	insertVersionSQL = `insert into %q (version, applied) values ($1, $2)`

	// Use Unix timestamp for time so it works for SQLite and PostgreSQL
	createTableSQL = `create table if not exists %q (
		version bigint not null primary key,
		applied int
	)`
)
