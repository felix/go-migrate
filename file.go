package migrate

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
)

// Regex matches the following pattern:
// 123_name.ext
var validFilename = regexp.MustCompile(`^([0-9]+)_(.*)\.(.*)$`)

// Just the path to the migration file
type fileMigration string

// NewFileMigrator creates a new set of migrations from a path
// Each one is run in a transaction.
func NewFileMigrator(db *sql.DB, path string, opts ...Option) (*Migrator, error) {
	migrations, err := readFiles(path)
	if err != nil {
		return nil, err
	}

	m := Migrator{db: db, migrations: migrations}

	for _, opt := range opts {
		if err = opt(&m); err != nil {
			return nil, err
		}
	}

	return &m, nil
}

// The Version is extracted from the filename
// It implements the Migration interface
func (fm fileMigration) Version() int {
	m := validFilename.FindStringSubmatch(path.Base(string(fm)))
	if len(m) == 4 {
		if version, err := strconv.ParseInt(m[1], 10, 32); err == nil {
			return int(version)
		}
	}
	return -1
}

// Run executes the migration
// It implements the Migration interface
func (fm fileMigration) Run(tx *sql.Tx) error {
	r, err := os.Open(string(fm))
	if err != nil {
		return err
	}
	defer r.Close()

	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	_, err = tx.Exec(string(buf[:]))
	return err
}

// Read all files in path
// They will be sorted by the migrator according to Version()
func readFiles(uri string) (migrations []Migration, err error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	// Host might be `.`
	p := u.Host + u.Path

	if len(p) == 0 {
		// Default to current directory
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		p = wd
	} else if p[0:1] == "." || p[0:1] != "/" {
		// Ensure path is absolute
		abs, err := filepath.Abs(p)
		if err != nil {
			return nil, err
		}
		p = abs
	}

	// Scan entire directory
	files, err := ioutil.ReadDir(p)
	if err != nil {
		return nil, err
	}

	seen := make(map[int]bool)

	for _, fi := range files {
		if !fi.IsDir() {
			fm := fileMigration(path.Join(p, fi.Name()))

			// Ignore invalid filenames
			v := fm.Version()
			if v == -1 {
				fmt.Printf("invalid version %d for %s\n", v, fm)
				continue
			}
			if seen[v] {
				fmt.Printf("duplicate version %d for %s\n", v, fm)
				continue
			}
			migrations = append(migrations, fm)
			seen[v] = true
		}
	}
	if len(migrations) == 0 {
		return nil, fmt.Errorf("no migrations found")
	}
	return migrations, nil
}
