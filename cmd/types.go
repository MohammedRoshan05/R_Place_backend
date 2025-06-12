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
 
// we use the websocekt connection to get username so we dont need that field. 
// we can use go lang itself to get the time of request, no need to send it from frontend as it can be altered.
type PlaceTileReq struct {
    TileNo int    `json:"tile_no"`
    Colour string `json:"colour"`
}


// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan UpdateTileReq

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	//Rate limitting fields
	lastUpdate	map[int]time.Time
	mu			sync.Mutex
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan UpdateTileReq

	//Account
	account *Account
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
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

