package driver

import (
	"context"
	"database/sql"
	"sync"

	"github.com/neilotoole/lg"

	"github.com/neilotoole/sq/libsq/core/cleanup"
	"github.com/neilotoole/sq/libsq/source"
)

// ScratchSrcFunc is a function that returns a scratch source.
// The caller is responsible for invoking cleanFn.
type ScratchSrcFunc func(log lg.Log, name string) (src *source.Source, cleanFn func() error, err error)

// scratchDatabase implements driver.Database.
// It exists as a separate type mostly to support cleanup of the
// created scratch database during Close. Possibly this type can
// be ditched entirely.
type scratchDatabase struct {
	log     lg.Log
	impl    Database
	cleanup *cleanup.Cleanup
	mu      sync.Mutex
	closed  bool
}

// DB implements driver.Database.
func (d *scratchDatabase) DB() *sql.DB {
	return d.impl.DB()
}

// TableMetadata implements driver.Database.
func (d *scratchDatabase) TableMetadata(ctx context.Context, tblName string) (*source.TableMetadata, error) {
	return d.impl.TableMetadata(ctx, tblName)
}

// SQLDriver implements driver.Database.
func (d *scratchDatabase) SQLDriver() SQLDriver {
	return d.impl.SQLDriver()
}

// Source implements driver.Database.
func (d *scratchDatabase) Source() *source.Source {
	return d.impl.Source()
}

// SourceMetadata implements driver.Database.
func (d *scratchDatabase) SourceMetadata(ctx context.Context) (*source.Metadata, error) {
	return d.impl.SourceMetadata(ctx)
}

// Close implements driver.Database.
func (d *scratchDatabase) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		panic("already closed")
	}

	d.log.Debugf("Closing scratch database [%s]", d.impl.Source())

	//debug.PrintStack()
	// No need to explicitly invoke c.impl.Close because it
	// has already been added to c.cleanup.
	err := d.cleanup.Run()
	d.log.Debugf("Closed scratch database [%s]: err=%v", d.impl.Source(), err)
	d.closed = true
	return err
}
