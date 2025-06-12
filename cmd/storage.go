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

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(string) error
	GetAccounts() ([]* Account,error)
	GetAccountByName(string) (*Account,error)
	Login(*LoginAccReq) (error)
	UpdateTile(update *UpdateTileReq) (error)
	getGrid() ([]*Tile, error)
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
		Colour varchar(50)
	)`
	_,err := s.db.Exec(query)
	if(err != nil){
		return err
	}

	var count int
    if err := s.db.QueryRow(`SELECT COUNT(*) FROM Grid`).Scan(&count); err != nil {
        return fmt.Errorf("count rows: %w", err)
    }
	if(count == 0){
		query = `insert into Grid
		(Username,Colour)
		values ($1, $2)`
		for i := 0; i < 100; i++ {
			_,err = s.db.Exec(query,
			"","black")
			if(err != nil){
				return err
			}
		}
	}
	return nil
}

func (s *PostgresDB) getGrid() ([]*Tile, error) {
	tilesDB,err := s.db.Query(`select * from Grid`)
	if(err != nil){
		return nil,err
	}
	tiles := []*Tile{}
	for tilesDB.Next(){
		tile,err := scanIntoTiles(tilesDB)
		if(err != nil){ 
			return nil,err
		}
		tiles = append(tiles,tile)
	}
	return tiles,nil
}

func (s *PostgresDB) UpdateTile(update *UpdateTileReq) (error){
	query := `UPDATE Grid 
		SET colour = $1, username = $2 
		WHERE id = $3`	
	res,err := s.db.Exec(query,update.Colour,update.Username,update.TileNo)
	n, err := res.RowsAffected()
    if err != nil {
        return err
    }
    if n == 0 {
        return fmt.Errorf("no grid cell found for id=%d and user=%q", update.TileNo, update.Username)
    }
    return nil
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

func (s *PostgresDB)CreateAccount(acc *Account) error {
	query := `insert into Account
	(Username,Email,Password,Created_At)
	values ($1,$2,$3,$4)
	`
	_,err := s.db.Exec(query,
		acc.Username,
		acc.Email,
		acc.Password,
		acc.CreatedAt)
	if (err != nil){
		return err
	}
	return nil
}

func (s *PostgresDB) DeleteAccount(username string) error {
	query := `delete from account where Username = $1`
	_,err := s.db.Query(query,username)
	return err
}

func (s *PostgresDB) GetAccountByName(email string) (*Account,error) {
	rows, err := s.db.Query("select * from account where Email = $1", email)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("Account with username %v is not found in db", email)
}


func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := new(Account)
	err := rows.Scan(&account.ID,
		&account.Username,
		&account.Email,
		&account.Password,
		&account.CreatedAt)
	return account, err
}

func scanIntoTiles(rows *sql.Rows) (*Tile, error) {
	tile := new(Tile)
	err := rows.Scan(&tile.TileNo,
		&tile.Username,
		&tile.Colour)
	return tile, err
}

func (s *PostgresDB) Login(req *LoginAccReq) (error) {
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
		fmt.Println("Raw password:", req.Password)
		fmt.Println("Hashed password from DB:", dbPassword)
		return err
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