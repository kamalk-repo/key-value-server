package library

// Status code list for http handler response
type StatusCode int64

const (
	Unknown StatusCode = iota - 1
	Success
	DBConnectionError
	StatementPreparationError
	InvalidJSONError
	JSONParseError
	StatementExecutionError
	DBResponseError
	InvalidKeyError
	KeyNotFoundError
	KeyDuplicationError
	InsertError
	DeleteError
)
