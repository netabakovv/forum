package logger

// Logger определяет интерфейс для логирования
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
}

// Field представляет поле с дополнительной информацией для лога
type Field struct {
	Key   string
	Value interface{}
}

// NewField создает новое поле для лога
func NewField(key string, value interface{}) Field {
	return Field{
		Key:   key,
		Value: value,
	}
}
