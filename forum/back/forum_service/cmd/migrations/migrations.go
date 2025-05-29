package main

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	dsn := "postgres://postgres:1@localhost:5555/forum?sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе: %v", err)
	}
	defer db.Close()

	migrationsDir := "./migrations/migrations/up"

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("Ошибка чтения директории миграций: %v", err)
	}

	var sqlFiles []fs.DirEntry
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file)
		}
	}

	sort.Slice(sqlFiles, func(i, j int) bool {
		return sqlFiles[i].Name() < sqlFiles[j].Name()
	})

	for _, file := range sqlFiles {
		path := filepath.Join(migrationsDir, file.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("Ошибка чтения %s: %v", path, err)
		}

		fmt.Printf("Выполнение миграции: %s\n", file.Name())
		if _, err := db.Exec(string(content)); err != nil {
			log.Fatalf("Ошибка выполнения миграции %s: %v", file.Name(), err)
		}
	}

	fmt.Println("Все миграции успешно применены.")
}
