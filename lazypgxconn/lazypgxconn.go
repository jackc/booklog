// package lazypgxconn is a lazy wrapper for a *pgx.Conn.
//
// The primary purpose is for interfaces such as HTTP handlers where the database connection may be needed from 0 to
// many times. If it is needed many times then using a *pgxpool.Pool directly would incur multiple inernal Acquire and
// Release calls. In addition, this means the queries could run on different connections. Acquire and Release could be
// called manually, but that adds unwelcome boilerplate. The obvious alternative is for wrapping code or middleware to
// handle the Acquire and Release and directly expose a *pgx.Conn. But that is inefficient if the handler does not need
// the connection as it incurs the Acquire and Release as well as consuming a connection for the duration of the
// handler's execution.
package lazypgxconn

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// TODO - maybe change acquire and release funcs to be part of a config struct. That could allow configuring a Hijack
// method.
//
// Also consider whether there should be some means of getting the memo. This would allow getting the *pgxpool.Conn.

// Conn is a wrapper around *pgx.Conn that defers creating or acquiring the connection until the first method call.
// Release must be called to close or release the connection.
//
// Conn wraps all methods of *pgx.Conn that return an error. Methods that do not return an error can be called by
// getting the underlying *pgx.Conn via Conn().
type Conn struct {
	acquire func() (conn *pgx.Conn, memo any, err error)
	release func(conn *pgx.Conn, memo any) error
	conn    *pgx.Conn
	memo    any
}

// New creates a new Conn. The acquire function will be called on first use. The release function will be called when
// Release is called if acquire has previously been called.
func New(acquire func() (conn *pgx.Conn, memo any, err error), release func(conn *pgx.Conn, memo any) error) *Conn {
	return &Conn{
		acquire: acquire,
		release: release,
	}
}

// Conn returns the underlying pgx.Conn. This can be used to call *pgx.Conn methods that do not return errors.
func (c *Conn) Conn() (*pgx.Conn, error) {
	if c.conn == nil {
		conn, memo, err := c.acquire()
		if err != nil {
			return nil, err
		}
		c.conn = conn
		c.memo = memo
	}
	return c.conn, nil
}

// Release releases the underlying connection. If the connection was never acquired or has already been released then no
// action will be taken.
func (c *Conn) Release() error {
	if c.conn != nil {
		err := c.release(c.conn, c.memo)
		if err != nil {
			return err
		}
		c.conn = nil
		c.memo = nil
	}
	return nil
}

func (c *Conn) Begin(ctx context.Context) (pgx.Tx, error) {
	conn, err := c.Conn()
	if err != nil {
		return nil, err
	}
	return conn.Begin(ctx)
}

func (c *Conn) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	conn, err := c.Conn()
	if err != nil {
		return nil, err
	}
	return conn.BeginTx(ctx, txOptions)
}

func (c *Conn) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	conn, err := c.Conn()
	if err != nil {
		return 0, err
	}
	return conn.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

func (c *Conn) Deallocate(ctx context.Context, name string) error {
	conn, err := c.Conn()
	if err != nil {
		return err
	}
	return conn.Deallocate(ctx, name)
}

func (c *Conn) DeallocateAll(ctx context.Context) error {
	conn, err := c.Conn()
	if err != nil {
		return err
	}
	return conn.DeallocateAll(ctx)
}

func (c *Conn) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	conn, err := c.Conn()
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	return conn.Exec(ctx, sql, arguments...)
}

func (c *Conn) LoadType(ctx context.Context, typeName string) (*pgtype.Type, error) {
	conn, err := c.Conn()
	if err != nil {
		return nil, err
	}
	return conn.LoadType(ctx, typeName)
}

func (c *Conn) Ping(ctx context.Context) error {
	conn, err := c.Conn()
	if err != nil {
		return err
	}
	return conn.Ping(ctx)
}

func (c *Conn) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	conn, err := c.Conn()
	if err != nil {
		return nil, err
	}
	return conn.Prepare(ctx, name, sql)
}

func (c *Conn) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	conn, err := c.Conn()
	if err != nil {
		return errRows{err: err}, err
	}
	return conn.Query(ctx, sql, args...)
}

type errRows struct {
	err error
}

func (errRows) Close()                                       {}
func (e errRows) Err() error                                 { return e.err }
func (errRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (errRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (errRows) Next() bool                                   { return false }
func (e errRows) Scan(dest ...any) error                     { return e.err }
func (e errRows) Values() ([]any, error)                     { return nil, e.err }
func (e errRows) RawValues() [][]byte                        { return nil }
func (e errRows) Conn() *pgx.Conn                            { return nil }

func (c *Conn) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	conn, err := c.Conn()
	if err != nil {
		return errRow{err: err}
	}
	return conn.QueryRow(ctx, sql, args...)
}

type errRow struct {
	err error
}

func (e errRow) Scan(dest ...any) error { return e.err }

func (c *Conn) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	conn, err := c.Conn()
	if err != nil {
		return errBatchResults{err: err}
	}
	return conn.SendBatch(ctx, b)
}

type errBatchResults struct {
	err error
}

func (br errBatchResults) Exec() (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, br.err
}

func (br errBatchResults) Query() (pgx.Rows, error) {
	return errRows{err: br.err}, br.err
}

func (br errBatchResults) QueryRow() pgx.Row {
	return errRow{err: br.err}
}

func (br errBatchResults) Close() error {
	return br.err
}

func (c *Conn) WaitForNotication(ctx context.Context) (*pgconn.Notification, error) {
	conn, err := c.Conn()
	if err != nil {
		return nil, err
	}
	return conn.WaitForNotification(ctx)
}
