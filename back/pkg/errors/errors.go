package errors

import "errors"

var (
	ErrEmptyTargetID     = errors.New("targetID не может быть пустым")
	ErrInvalidVote       = errors.New("некорректные данные для голосования")
	ErrEmptyMessage      = errors.New("пустое сообщение")
	ErrMessageTooLong    = errors.New("сообщение слишком длинное")
	ErrEmptyComment      = errors.New("пустой комментарий")
	ErrCleanupOldMessage = errors.New("ошибка очистки старых сообщений")
	ErrRegister          = errors.New("ошибка регистрации")
	ErrNotFound          = errors.New("пользователь не найден")

	// Ошибки репозитория
	ErrUserNotFound      = errors.New("пользователь не найден")
	ErrDuplicateUsername = errors.New("пользователь с таким именем уже существует")
	ErrTokenNotFound     = errors.New("токен обновления не найден")
	ErrTokenExpired      = errors.New("срок действия токена истек")
	ErrCommentNotFound   = errors.New("комментарий не найден")

	// Ошибки аутентификации
	ErrInvalidCredentials = errors.New("неверные учетные данные")
	ErrWrongPassword      = errors.New("неверный пароль")
	ErrTokenInvalid       = errors.New("недействительный токен")

	// Ошибки авторизации
	ErrNotAuthorized    = errors.New("не авторизован")
	ErrPermissionDenied = errors.New("отказано в доступе")
	ErrTokenRevoked     = errors.New("токен был отозван")

	// Ошибки валидации
	ErrUsernameTooShort = errors.New("слишком маленькое имя")
	ErrUsernameTooLong  = errors.New("слишком длинное имя")
	ErrEmptyUsername    = errors.New("имя пользователя не может быть пустым")
	ErrEmptyPassword    = errors.New("пароль не может быть пустым")
	ErrInvalidUsername  = errors.New("некорректный формат имени пользователя")
	ErrWeakPassword     = errors.New("слишком слабый пароль")

	// Ошибки базы данных
	ErrDB                = errors.New("ошибка бд")
	ErrDBConnection      = errors.New("ошибка подключения к базе данных")
	ErrDBQuery           = errors.New("ошибка запроса к базе данных")
	ErrTransactionFailed = errors.New("ошибка транзакции")
	ErrDeleteFailed      = errors.New("ошибка удаления")
)
