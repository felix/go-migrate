# Go migrations

[![GoDoc](https://godoc.org/github.com/felix/go-migrate?status.svg)](http://godoc.org/github.com/felix/go-migrate)
[![Build Status](https://cloud.drone.io/api/badges/felix/go-migrate/status.svg)](https://cloud.drone.io/felix/go-migrate)
[![Coverage Status](https://coveralls.io/repos/github/felix/go-migrate/badge.svg?branch=master)](https://coveralls.io/github/felix/go-migrate?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/felix/go-migrate)](https://goreportcard.com/report/github.com/felix/go-migrate)

A very simple migration library for your Go projects.

```go
import "github.com/felix/go-migrate"
```

or

```go
import "src.userspace.com.au/felix/go-migrate"
```

## Features

- Any database/sql drivers should work. See https://github.com/golang/go/wiki/SQLDrivers

- Each migration is run in a transaction so there should be no 'dirty' state.

- Only migrates "up". I have never used a "down" migration.

- Enables a callback function to suit whatever logging implementation you choose.

## Usage

```go
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

