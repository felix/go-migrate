package migrate

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

const testDB = "file:test?mode=memory&cache=shared"

type mig struct {
	version int64
	sql     string
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
	{version: 4, sql: "insert into nonexistant (pk) values (2)"},
	{version: 1, sql: "create table if not exists test1 (pk bigint not null primary key)"},
}

func TestMigrate(t *testing.T) {
	// Load migrations
	var migrations []Migration
	for _, m := range testMigrations {
		migrations = append(migrations, createMigration(m.version, m.sql))
	}

	db, err := sql.Open("sqlite3", testDB)
	if err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	defer db.Close()
	//db.SetMaxOpenConns(1)

	migrator := Migrator{
		db:         db,
		migrations: migrations,
	}

	v, err := migrator.Version()
	if err != nil {
		t.Fatalf("Migrator.Version() failed: %v", err)
	}
	if v != NilVersion {
		t.Fatalf("Migrator.Version() should be NilVersion, got %d", v)
	}

	err = migrator.MigrateTo(3)
	if err != nil {
		t.Fatalf("Migrator.MigrateTo(3) failed: %v", err)
	}

	v, err = migrator.Version()
	if err != nil {
		t.Fatalf("Migrator.Version() failed: %v", err)
	}

	if int(v) != len(migrations)-1 {
		t.Errorf("expected migration version %d, got %d", len(migrations)-1, v)
	}

	err = migrator.MigrateTo(4)
	if err == nil {
		t.Fatalf("Migrator.MigrateTo(4) should have failed")
	}

	v, err = migrator.Version()
	if err != nil {
		t.Fatalf("Migrator.Version() failed: %v", err)
	}

	if int(v) != len(migrations)-1 {
		t.Errorf("expected migration version %d, got %d", len(migrations)-1, v)
	}

	var result int64
	err = db.QueryRow(`select pk from test1`).Scan(&result)
	if err != nil {
		t.Fatal(err)
	}
}
