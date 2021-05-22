// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by Apache-2.0
// license that can be found in the LICENSE file.

package test

import (
	"context"
	"flag"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var flagParseOnce sync.Once

func NewTestDB(t *testing.T) (testDB *gorm.DB, cleanup func(...string) error) {
	dsn := os.ExpandEnv("postgres://$PGUSER:$PGPASSWORD@$PGHOST:$PGPORT/$PGDATABASE?sslmode=$PGSSLMODE")
	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		t.Fatalf("Failed to open connection: %v", err)
	}

	ctx := context.Background()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	dbname := "cardinal-test-" + strconv.FormatUint(rng.Uint64(), 10)

	err = db.WithContext(ctx).Exec(`CREATE DATABASE ` + QuoteIdentifier(dbname)).Error
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	cfg, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("Failed to parse DSN: %v", err)
	}
	cfg.Path = "/" + dbname

	flagParseOnce.Do(flag.Parse)

	testDB, err = gorm.Open(postgres.Open(cfg.String()))
	if err != nil {
		t.Fatalf("Failed to open test connection: %v", err)
	}

	t.Cleanup(func() {
		defer func() {
			if database, err := db.DB(); err == nil {
				_ = database.Close()
			}
		}()

		if t.Failed() {
			t.Logf("DATABASE %s left intact for inspection", dbname)
			return
		}

		database, err := testDB.DB()
		if err != nil {
			t.Fatalf("Failed to get currently open database: %v", err)
		}

		err = database.Close()
		if err != nil {
			t.Fatalf("Failed to close currently open database: %v", err)
		}

		err = db.WithContext(ctx).Exec(`DROP DATABASE ` + QuoteIdentifier(dbname)).Error
		if err != nil {
			t.Fatalf("Failed to drop test database: %v", err)
		}
	})

	return testDB, func(tables ...string) error {
		if t.Failed() {
			return nil
		}

		for _, table := range tables {
			err := testDB.WithContext(ctx).Exec(`TRUNCATE TABLE ` + QuoteIdentifier(table) + ` RESTART IDENTITY CASCADE`).Error
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// QuoteIdentifier quotes an "identifier" (e.g. a table or a column name) to be
// used as part of an SQL statement.
func QuoteIdentifier(s string) string {
	return `"` + strings.Replace(s, `"`, `""`, -1) + `"`
}
