package library

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// --------------- KEY VALUE STORE CODE: START ---------------

// Repository of reader friendly messages corresponding to status codes for response
func (kv KVStore) getErrorDescription(status StatusCode) string {
	var response string
	switch status {
	case Unknown:
		response = "Unknown error"
	case Success:
		response = "No error"
	case DBConnectionError:
		response = "Database connection failed. Request Aborted"
	case StatementPreparationError:
		response = "unable to prepare statement for DB query"
	case InvalidJSONError:
		response = "Invalid JSON"
	case JSONParseError:
		response = "Unable to parse server response to JSON"
	case KeyDuplicationError:
		response = "Key already present"
	case StatementExecutionError:
		response = "unable to execute database query"
	case DBResponseError:
		response = "unable to get DB response for query execution"
	case InsertError:
		response = "unable to add given key in DB"
	case InvalidKeyError:
		response = "Invalid key"
	case KeyNotFoundError:
		response = "Key not found"
	}
	return response
}

// Cache mode enum
type CacheMode int64

const (
	WriteThrough = iota
	WriteBack
)

// Request and response object for http handler functions
type KVData struct {
	Key   int    `json:"Key"`
	Value string `json:"Value"`
}
type KVResponse struct {
	Status StatusCode  `json:"Status"`
	Error  string      `json:"Error"`
	Data   interface{} `json:"Data"`
}

// Key value entity to manage key store. Can be configured for database and cache operation
type KVStore struct {
	//cache                *Cache
	cache                *LRUCache
	mode                 CacheMode
	DB                   *sql.DB
	IsDbConnectionFailed bool
}

// Utility function to initialize key value store with default values
func NewKVStore(cacheLength int, modeOfCache CacheMode) *KVStore {
	return &KVStore{
		cache:                NewLRUCache(cacheLength),
		IsDbConnectionFailed: false,
		mode:                 modeOfCache,
	}
}

// Utitlity method to standardize response across http handler
// Standard response fields: Status, Error, Data
// Data field provides flexibility to http handler to provide custom response data to client
func (kv KVStore) respond(status StatusCode, w http.ResponseWriter, data interface{}) {
	resp := KVResponse{status, kv.getErrorDescription(status), data}

	if status == Success {
		w.WriteHeader(http.StatusOK)
	} else if status == KeyNotFoundError {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(resp)
}
func (kv *KVStore) checkIfKeyExists(key int) (StatusCode, error) {

	var status StatusCode
	var err error
	if kv.IsDbConnectionFailed {
		status, err = DBConnectionError, fmt.Errorf("fatal error. Exectuion cannot continue")
	} else {
		stmt, err := kv.DB.Prepare("SELECT EXISTS(SELECT 1 FROM key_store_table WHERE key_name = ?)")
		if err != nil {
			status = StatementPreparationError
		} else {
			var exists bool
			err = stmt.QueryRow(key).Scan(&exists)
			if err != nil {
				status = DBResponseError
			} else {
				if exists == true {
					status, err = Success, nil
				} else {
					status, err = KeyNotFoundError, fmt.Errorf("unable to find key: %d", key)
				}
			}
		}
	}
	return status, err
}
func (kv *KVStore) insertKey(key int, value string) (StatusCode, error) {

	var status StatusCode
	var resError error
	status, resError = Success, nil

	if kv.IsDbConnectionFailed {
		status, resError = DBConnectionError, fmt.Errorf("fatal error. Exectuion cannot continue")
	}

	stmt, err := kv.DB.Prepare("INSERT INTO key_store_table (key_name, key_value) VALUES (?, ?)")
	if err != nil {
		status, resError = StatementPreparationError, err
	}

	dbResult, err := stmt.Exec(key, value)
	if err != nil {
		status, resError = InsertError, err
	}

	rowsAffected, err := dbResult.RowsAffected()
	if err != nil {
		status, resError = DBResponseError, fmt.Errorf("unable to verify for insert of key: %d", key)
	}

	if rowsAffected == 1 {
		status, resError = Success, nil
	} else {
		status, resError = InsertError, fmt.Errorf("unable to insert key: %d", key)
	}

	return status, resError
}
func (kv *KVStore) readKey(key int) (StatusCode, string, error) {

	var status StatusCode
	var resError error
	var value string

	if kv.IsDbConnectionFailed {
		status, value, resError = DBConnectionError, "", fmt.Errorf("fatal error. exectuion cannot continue")
	} else {
		stmt, err := kv.DB.Prepare("SELECT key_value FROM key_store_table WHERE key_name = ?")
		if err != nil {
			status, value, resError = StatementPreparationError, "", err
		} else {

			var keyValue string
			err = stmt.QueryRow(key).Scan(&keyValue)
			if err == sql.ErrNoRows {
				status, value, resError = KeyNotFoundError, "", fmt.Errorf("unable to find key: %d", key)
			} else if err != nil {
				status, value, resError = DBResponseError, "", err
			} else {
				status, value, resError = Success, keyValue, nil
			}
		}
	}
	return status, value, resError
}
func (kv *KVStore) updateKey(key int, value string) (StatusCode, error) {

	var status StatusCode
	var resError error

	if kv.IsDbConnectionFailed {
		status, resError = DBConnectionError, fmt.Errorf("fatal error. Exectuion cannot continue")
	}

	stmt, err := kv.DB.Prepare("UPDATE key_store_table SET key_value = ? WHERE key_name = ?")
	if err != nil {
		status, resError = StatementPreparationError, err
	}

	dbResult, err := stmt.Exec(value, key)
	if err != nil {
		status, resError = DBResponseError, err
	}

	rowsAffected, err := dbResult.RowsAffected()
	if err != nil {
		status, resError = DBResponseError, fmt.Errorf("unable to verify for update of key: %d", key)
	}

	if rowsAffected > 0 {
		status, resError = Success, nil
	} else {
		status, resError = DeleteError, fmt.Errorf("unable to update key: %d", key)
	}
	return status, resError
}
func (kv *KVStore) deleteKey(key int) (StatusCode, error) {
	if kv.IsDbConnectionFailed {
		return DBConnectionError, fmt.Errorf("fatal error. exectuion cannot continue")
	}

	stmt, err := kv.DB.Prepare("DELETE FROM key_store_table WHERE key_name = ?")
	if err != nil {
		return StatementPreparationError, err
	}

	dbResult, err := stmt.Exec(key)
	if err != nil {
		return DBResponseError, err
	}

	rowsAffected, err := dbResult.RowsAffected()
	if err != nil {
		return DBResponseError, fmt.Errorf("unable to verify for deletion of key: %d", key)
	}

	if rowsAffected == 1 {
		return Success, nil
	} else {
		return DeleteError, fmt.Errorf("unable to delete key: %d", key)
	}
}

// HTTP handler functions
// For POST method
func (kv *KVStore) CreateHandler(w http.ResponseWriter, r *http.Request) {

	var status StatusCode
	var data interface{}
	var resError error
	var req KVData

	if resError = json.NewDecoder(r.Body).Decode(&req); resError != nil {
		status, data = InvalidJSONError, `Required JSON format: 
											{
												"Key": int,
												"Value": string
											}`

		kv.respond(status, w, data)
		return
	}

	if _, exists := kv.cache.CheckKey(req.Key); exists {
		status, data = KeyDuplicationError, fmt.Sprintf("Key already present. Key: %d", req.Key)
	} else {

		status, _ = kv.checkIfKeyExists(req.Key)
		if status == Success {
			status, data = KeyDuplicationError, fmt.Sprintf("Key already present. Key: %d", req.Key)
		} else {
			if kv.mode == WriteThrough {
				status, resError = kv.insertKey(req.Key, req.Value)
				if status != Success {
					data = resError
				}
			}

			// In one case only when there is error while writing in DB during Write Through mode,
			// Don't write in cache
			if kv.mode != WriteThrough || (kv.mode == WriteThrough && status == Success) {
				kv.cache.Add(&Node{key: req.Key, keyValue: req.Value})
				data = fmt.Sprintf("Inserted key: %d", req.Key)
				status = Success
			}
		}
	}
	kv.respond(status, w, data)
}

// For GET method
func (kv *KVStore) ReadHandler(w http.ResponseWriter, r *http.Request) {
	keyStr := r.URL.Query().Get("key")
	key, err := strconv.Atoi(keyStr)
	if err != nil {
		kv.respond(InvalidKeyError, w, "Key need to be of integer type only")
		return
	}

	var status StatusCode
	var resError error
	var data interface{}
	var keyValue string

	if keyNode, exists := kv.cache.CheckKey(key); !exists {
		status, keyValue, resError = kv.readKey(key)
		if status == Success {
			kv.cache.Add(&Node{key: key, keyValue: keyValue})
			data = KVData{Key: key, Value: keyValue}
		} else if status == KeyNotFoundError {
			data = fmt.Sprintf("")
		} else {
			data = resError
			status = Unknown
		}
	} else {
		data = KVData{Key: key, Value: keyNode.keyValue}
		status = Success
	}

	kv.respond(status, w, data)
}

// For PUT method
func (kv *KVStore) UpdateHandler(w http.ResponseWriter, r *http.Request) {

	var status StatusCode
	var data interface{}
	var existsInDB, existsInCache bool
	var req KVData

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		kv.respond(InvalidJSONError, w, `Required JSON format: 
		{
			"Key": int,
			"Value": string
		}`)
		return
	}

	if _, existsInCache = kv.cache.CheckKey(req.Key); !existsInCache {

		// If key not in cache, check in database
		status, _ = kv.checkIfKeyExists(req.Key)
		if status == Success {
			existsInDB = true
		}
	}

	if !existsInDB && !existsInCache {
		status, data = KeyNotFoundError, fmt.Sprintf("Key not found. Key: %d", req.Key)
	} else {
		if kv.mode == WriteThrough {
			var err error
			status, err = kv.updateKey(req.Key, req.Value)
			if status != Success {
				data = err
			}
		}

		// In one case only when there is error while writing in DB during Write Through mode,
		// Don't write in cache
		if kv.mode != WriteThrough || (kv.mode == WriteThrough && status == Success) {
			if !existsInCache {
				kv.cache.Add(&Node{key: req.Key, keyValue: req.Value})
			} else {
				kv.cache.UpdateKey(req.Key, req.Value)
			}
			data = fmt.Sprintf("Updated key: %d", req.Key)
		}
	}

	kv.respond(status, w, data)
}

// For DELETE method
func (kv *KVStore) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	keyStr := r.URL.Query().Get("key")
	key, err := strconv.Atoi(keyStr)
	if err != nil {
		kv.respond(InvalidKeyError, w, "Key need to be of integer type only")
		return
	}

	var existsInCache, existsInDB bool
	var keyNode *Node
	var data interface{}
	var status StatusCode

	// Present in cache, delete in cache
	if keyNode, existsInCache = kv.cache.CheckKey(key); existsInCache {
		kv.cache.Remove(keyNode)
	}

	// Present in DB, delete in DB
	if status, _ = kv.checkIfKeyExists(key); status == Success {
		existsInDB = true
		status, err := kv.deleteKey(key)
		if status != Success {
			data = err
		}
	}

	if !existsInCache && !existsInDB {
		status, data = KeyNotFoundError, fmt.Sprintf("Key not found. Key: %d", key)
	} else if existsInCache && !existsInDB {
		status, data = Success, fmt.Sprintf("Deleted key: %d", key)
	} else if existsInDB && status == Success {
		status, data = Success, fmt.Sprintf("Deleted key: %d", key)
	}

	kv.respond(status, w, data)
}

// --------------- KEY VALUE STORE CODE: END ---------------
