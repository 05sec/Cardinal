// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by Apache-2.0
// license that can be found in the LICENSE file.

package test

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewTestDB(t *testing.T) (testDB *gorm.DB, cleanup func(...string) error) {
	dsn := os.ExpandEnv("$DBUSER:$DBPASSWORD@tcp($DBHOST:$DBPORT)/?parseTime=true&loc=Local&charset=utf8mb4,utf8")
	db, err := gorm.Open(mysql.Open(dsn))
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

	dsn = os.ExpandEnv("$DBUSER:$DBPASSWORD@tcp($DBHOST:$DBPORT)/" + dbname + "?parseTime=true&loc=Local&charset=utf8mb4,utf8")
	testDB, err = gorm.Open(mysql.Open(dsn))
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

		err := testDB.WithContext(ctx).Exec(`DROP DATABASE ` + QuoteIdentifier(dbname)).Error
		if err != nil {
			t.Fatalf("Failed to drop test database: %v", err)
		}
	})

	return testDB, func(tables ...string) error {
		if t.Failed() {
			return nil
		}

		for _, table := range tables {
			err := testDB.WithContext(ctx).Exec(`TRUNCATE TABLE ` + QuoteIdentifier(table)).Error
			if err != nil {
				return err
			}
		}
		return nil
	}
}

// QuoteIdentifier quotes an "identifier" (e.g. a table or a column name) to be
// used as part of an SQL statement.
func QuoteIdentifier(name string) string {
	end := strings.IndexRune(name, 0)
	if end > -1 {
		name = name[:end]
	}
	return "`" + strings.Replace(name, "`", "``", -1) + "`"
}
