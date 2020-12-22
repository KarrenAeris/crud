package customers

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

//Customer представляет информацию о покупателе.
type Customer struct {
	ID       int64     `json:"id"`
	Name     string    `json:"name"`
	Phone    string    `json:"phone"`
	Password string    `json:"password"`
	Active   bool      `json:"active"`
	Created  time.Time `json:"created"`
}

//Product представляет информацию о покупке.
type Product struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
	Qty   int    `json:"qty"`
}

//All ....
func (s *Service) All(ctx context.Context) (cs []*Customer, err error) {

	sqlStatement := `SELECT * FROM customers`

	rows, err := s.pool.Query(ctx, sqlStatement)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		item := &Customer{}
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Phone,
			&item.Active,
			&item.Created,
		)
		if err != nil {
			log.Println(err)
		}

		cs = append(cs, item)
	}

	return cs, nil
}

//AllActive ....
func (s *Service) AllActive(ctx context.Context) (cs []*Customer, err error) {

	sqlStatement := `SELECT * FROM customers WHERE active`

	rows, err := s.pool.Query(ctx, sqlStatement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item := &Customer{}
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Phone,
			&item.Active,
			&item.Created,
		)
		if err != nil {
			log.Println(err)
		}
		cs = append(cs, item)
	}

	return cs, nil
}

//ByID ...
func (s *Service) ByID(ctx context.Context, id int64) (*Customer, error) {
	item := &Customer{}

	sqlStatement := `SELECT id, name, phone, active, created FROM customers WHERE id = $1`
	err := s.pool.QueryRow(ctx, sqlStatement, id).Scan(
		&item.ID,
		&item.Name,
		&item.Phone,
		&item.Active,
		&item.Created,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	return item, nil
}

//ChangeActive ...
func (s *Service) ChangeActive(ctx context.Context, id int64, active bool) (*Customer, error) {
	item := &Customer{}

	sqlStatement := `UPDATE customers SET active = $2 where id = $1 RETURNING *`
	err := s.pool.QueryRow(ctx, sqlStatement, id, active).Scan(
		&item.ID,
		&item.Name,
		&item.Phone,
		&item.Active,
		&item.Created,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}
	return item, nil
}

//Delete ...
func (s *Service) Delete(ctx context.Context, id int64) (*Customer, error) {
	item := &Customer{}

	sqlStatement := `DELETE FROM customers WHERE id = $1 RETURNING *`
	err := s.pool.QueryRow(ctx, sqlStatement, id).Scan(
		&item.ID,
		&item.Name,
		&item.Phone,
		&item.Active,
		&item.Created,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	return item, nil
}

//Save ...
func (s *Service) Save(ctx context.Context, customer *Customer) (c *Customer, err error) {

	item := &Customer{}

	if customer.ID == 0 {
		sqlStatement := `INSERT INTO customers(name, phone, password) VALUES($1, $2, $3) RETURNING *`
		err = s.pool.QueryRow(ctx, sqlStatement, customer.Name, customer.Phone, customer.Password).Scan(
			&item.ID,
			&item.Name,
			&item.Phone,
			&item.Password,
			&item.Active,
			&item.Created,
		)
	} else {
		sqlStatement := `UPDATE customers SET name = $1, phone = $2, password = $3 WHERE id = $4 RETURNING *`
		err = s.pool.QueryRow(ctx, sqlStatement, customer.Name, customer.Phone, customer.Password, customer.ID).Scan(
			&item.ID,
			&item.Name,
			&item.Phone,
			&item.Password,
			&item.Active,
			&item.Created,
		)
	}

	if err != nil {
		log.Print(err)
		return nil, ErrInternal
	}

	return item, nil
}

//Token .... метод для генерации токена
func (s *Service) Token(ctx context.Context, phone, password string) (string, error) {

	var hash string
	var id int64

	err := s.pool.QueryRow(ctx, "select id, password from customers where phone = $1", phone).Scan(&id, &hash)
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

	//генерируем токен
	buffer := make([]byte, 256)
	n, err := rand.Read(buffer)
	if n != len(buffer) || err != nil {
		return "", ErrInternal
	}

	token := hex.EncodeToString(buffer)
	_, err = s.pool.Exec(ctx, "insert into customers_tokens(token, customer_id) values($1, $2)", token, id)
	if err != nil {
		return "", ErrInternal
	}

	return token, nil
}

//Products ...
func (s *Service) Products(ctx context.Context) ([]*Product, error) {

	items := make([]*Product, 0)

	sqlStatement := `SELECT id, name, price, qty FROM products WHERE active = true ORDER by id LIMIT 500`
	rows, err := s.pool.Query(ctx, sqlStatement)

	if err != nil {
		if err == pgx.ErrNoRows {
			return items, nil
		}
		return nil, ErrInternal
	}

	defer rows.Close()

	for rows.Next() {
		item := &Product{}
		err = rows.Scan(&item.ID, &item.Name, &item.Price, &item.Qty)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

//IDByToken ....
func (s *Service) IDByToken(ctx context.Context, token string) (int64, error) {
	var id int64
	sqlStatement := `SELECT customer_id FROM customers_tokens WHERE token = $1`
	err := s.pool.QueryRow(ctx, sqlStatement, token).Scan(&id)

	if err != nil {

		if err == pgx.ErrNoRows {
			return 0, nil
		}

		return 0, ErrInternal
	}

	return id, nil
}