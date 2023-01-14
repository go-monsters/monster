package postgresql

import (
	"context"
	"sync"
	"time"

	"github.com/go-monsters/monster/pkg/database"
	postgres "go.elastic.co/apm/module/apmgormv2/v2/driver/postgres"
	"gorm.io/gorm"
)

type Postgresql struct {
	address    string
	dbConnOnce sync.Once
	db         *gorm.DB
}

func New(address string) database.Database {
	return &Postgresql{
		address: address,
	}
}

func (m *Postgresql) GetConnection(ctx context.Context) *gorm.DB {
	if m.db == nil {
		m.dbConnOnce.Do(func() {
			var err error
			m.db, err = gorm.Open(postgres.Open(m.address), &gorm.Config{})
			if err != nil {
				panic(err)
			}

			db, err := m.db.DB()
			if err != nil {
				panic(err)
			}
			db.SetMaxIdleConns(10)
			db.SetMaxOpenConns(20)
			db.SetConnMaxLifetime(time.Hour)
		})
	}

	return m.db.WithContext(ctx)
}
