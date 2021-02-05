package migrate

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

const testDB = "file:test?mode=memory&cache=shared"

type testMigration struct {
	version int
	sql     string
	err     error
}

func (m testMigration) Version() int {
	return m.version
}

func (m testMigration) Run(tx *sql.Tx) error {
	if m.err != nil {
		return m.err
	}
	_, err := tx.Exec(m.sql)
	return err
}

var _ Migration = &testMigration{}

func prepareMigrations(in []testMigration) []Migration {
	out := make([]Migration, len(in))
	for i, m := range in {
		out[i] = Migration(m)
	}
	return out
}

func TestMigrate(t *testing.T) {
	tests := map[string]struct {
		migrations []testMigration
		// expected
		vers   int
		failed bool
	}{
		"ordered": {
			migrations: []testMigration{
				{
					version: 1,
					sql:     "create table if not exists test1 (pk bigint not null primary key)",
					err:     nil,
				},
				{
					version: 2,
					sql:     "insert into test1 (pk) values (1)",
					err:     nil,
				},
				{
					version: 3,
					sql:     "insert into test1 (pk) values (2)",
					err:     nil,
				},
				{
					version: 4,
					sql:     "insert into test1 (pk) values (3)",
					err:     nil,
				},
			},
			vers:   4,
			failed: false,
		},
		"errorAt5": {
			migrations: []testMigration{
				{
					version: 1,
					sql:     "create table if not exists test1 (pk bigint not null primary key)",
					err:     nil,
				},
				{
					version: 2,
					sql:     "insert into test1 (pk) values (1)",
					err:     nil,
				},
				{
					version: 3,
					sql:     "insert into test1 (pk) values (2)",
					err:     nil,
				},
				{
					version: 4,
					sql:     "insert into test1 (pk) values (3)",
					err:     nil,
				},
				{
					version: 5,
					sql:     "insert into nonexistant (pk) values (2)",
					err:     nil,
				},
			},
			vers:   4,
			failed: true,
		},
		"unordered": {
			migrations: []testMigration{
				{
					version: 3,
					sql:     "insert into test1 (pk) values (2)",
					err:     nil,
				},
				{
					version: 4,
					sql:     "insert into test1 (pk) values (3)",
					err:     nil,
				},
				{
					version: 2,
					sql:     "insert into test1 (pk) values (1)",
					err:     nil,
				},
				{
					version: 1,
					sql:     "create table if not exists test1 (pk bigint not null primary key)",
					err:     nil,
				},
			},
			vers:   4,
			failed: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			db, err := sql.Open("sqlite3", testDB)
			if err != nil {
				t.Fatalf("DB setup failed: %v", err)
			}
			defer db.Close()

			migrator := Migrator{
				db:         db,
				migrations: prepareMigrations(tt.migrations),
			}

			// No version yet
			v, err := migrator.Version()
			if err != nil {
				t.Fatalf("Migrator.Version() failed: %v", err)
			}
			if v != NilVersion {
				t.Fatalf("Migrator.Version() should be NilVersion, got %d", v)
			}

			if err = migrator.Migrate(); (err != nil) != tt.failed {
				t.Errorf("got %s, unexpected", err)
			}
			v, err = migrator.Version()
			if err != nil {
				t.Fatalf("Migrator.Version() failed: %v", err)
			}
			if v != tt.vers {
				t.Errorf("got %d, want %d", v, tt.vers)
			}
		})
	}
}
