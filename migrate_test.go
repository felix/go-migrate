package migrate

import (
	"database/sql"
	"fmt"
	"sync"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

const testDB = "file:test?mode=memory&cache=shared"

type mig struct {
	version int64
	sql     string
	cb      ResultFunc
}

func createMigration(v int64, sql string) Migration {
	return &mig{version: v, sql: sql}
}

func (m mig) Version() int64 { return m.version }

func (m mig) Run(tx *sql.Tx) error {
	_, err := tx.Exec(m.sql)
	return err
}

var testMigrations = []struct {
	version int64
	sql     string
}{
	// Out of order please
	{version: 2, sql: "insert into test1 (pk) values (1)"},
	{version: 3, sql: "insert into test1 (pk) values (2)"},
	{version: 5, sql: "insert into nonexistant (pk) values (2)"},
	{version: 1, sql: "create table if not exists test1 (pk bigint not null primary key)"},
	{version: 4, sql: "insert into test1 (pk) values (3)"},
}

func TestMigrate(t *testing.T) {
	// Load migrations
	var migrations []Migration
	for _, m := range testMigrations {
		migrations = append(migrations, createMigration(m.version, m.sql))
	}
	var output string
	var outputMutex sync.RWMutex

	db, err := sql.Open("sqlite3", testDB)
	if err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	defer db.Close()
	//db.SetMaxOpenConns(1)

	migrator := Migrator{
		db:         db,
		migrations: migrations,
		callback: ResultFunc(func(max, curr int64, err error) {
			outputMutex.Lock()
			defer outputMutex.Unlock()
			output = fmt.Sprintf("Maximum: %d, Current: %d, Error: %s\n", max, curr, err)
		}),
	}

	// No version yet
	v, err := migrator.Version()
	if err != nil {
		t.Fatalf("Migrator.Version() failed: %v", err)
	}
	if v != NilVersion {
		t.Fatalf("Migrator.Version() should be NilVersion, got %d", v)
	}

	// Next version
	if err = migrator.MigrateTo(2); err != nil {
		t.Fatalf("Migrator.MigrateTo(2) failed: %v", err)
	}
	v, err = migrator.Version()
	if err != nil {
		t.Fatalf("Migrator.Version() failed: %v", err)
	}
	if int(v) != 2 {
		t.Errorf("expected migration version 2, got %d", v)
	}

	// Previous version
	if err = migrator.MigrateTo(2); err != nil {
		t.Fatalf("Migrator.MigrateTo(2) failed: %v", err)
	}
	v, err = migrator.Version()
	if err != nil {
		t.Fatalf("Migrator.Version() failed: %v", err)
	}
	if int(v) != 2 {
		t.Errorf("expected migration version 2, got %d", v)
	}

	// All the way, with eventual failure
	if err = migrator.Migrate(); err == nil {
		t.Fatalf("Migrator.Migrate() should have failed")
	}
	v, err = migrator.Version()
	if err != nil {
		t.Fatalf("Migrator.Version() failed: %v", err)
	}
	if int(v) != 4 {
		t.Errorf("expected migration version 4, got %d", v)
	}
	expected := "Maximum: 5, Current: 5, Error: migration 5 failed: no such table: nonexistant\n"
	outputMutex.RLock()
	if output != expected {
		t.Errorf("expected output %q, got %q", expected, output)
	}
	outputMutex.RUnlock()

	var result int64
	err = db.QueryRow(`select pk from test1`).Scan(&result)
	if err != nil {
		t.Fatal(err)
	}
}
