package main

import (
	"database/sql"
	"fmt"
	lib "key-value-server/library"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {

	var err error

	err = godotenv.Load()
	if err != nil {
		log.Printf("Unable to get environment variables and setup. Aborting operation")
	} else {
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		passowrd := os.Getenv("DB_PASSWORD")
		database := os.Getenv("DB_NAME")
		cacheSize, err := strconv.Atoi(os.Getenv("CACHE_SIZE"))

		if err != nil || cacheSize <= 0 {
			log.Printf("Invalid cache size value. Using default value of 10 for cache size")
			cacheSize = 10
		}

		store := lib.NewKVStore(cacheSize, lib.WriteThrough)
		connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, passowrd, host, port, database)
		store.DB, err = sql.Open("mysql", connectionString) // "application:root@tcp(localhost:3306)/kv_store"
		if err != nil {
			store.IsDbConnectionFailed = true
		} else {
			store.IsDbConnectionFailed = false
		}
		defer store.DB.Close()

		if err != nil {
			log.Printf("Unable to establish database connection. Aborting operation. \n Error: %s", err)
		} else {
			http.HandleFunc("/kvstore", func(w http.ResponseWriter, req *http.Request) {
				switch req.Method {
				case http.MethodGet:
					store.ReadHandler(w, req)
				case http.MethodPost:
					store.CreateHandler(w, req)
				case http.MethodPut:
					store.UpdateHandler(w, req)
				case http.MethodDelete:
					store.DeleteHandler(w, req)
				default:
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			})

			fmt.Println("Server running on http://localhost:8080")
			log.Fatal(http.ListenAndServe(":8080", nil))
		}
	}
}
