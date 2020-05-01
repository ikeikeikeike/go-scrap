```go
package rds

import (
	"context"
	"testing"
)

func TestConn(t *testing.T) {
	t.Helper()

	db := testDB(t)
	defer db.Close()

	repo := &myRepo{db: db}
	if db != repo.Conn() {
		t.Fatal("Miss match Conn() value")
	}
}

func TestCreate(t *testing.T) {
	t.Helper()

	db := testDB(t)
	defer db.Close()

	repo := &myRepo{db: db}

	m := &models.my{ID: 1}
	if err := repo.Create(context.TODO(), db, m); err != nil {
		t.Fatalf("Create() error: %s", err)
	}
}
```
