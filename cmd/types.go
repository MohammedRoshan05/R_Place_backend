package cmd

import "time"

type Account struct{
	ID				int				`json:"id"`
	Email 			string			`json:"email"`
	Password		string			`json:"-"`
	Username		string			`json:"username"`
	CreatedAt		time.Time		`json:"created_At"`
}

type CreateAccReq struct{
	Username		string		`json:"username"`
	Email 			string		`json:"email"`
	Password		string		`json:"password"`
}

type LoginAccReq struct {
	Email 			string		`json:"email"`
	Password		string		`json:"password"`
}

type UpdateTileReq struct {
	TileNo			int			`json:"tile_no"`
	UpdateTime		string		`json:"update_time"`
	Username		string		`json:"username"`
	Colour			string		`json:"colour"`
}

