package main

import (
	"database/sql"
	"fmt"
	lib "key-value-server/library"
	"log"
	"net/http"
	_ "net/http/pprof"

	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {

	// Courtsey: https://medium.com/@jhathnagoda/go-profiling-with-pprof-a-step-by-step-guide-a62323915cb0
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	var err error

	config := lib.NewConfig()
	if !config.IsConfigValid {
		log.Printf("Unable to get environment variables and setup. Aborting operation")
	} else {
		store := lib.NewKVStore(lib.NewConfig().CacheSize, lib.WriteThrough)
		connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", config.User, config.Password, config.Host, config.Port, config.DatabaseName)
		store.DB, err = sql.Open("mysql", connectionString) // "application:root@tcp(localhost:3306)/kv_store"
		defer store.DB.Close()
		if err != nil {
			store.IsDbConnectionFailed = true
		} else {
			store.IsDbConnectionFailed = false
			store.DB.SetMaxOpenConns(500)
			store.DB.SetMaxIdleConns(20)
			store.DB.SetConnMaxLifetime(5 * time.Minute)
			store.DB.SetConnMaxIdleTime(30 * time.Second)
		}

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

			http.HandleFunc("/hello", func(w http.ResponseWriter, req *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/text")
				w.Write([]byte("Hello  From Server"))
				return
			})

			serverURL := fmt.Sprintf("%s:%s", config.ServerHost, config.ServerPort)
			server := &http.Server{
				Addr:         serverURL,
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 2 * time.Minute,
				IdleTimeout:  30 * time.Second,
			}
			fmt.Printf("Server running on http://%s:%s", config.ServerHost, config.ServerPort)
			log.Fatal(server.ListenAndServe())
		}
	}
}
