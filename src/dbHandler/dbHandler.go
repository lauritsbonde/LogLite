package dbhandler

type DBHandler interface {
	Put(table string, data map[string]interface{}) error
	Get(table string, conditions map[string]interface{}) ([]map[string]interface{}, error)
	Close() error
}