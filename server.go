package main

import (
	"database/sql"
	"fmt"
	lib "key-value-server/library"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func main() {

	var err error

	config := lib.NewConfig()
	if !config.IsConfigValid {
		log.Printf("Unable to get environment variables and setup. Aborting operation")
	} else {
		store := lib.NewKVStore(lib.NewConfig().CacheSize, lib.WriteThrough)
		connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", config.User, config.Password, config.Host, config.Port, config.DatabaseName)
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

			serverURL := fmt.Sprintf("%s:%s", config.ServerHost, config.ServerPort)
			fmt.Printf("Server running on http://%s:%s", config.ServerHost, config.ServerPort)
			log.Fatal(http.ListenAndServe(serverURL, nil))
		}
	}
}
