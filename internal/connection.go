package internal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/elgris/sqrl"
	"github.com/google/wire"
	"github.com/jmoiron/sqlx"
	"github.com/tinhtran24/xo-patcher/utils"
)

type IDb interface {
	Select(ctx context.Context, dest interface{}, sqlizer sqrl.Sqlizer) error
	Get(ctx context.Context, dest interface{}, sqlizer sqrl.Sqlizer) error
	Exec(ctx context.Context, sqlizer sqrl.Sqlizer) (sql.Result, error)
	BeginTxx(ctx context.Context) (*sqlx.Tx, error)
}

type DBOptions struct {
	FileGen IPatchSQLFileGen
}

type DB struct {
	*DBOptions
	DB *sqlx.DB
}

var NewDB = wire.NewSet(
	wire.Struct(new(DBOptions), "*"),
	OpenConnection,
	wire.Bind(new(IDb), new(DB)),
)

func OpenConnection(ctx context.Context, options *DBOptions) *DB {

	connection, err := getConnectionContext(ctx)
	if err != nil {
		log.Fatal(err)
	}

	sqlxDB, err := sqlx.Open("mysql", connection)
	if err != nil {
		log.Fatal(err)
	}
	return &DB{DBOptions: options, DB: sqlxDB}
}

func (db *DB) Get(ctx context.Context, dest interface{}, sqlizer sqrl.Sqlizer) error {
	query, args, err := sqlizer.ToSql()
	if err != nil {
		return err
	}

	if tx := getTransactionContext(ctx); tx != nil {
		err = tx.GetContext(ctx, dest, query, args...)
	} else {
		err = db.DB.GetContext(ctx, dest, query, args...)
	}

	if err != nil {
		return err
	}
	return nil
}

func (db *DB) Select(ctx context.Context, dest interface{}, sqlizer sqrl.Sqlizer) error {
	query, args, err := sqlizer.ToSql()
	if err != nil {
		return err
	}

	if tx := getTransactionContext(ctx); tx != nil {
		err = tx.SelectContext(ctx, dest, query, args...)
	} else {
		err = db.DB.SelectContext(ctx, dest, query, args...)
	}

	if err != nil {
		return err
	}
	return nil
}

func (db *DB) Exec(ctx context.Context, sqlizer sqrl.Sqlizer) (sql.Result, error) {

	// If prod don't attempt exec, just write to file
	if getProdContext(ctx) {
		err := db.FileGen.Write(sqlizer)
		return nil, err
	}

	query, args, err := sqlizer.ToSql()
	if err != nil {
		return nil, err
	}

	var res sql.Result
	if tx := getTransactionContext(ctx); tx != nil {
		res, err = tx.ExecContext(ctx, query, args...)
	} else {
		res, err = db.DB.ExecContext(ctx, query, args...)
	}

	if err != nil {
		return nil, err
	}

	// write to file if query was successful
	err = db.FileGen.Write(sqlizer)

	return res, err
}

func (db *DB) BeginTxx(ctx context.Context) (*sqlx.Tx, error) {
	return db.DB.BeginTxx(ctx, nil)
}

// Transaction

type key string

const TransactionKey key = "transaction_key"

func WrapInTransaction(ctx context.Context, db IDb, f func(ctx context.Context) error) error {

	// if transaction already exists
	tx := getTransactionContext(ctx)
	if tx != nil {
		return f(ctx)
	}

	tx, err := db.BeginTxx(ctx)
	if err != nil {
		return err
	}

	// handle traction
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			fmt.Println("Rollback due to panic")
			panic(r)
		}
		if getCommitContext(ctx) {
			tx.Commit()
		} else {
			tx.Rollback()
		}
	}()

	// save tx for next time. if same connection. Maybe not needed for xo-patcher 🤔
	newContext := context.WithValue(ctx, TransactionKey, tx)
	return f(newContext)
}

// Get Contexts

func getTransactionContext(ctx context.Context) *sqlx.Tx {
	if tx, ok := ctx.Value(TransactionKey).(*sqlx.Tx); ok {
		return tx
	}
	return nil
}

func getConnectionContext(ctx context.Context) (string, error) {
	if value, ok := ctx.Value(utils.Connection).(string); ok {
		return value, nil
	}
	return "", errors.New("connection context invalid")
}

func getCommitContext(ctx context.Context) bool {
	if value, ok := ctx.Value(utils.Commit).(bool); ok {
		return value
	}
	return false
}

func getProdContext(ctx context.Context) bool {
	if value, ok := ctx.Value(utils.Prod).(bool); ok {
		return value
	}
	return false
}
