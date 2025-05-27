package cmd

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type PostgresDB struct{
	db *sql.DB
}

type Storage interface{
	CreateAccount(*Account) error
	DeleteAccount(string) error
	GetAccounts() ([]* Account,error)
	GetAccountByName(string) (*Account,error)
	Login(LoginAccReq) (error)
	UpdateTile(string,string,int) (error)
}

func ConnectDB()(*PostgresDB,error){
	connstr := "user=postgres dbname=postgres password=dplace sslmode=disable"
	db,err := sql.Open("postgres",connstr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresDB{
		db: db,
	},nil
}

func (s *PostgresDB) Init() error{
	if err := s.createAccountTable(); err != nil {
		return err
	}
	if err := s.createGridTable(); err != nil {
		return err
	}
	return nil
}

func (s *PostgresDB) createAccountTable() error {
	query := `create table if not exists Account(
		id serial primary key,
		Username varchar(250),
		Email varchar(250),
		Password varchar(1550),
		Created_At timestamp
	)`
	_,err := s.db.Exec(query)
	return err
}

func (s *PostgresDB) createGridTable() error {
	query := `create table if not exists Grid(
		id serial primary key,
		Username varchar(50),
		Colour varchar(50),
		LastUpdation timestamp
	)`
	_,err := s.db.Exec(query)
	return err
}

func (s *PostgresDB)CreateAccount(acc *Account) error {
	query := `insert into Account
	(Username,Email,Password,Created_At)
	values ($1,$2,$3,$4)
	`
	fmt.Println("Reach 4")

	_,err := s.db.Exec(query,
		acc.Username,
		acc.Email,
		acc.Password,
		acc.CreatedAt)
	if (err != nil){
		return err
	}
	fmt.Println("Username length:", len(acc.Email))
	fmt.Println("Email length:", len(acc.Password))
	fmt.Println("Password length:", len(acc.Username))
	return nil
}

func (s *PostgresDB) DeleteAccount(username string) error {
	query := `delete from account where Username = $1`
	_,err := s.db.Query(query,username)
	return err
}

func (s *PostgresDB) GetAccountByName(username string) (*Account,error) {
	rows, err := s.db.Query("select * from account where Username = $1", username)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("Account with username %v is not found in db", username)
}

func (s *PostgresDB) GetAccounts() ([]* Account,error) {
	rows, err := s.db.Query("select * from account")
	if err != nil {
		return nil, err
	}
	accounts := []*Account{}
	for rows.Next() {
		account, err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := new(Account)
	err := rows.Scan(&account.ID,
		&account.Username,
		&account.Email,
		&account.CreatedAt)
	return account, err

}

func (s *PostgresDB ) Login(req LoginAccReq) (error) {
	row := s.db.QueryRow(
        `SELECT password
           FROM account
          WHERE email = $1`,
        req.Email,
    )
	dbPassword := ""
	if err := row.Scan(&dbPassword); err != nil {
		return err
	}
	
	if err := bcrypt.CompareHashAndPassword([]byte(dbPassword),[]byte(req.Password)); err != nil {
		return err
	}
	return nil
}

func (s *PostgresDB) UpdateTile(username,colour string,tileID int) (error){
	query := `update Grid set Colour = $1, LastUpdation = NOW()
		where id = $2 AND Username = $3`
	res,err := s.db.Exec(query,colour,tileID,username)
	
	n, err := res.RowsAffected()
    if err != nil {
        return err
    }
    if n == 0 {
        return fmt.Errorf("no grid cell found for id=%d and user=%q", tileID, username)
    }
    return nil
}

func NewAccount( Email,Password,Username string) (*Account, error) {
	encPW, err := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &Account{
		Email: Email,
		Password:  string(encPW),
		Username: Username,
		CreatedAt: time.Now().UTC(),
	}, nil
}