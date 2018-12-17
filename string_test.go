package migrate

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestStringMigrator(t *testing.T) {
	db, err := sql.Open("sqlite3", testDB)
	if err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	defer db.Close()

	tests := []string{
		"create table if not exists test1 (pk bigint not null primary key);",
		"insert into test1 (pk) values (1)",
	}

	migrator, err := NewStringMigrator(db, tests)
	if err != nil {
		t.Fatal(err)
	}

	if v, _ := migrator.Version(); v != NilVersion {
		t.Errorf("expected migration version NilVersion, got %d", v)
	}

	if c := len(migrator.migrations); c != 2 {
		t.Errorf("expected migration count = 2, got %d", c)
	}

	err = migrator.MigrateTo(1)
	if err != nil {
		t.Fatalf("Migrator.MigrateTo(3) failed: %v", err)
	}

	v, err := migrator.Version()
	if err != nil {
		t.Fatalf("Migrator.Version() failed: %v", err)
	}

	if int(v) != len(migrator.migrations)-1 {
		t.Errorf("expected migration version %d, got %d", len(migrator.migrations)-1, v)
	}
}
