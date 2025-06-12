package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan UpdateTileReq),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		lastUpdate: make(map[int]time.Time),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

//ok so in readpump we read the rquests and validate it. in writepump all we do is push all messages from readpump of one user to writepump of every user ?

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump(s Storage) {
	fmt.Println("Starting readpump")
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	tiles,err := s.getGrid()
	if(err != nil) {
		return
	}
	updates := tilesToUpdates(tiles)

	for i := range updates {
		c.send <- updates[i]
	}

	for {
		var req PlaceTileReq
		if err := c.conn.ReadJSON(&req); err != nil {
			fmt.Println("Whats happening")
			fmt.Println(err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		fmt.Println("Received message from channel")
		update := UpdateTileReq{
			Username: c.account.Username,
			TileNo: req.TileNo,
			Colour: req.Colour,
			UpdateTime: time.Now().UTC(),
		}

		now := time.Now().UTC()
		c.hub.mu.Lock()
		last,exists := c.hub.lastUpdate[req.TileNo]
		if !exists || now.Sub(last) >= time.Second {
			// OK to proceed
            c.hub.lastUpdate[req.TileNo] = now
            c.hub.mu.Unlock()
			s.UpdateTile(&update)
			c.hub.broadcast <- update
		}else {
            // too soon, drop it
            c.hub.mu.Unlock()
        }

	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	fmt.Println("Starting writepump")

	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteJSON(message); err != nil {
				fmt.Println("Failed to write into channel")
				return  
			}

			fmt.Println("Successfully wrote mssage to channel")

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}


