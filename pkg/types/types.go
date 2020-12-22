package types

import (
	"errors"
	"time"
)

var (
	//ErrNotFound возвращается, когда поле не найден
	ErrNotFound = errors.New("item not found")

	//ErrInternal возвращается, когда произошла внутернняя ошибка.
	ErrInternal = errors.New("internal error")

	//ErrTokenNotFound возвращается, когда токен не найден
	ErrTokenNotFound = errors.New("token not found")

	//ErrNoSuchUser возвращается, когда пользователь не найден
	ErrNoSuchUser = errors.New("no such user")

	//ErrInvalidPassword возвращается, когда пороль не верен
	ErrInvalidPassword = errors.New("invalid password")

	//ErrPhoneUsed возвращается, когда телефон (логин) не найден
	ErrPhoneUsed = errors.New("phone alredy registered")

	//ErrExpireToken возвращается, когда время ожидания токена истекает
	ErrExpireToken = errors.New("token expired")
)

//Manager представляет информацию о продавцов.
type Manager struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Salary      int64     `json:"salary"`
	Plan        int64     `json:"plan"`
	BossID      int64     `json:"boss_id"`
	Departament string    `json:"departament"`
	Phone       string    `json:"phone"`
	Password    string    `json:"password"`
	IsAdmin     bool      `json:"is_admin"`
	Created     time.Time `json:"created"`
}

//Product представляет информацию о покупатках.
type Product struct {
	ID      int64     `json:"id"`
	Name    string    `json:"name"`
	Price   int       `json:"price"`
	Qty     int       `json:"qty"`
	Active  bool      `json:"active"`
	Created time.Time `json:"created"`
}

//Sale представляет информацию о скидках.
type Sale struct {
	ID         int64           `json:"id"`
	ManagerID  int64           `json:"manager_id"`
	CustomerID int64           `json:"customer_id"`
	Created    time.Time       `json:"created"`
	Positions  []*SalePosition `json:"positions"`
}

//SalePosition представляет информацию о позиции скидки.
type SalePosition struct {
	ID        int64     `json:"id"`
	ProductID int64     `json:"product_id"`
	SaleID    int64     `json:"sale_id"`
	Price     int       `json:"price"`
	Qty       int       `json:"qty"`
	Created   time.Time `json:"created"`
}

//Customer представляет информацию о покупателе.
type Customer struct {
	ID      int64     `json:"id"`
	Name    string    `json:"name"`
	Phone   string    `json:"phone"`
	Active  bool      `json:"active"`
	Created time.Time `json:"created"`
}
