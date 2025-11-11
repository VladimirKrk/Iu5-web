package api_types

// ErrorResponse стандартная структура для ответа с ошибкой
type ErrorResponse struct {
	Error string `json:"error"`
}

// TokenResponse стандартная структура для ответа с JWT токеном
type TokenResponse struct {
	Token string `json:"token"`
}

// StatusResponse стандартная структура для ответа со статусом
type StatusResponse struct {
	Status string `json:"status"`
}
