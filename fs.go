// +build go1.16

package migrate

import (
	"database/sql"
	"fmt"
	"io/fs"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
)

type fsMigration struct {
	name string
	fs   fs.FS
}

func NewFSMigrator(db *sql.DB, fs fs.ReadDirFS, opts ...Option) (*Migrator, error) {
	migrations, err := readFS(fs, ".")
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
func (fm fsMigration) Version() int {
	m := validFilename.FindStringSubmatch(path.Base(fm.name))
	if len(m) == 4 {
		if version, err := strconv.ParseInt(m[1], 10, 32); err == nil {
			return int(version)
		}
	}
	return -1
}

// Run executes the migration
// It implements the Migration interface
func (fm fsMigration) Run(tx *sql.Tx) error {
	f, err := fm.fs.Open(fm.name)
	if err != nil {
		return err
	}
	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	_, err = tx.Exec(string(buf[:]))
	return err
}

// Read all files in path
// They will be sorted by the migrator according to Version()
func readFS(fs fs.ReadDirFS, root string) ([]Migration, error) {

	// Scan entire directory
	files, err := fs.ReadDir(root)
	if err != nil {
		return nil, err
	}

	migrations := make([]Migration, 0)
	seen := make(map[int]bool)

	for _, fi := range files {
		if strings.HasPrefix(fi.Name(), ".") {
			continue
		}
		if fi.IsDir() {
			fms, err := readFS(fs, fi.Name())
			if err != nil {
				return nil, err
			}
			migrations = append(migrations, fms...)
		} else {
			fm := fsMigration{name: path.Join(root, fi.Name()), fs: fs}

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
