package migrate_test

import (
	"database/sql"
	"embed"
	"fmt"

	"src.userspace.com.au/migrate"
)

func ExampleNewFileMigrator() {
	db, err := sql.Open("sqlite3", "file:test?mode=memory&cache=shared")
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}
	defer db.Close()

	// Relative path to migration files
	migrator, err := migrate.NewFileMigrator(db, "file://testdata/")
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}

	// Migrate all the way
	err = migrator.Migrate()
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}

	v, err := migrator.Version()
	fmt.Printf("database at version %d\n", v)
	// Output: database at version 2
}

func ExampleNewFSMigrator() {
	//go:embed testdata/*.sql
	var migrations embed.FS

	db, err := sql.Open("sqlite3", "file:test?mode=memory&cache=shared")
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}
	defer db.Close()

	// Relative path to migration files
	migrator, err := migrate.NewFSMigrator(db, migrations)
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}

	// Migrate all the way
	err = migrator.Migrate()
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}

	v, err := migrator.Version()
	fmt.Printf("database at version %d\n", v)
	// Output: database at version 2
}

func ExampleNewStringMigrator() {

	migrations := []string{
		"create table if not exists test1 (pk bigint not null primary key);",
		"insert into test1 (pk) values (1)",
	}

	db, err := sql.Open("sqlite3", "file:test?mode=memory&cache=shared")
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}
	defer db.Close()

	// Relative path to migration files
	migrator, err := migrate.NewStringMigrator(db, migrations)
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}

	// Migrate all the way
	err = migrator.Migrate()
	if err != nil {
		fmt.Printf("error: %s\n", err)
	}

	v, err := migrator.Version()
	fmt.Printf("database at version %d\n", v)
	// Output: database at version 2
}
