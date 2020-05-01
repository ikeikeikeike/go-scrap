package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"
	"time"

	sqle "github.com/src-d/go-mysql-server"
	"github.com/src-d/go-mysql-server/auth"
	"github.com/src-d/go-mysql-server/memory"
	"github.com/src-d/go-mysql-server/server"
	"github.com/volatiletech/sqlboiler/drivers/sqlboiler-mysql/driver"
)

var (
	tUser = "root"
	tHost = "127.0.0.1"
	tPort = 13306
	tName = "tablename"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := driver.MySQLBuildQueryString(tUser, "", tName, tHost, tPort, "false")

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("mysql test server: %s", err)
	}

	return db
}

func createTestDB() *memory.Database {
	return memory.NewDatabase(tName)
}

func createTestTable() {
	out, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		fmt.Printf("no Go-Modules found: %v\n", err)
		os.Exit(30)
	}

	sqlfile := filepath.Join(path.Dir(string(out)), "cmd/migrate/schema.sql")
	filein, err := ioutil.ReadFile(sqlfile)
	if err != nil {
		fmt.Printf("no Go-Modules found: %v\n", err)
		os.Exit(35)
	}

	stdin := bytes.NewReader(filein)
	stderr := &bytes.Buffer{}

	restore := exec.Command(
		"mysql",
		"-u", tUser,
		"-h", tHost,
		"-P", fmt.Sprint(tPort),
		"-D", tName,
	)
	restore.Stdin = stdin
	restore.Stderr = stderr

	if err := restore.Start(); err != nil {
		fmt.Printf("start mysql command: %s\n", err)
		os.Exit(40)
	}

	if err := restore.Wait(); err != nil {
		fmt.Printf("wait for mysql command: %s\n\n%s", err, stderr.String())
		os.Exit(50)
	}
}

func TestMain(m *testing.M) {
	driver := sqle.NewDefault()
	driver.AddDatabase(createTestDB())

	config := server.Config{
		Protocol: "tcp",
		Address:  fmt.Sprintf("%s:%d", tHost, tPort),
		Auth:     auth.NewNativeSingle(tUser, "", auth.AllPermissions),
	}

	s, err := server.NewDefaultServer(config, driver)
	if err != nil {
		fmt.Printf("NewDefaultServer returned non-nil error: %v\n", err)
		os.Exit(10)
	}

	defer s.Close()

	// MySQL Server
	go func(s *server.Server) {
		if err := s.Start(); err != nil {
			fmt.Printf("mysql server returned non-nil error: %v\n", err)
			os.Exit(20)
		}
	}(s)

	time.Sleep(time.Second * 3)

	createTestTable()

	os.Exit(m.Run())
}
