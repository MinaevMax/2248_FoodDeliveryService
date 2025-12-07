package jwt

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Секретный ключ для подписи токенов. Это значение следует хранить в безопасном месте.
var jwtSecret = []byte("your-secret-key") // Не забудьте заменить это на более безопасный секретный ключ

// Структура для полезных данных, которые будут храниться в JWT
type Claims struct {
	UserID string `json:"userID"`
	Email  string `json:"email"`
	jwt.StandardClaims
}

// Функция для генерации JWT токена
func GenerateToken(userID, email string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 1).Unix(), // Токен будет действовать 1 час
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// Функция для валидации и парсинга JWT токена
func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
