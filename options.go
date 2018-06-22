package migrate

//import "context"

// An Option configures a migrator
type Option func(*Migrator) error

// SetVersionTable configures the table used for recording the schema version
func SetVersionTable(vt string) Option {
	return func(m *Migrator) error {
		m.versionTable = &vt
		return nil
	}
}

// SetCallback configures the table used for recording the schema version
func SetCallback(cb ResultFunc) Option {
	return func(m *Migrator) error {
		m.callback = cb
		return nil
	}
}

// SetContext configures the context for queries
/*
func SetContext(ctx context.Context) Option {
	return func(m *Migrator) error {
		m.ctx = ctx
		return nil
	}
}
*/
