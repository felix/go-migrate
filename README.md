# Go migrations

A very simple migration library for your Go projects.

## Features

- Any database/sql compatible drivers should work. See
  https://github.com/golang/go/wiki/SQLDrivers

- Each migration is run in a transaction so there should be no 'dirty' state.

- Only migrates "up". I have never used a "down" migration.

- Enables a callback function to suit whatever logging implementation you
  choose.

- Supports Go1.16's io/fs interface for embedded migrations.

## Example usage

```go
import "src.userspace.com.au/migrate"
// or import "github.com/felix/migrate"

db, err := sql.Open("pgx", uri)
//db, err := sql.Open("sqlite3", uri)
if err != nil {
    return err
}
defer db.Close()

// Relative path to migration files
migrator, err := migrate.NewFileMigrator(db, "file://migrations/")
if err != nil {
    return fmt.Errorf("failed to create migrator: %s", err)
}

// Migrate all the way
err = migrator.Migrate()
if err != nil {
    return fmt.Errorf("failed to migrate: %s", err)
}

v, err := migrator.Version()
fmt.Printf("database at version %d\n", v)
```

