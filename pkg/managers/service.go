package managers

import (
	"context"
	"log"
	"strconv"

	"golang.org/x/crypto/bcrypt"

	"github.com/KarrenAeris/crud/pkg/types"
	"github.com/KarrenAeris/crud/pkg/utils"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

//Service ...
type Service struct {
	pool *pgxpool.Pool
}

//NewService ...
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

//IDByToken ...
func (s *Service) IDByToken(ctx context.Context, token string) (int64, error) {
	var id int64
	sqlStatement := `select manager_id from managers_tokens where token = $1`

	err := s.pool.QueryRow(ctx, sqlStatement, token).Scan(&id)

	if err != nil {
		log.Print(err)
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, nil
	}

	return id, nil
}

//IsAdmin ...
func (s *Service) IsAdmin(ctx context.Context, id int64) (isAdmin bool) {
	sqlStmt := `select is_admin from managers  where id = $1`
	err := s.pool.QueryRow(ctx, sqlStmt, id).Scan(&isAdmin)
	if err != nil {
		return false
	}
	return
}

//Create ...
func (s *Service) Create(ctx context.Context, item *types.Manager) (string, error) {
	var token string
	var id int64

	sqlStmt := `insert into managers(name,phone,is_admin) values ($1,$2,$3) on conflict (phone) do nothing returning id;`
	err := s.pool.QueryRow(ctx, sqlStmt, item.Name, item.Phone, item.IsAdmin).Scan(&id)
	if err != nil {
		log.Print(err)
		return "", types.ErrInternal
	}

	token, err = utils.GenerateTokenStr()
	if err != nil {
		return "", err
	}

	_, err = s.pool.Exec(ctx, `insert into managers_tokens(token,manager_id) values($1,$2)`, token, id)
	if err != nil {
		return "", types.ErrInternal
	}

	return token, nil
}

//Token ...
func (s *Service) Token(ctx context.Context, phone, password string) (token string, err error) {
	var hash string
	var id int64
	err = s.pool.QueryRow(ctx, `select id,password from managers where phone = $1`, phone).Scan(&id, &hash)
	log.Println(err)
	if err == pgx.ErrNoRows {
		return "", types.ErrInvalidPassword
	}
	if err != nil {
		return "", types.ErrInternal
	}

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	log.Println(err)
	if err != nil {
		return "", types.ErrInvalidPassword
	}

	token, err = utils.GenerateTokenStr()
	log.Println(err)
	if err != nil {
		return "", err
	}

	_, err = s.pool.Exec(ctx, `insert into managers_tokens(token,manager_id) values($1,$2)`, token, id)
	log.Println(err)
	if err != nil {
		return "", types.ErrInternal
	}

	return token, nil
}

//SaveProduct ...
func (s *Service) SaveProduct(ctx context.Context, product *types.Product) (*types.Product, error) {

	var err error

	if product.ID == 0 {
		sqlstmt := `insert into products(name,qty,price) values ($1,$2,$3) returning id,name,qty,price,active,created;`
		err = s.pool.QueryRow(ctx, sqlstmt, product.Name, product.Qty, product.Price).
			Scan(&product.ID, &product.Name, &product.Qty, &product.Price, &product.Active, &product.Created)
	} else {
		sqlstmt := `update  products set  name=$1, qty=$2,price=$3  where id = $4 returning id,name,qty,price,active,created;`
		err = s.pool.QueryRow(ctx, sqlstmt, product.Name, product.Qty, product.Price, product.ID).
			Scan(&product.ID, &product.Name, &product.Qty, &product.Price, &product.Active, &product.Created)
	}

	if err != nil {
		log.Print(err)
		return nil, types.ErrInternal
	}
	return product, nil
}

//MakeSalePosition ...
func (s *Service) MakeSalePosition(ctx context.Context, position *types.SalePosition) bool {
	active := false
	qty := 0
	if err := s.pool.QueryRow(ctx, `select qty,active from products where id = $1`, position.ProductID).
		Scan(&qty, &active); err != nil {
		return false
	}
	if qty < position.Qty || !active {
		return false
	}
	if _, err := s.pool.Exec(ctx, `update products set qty = $1 where id = $2`, qty-position.Qty, position.ProductID); err != nil {
		log.Print(err)
		return false
	}
	return true
}

//MakeSale ...
func (s *Service) MakeSale(ctx context.Context, sale *types.Sale) (*types.Sale, error) {

	positionsSQLstmt := "insert into sales_positions (sale_id,product_id,qty,price) values "

	sqlstmt := `insert into sales(manager_id,customer_id) values ($1,$2) returning id, created;`

	err := s.pool.QueryRow(ctx, sqlstmt, sale.ManagerID, sale.CustomerID).Scan(&sale.ID, &sale.Created)
	if err != nil {
		log.Print(err)
		return nil, types.ErrInternal
	}
	for _, position := range sale.Positions {
		if !s.MakeSalePosition(ctx, position) {
			log.Print("Invalid position")
			return nil, types.ErrInternal
		}
		positionsSQLstmt += "(" + strconv.FormatInt(sale.ID, 10) + "," + strconv.FormatInt(position.ProductID, 10) + "," + strconv.Itoa(position.Price) + "," + strconv.Itoa(position.Qty) + "),"
	}

	positionsSQLstmt = positionsSQLstmt[0 : len(positionsSQLstmt)-1]

	log.Print(positionsSQLstmt)
	_, err = s.pool.Exec(ctx, positionsSQLstmt)
	if err != nil {
		log.Print(err)
		return nil, types.ErrInternal
	}

	return sale, nil
}

//GetSales ...
func (s *Service) GetSales(ctx context.Context, id int64) (sum int, err error) {

	sqlstmt := `
	select coalesce(sum(sp.qty * sp.price),0) total
	from managers m
	left join sales s on s.manager_id= $1
	left join sales_positions sp on sp.sale_id = s.id
	group by m.id
	limit 1`

	err = s.pool.QueryRow(ctx, sqlstmt, id).Scan(&sum)
	if err != nil {
		log.Print(err)
		return 0, types.ErrInternal
	}
	return sum, nil
}

//Products ...
func (s *Service) Products(ctx context.Context) ([]*types.Product, error) {

	items := make([]*types.Product, 0)

	sqlstmt := `select id, name, price, qty from products where active = true order by id limit 500`
	rows, err := s.pool.Query(ctx, sqlstmt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return items, nil
		}
		return nil, types.ErrInternal
	}
	defer rows.Close()

	for rows.Next() {
		item := &types.Product{}
		err = rows.Scan(&item.ID, &item.Name, &item.Price, &item.Qty)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

//RemoveProductByID ...
func (s *Service) RemoveProductByID(ctx context.Context, id int64) (err error) {

	_, err = s.pool.Exec(ctx, `delete from products where id = $1`, id)
	if err != nil {
		log.Print(err)
		return types.ErrInternal
	}
	return nil
}

//RemoveCustomerByID ...
func (s *Service) RemoveCustomerByID(ctx context.Context, id int64) (err error) {

	_, err = s.pool.Exec(ctx, `DELETE from customers where id = $1`, id)
	if err != nil {
		log.Print(err)
		return types.ErrInternal
	}
	return nil
}

//Customers ...
func (s *Service) Customers(ctx context.Context) ([]*types.Customer, error) {

	items := make([]*types.Customer, 0)
	sqlstmt := `select id, name, phone, active, created from customers where active = true order by id limit 500`
	rows, err := s.pool.Query(ctx, sqlstmt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return items, nil
		}
		return nil, types.ErrInternal
	}
	defer rows.Close()

	for rows.Next() {
		item := &types.Customer{}
		err = rows.Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
		if err != nil {
			log.Print(err)
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

//ChangeCustomer ...
func (s *Service) ChangeCustomer(ctx context.Context, customer *types.Customer) (*types.Customer, error) {

	sqlstmt := `update customers set name = $2, phone = $3, active = $4  where id = $1 returning name,phone,active`

	if err := s.pool.QueryRow(ctx, sqlstmt, customer.ID, customer.Name, customer.Phone, customer.Active).
		Scan(&customer.Name, &customer.Phone, &customer.Active); err != nil {
		log.Print(err)
		return nil, types.ErrInternal
	}

	return customer, nil
}
