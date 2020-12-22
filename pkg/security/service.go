package security

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var (
	//ErrNotFound возвращается, когда покупатель не найден.
	ErrNotFound = errors.New("item not found")

	//ErrInternal возвращается, когда произошла внутернняя ошибка.
	ErrInternal = errors.New("internal error")

	//ErrNoSuchUser возвращается, когда пользователь не найден
	ErrNoSuchUser = errors.New("no such user")

	//ErrInvalidPassword возвращается, когда пороль не верен
	ErrInvalidPassword = errors.New("invalid password")

	//ErrExpireToken возвращается, когда время ожидания токена истекает
	ErrExpireToken = errors.New("token expired")
)

//Service описывает сервис работы с покупателям.
type Service struct {
	pool *pgxpool.Pool
}

//NewService создаёт сервис.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// Auth - sign in
func (s *Service) Auth(login string, password string) bool {
	query := `SELECT login, password FROM managers WHERE login = $1 AND password = $2`

	err := s.pool.QueryRow(context.Background(), query, login, password).Scan(&login, &password)
	if err != nil {
		log.Print(err)
		return false
	}

	return true
}

//TokenForCustomer ....
func (s *Service) TokenForCustomer(ctx context.Context, phone string, password string) (token string, err error) {
	var hash string
	var id int64

	err = s.pool.QueryRow(ctx, `SELECT id, password FROM customers WHERE phone = $1`, phone).Scan(&id, &hash)
	if err == pgx.ErrNoRows {
		return "", ErrNoSuchUser
	}
	if err != nil {
		return "", ErrInternal
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return "", ErrInvalidPassword
	}

	buffer := make([]byte, 256)

	n, err := rand.Read(buffer)

	if n != len(buffer) || err != nil {
		return "", ErrInternal
	}

	token = hex.EncodeToString(buffer)
	_, err = s.pool.Exec(ctx, `INSERT INTO customers_tokens(token,customer_id) VALUES($1, $2)`, token, id)
	if err != nil {
		return "", ErrInternal
	}

	return token, nil
}

//AuthenticateCustomer ...
func (s *Service) AuthenticateCustomer(ctx context.Context, token string) (id int64, err error) {
	var expire time.Time

	err = s.pool.QueryRow(ctx, `SELECT customer_id, expire FROM customers_tokens WHERE token = $1`, token).Scan(&id, &expire)
	if err == pgx.ErrNoRows {
		return 0, ErrNoSuchUser
	}
	if err != nil {
		return 0, ErrInternal
	}

	tNow := time.Now().Unix()
	tEnd := expire.Unix()

	if tNow > tEnd {
		return 0, ErrExpireToken
	}

	return id, nil
}
