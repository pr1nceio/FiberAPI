package utils

import (
	"context"
	"database/sql"
	"fmt"
	mysql "github.com/cradio/gorm_mysql"
	gorm "github.com/cradio/gormx"
	"github.com/cradio/gormx/logger"
	"log"
	"os"
)

// MultiSQL allows to store multiple raw *sql.DB connections and create *gorm.DB instances out of them
type MultiSQL struct {
	db       *sql.DB
	conns    map[string]*gorm.DB
	mutators map[string]func(db string) string
}

func NewMultiSQL(db *sql.DB) *MultiSQL {
	return &MultiSQL{db: db, conns: make(map[string]*gorm.DB), mutators: make(map[string]func(db string) string)}
}

// AddMutator adds database name mutator function with specified name
//
// Mutator example:
//
//	func(db string) string {
//	   return db+"_backup"
//	}
func (m *MultiSQL) AddMutator(name string, f func(db string) string) {
	m.mutators[name] = f
}

func (m *MultiSQL) DelMutator(name string) {
	delete(m.mutators, name)
}

func (m *MultiSQL) Mutate(name, value string) string {
	if v, ok := m.mutators[name]; ok {
		return v(value)
	}
	return value
}

func (m *MultiSQL) Raw() *sql.DB {
	return m.db
}

func (m *MultiSQL) OpenMutated(name, db string) (*gorm.DB, error) {
	return m.Open(m.Mutate(name, db))
}

func (m *MultiSQL) DisposeMutated(name, db string) {
	m.Dispose(m.Mutate(name, db))
}

func (m *MultiSQL) OpenCached(db string) (*gorm.DB, error) {

	if v, ok := m.conns[db]; ok {
		return v, nil
	}
	return m.Open(db)
}

func (m *MultiSQL) Open(db string) (*gorm.DB, error) {

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			LogLevel: logger.Info, // Log level
		},
	)
	gdb, err := gorm.Open(mysql.New(mysql.Config{Conn: m.db}), &gorm.Config{
		Logger:                 newLogger,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, err
	}
	gdb = gdb.Set("db", db)
	m.conns[db] = gdb
	return gdb, nil
}

func (m *MultiSQL) Dispose(db string) {
	delete(m.conns, db)
}

func (m *MultiSQL) UTable(db *gorm.DB, table string) *gorm.DB {
	if cdb, ok := db.Get("db"); ok {
		return db.WithContext(context.Background()).Table(fmt.Sprintf("`%s`.`%s`", cdb.(string), table))
	}
	return db
}
