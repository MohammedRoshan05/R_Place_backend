package cmd

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Account struct{
	ID				int				`json:"id"`
	Email 			string			`json:"email"`
	Password		string			`json:"-"`
	Username		string			`json:"username"`
	CreatedAt		time.Time		`json:"created_At"`
}

type ctxKey string

const accountKey ctxKey = "account"

type CreateAccReq struct{
	Username		string		`json:"username"`
	Email 			string		`json:"email"`
	Password		string		`json:"password"`
}

type LoginAccReq struct {
	Email 			string		`json:"email"`
	Password		string		`json:"password"`
}

type Tile struct {
	TileNo			int			`json:"tile_no"`
	Username		string		`json:"username"`
	Colour			string		`json:"colour"`
}
type UpdateTileReq struct {
	TileNo			int			`json:"tile_no"`
	UpdateTime		time.Time	`json:"update_time"`
	Username		string		`json:"username"`
	Colour			string		`json:"colour"`
}

func tilesToUpdates(tiles []*Tile) []UpdateTileReq {
    updates := make([]UpdateTileReq, 0, len(tiles))
    for _, t := range tiles {
        updates = append(updates, UpdateTileReq{
            TileNo:     t.TileNo,
            Username:   t.Username,
            Colour:     t.Colour,
            UpdateTime: time.Now().UTC(),
        })
    }
    return updates
}

type PlaceTileReq struct {
    TileNo int    `json:"tile_no"`
    Colour string `json:"colour"`
}


type Hub struct {
	clients map[*Client]bool

	broadcast chan UpdateTileReq

	register chan *Client

	unregister chan *Client

	lastUpdate	map[int]time.Time
	mu			sync.Mutex
}

type Client struct {
	hub *Hub

	conn *websocket.Conn

	send chan UpdateTileReq

	account *Account
}

const (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = (pongWait * 9) / 10

	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

