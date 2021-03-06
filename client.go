package noeq

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"log"
)

var (
	ErrNoAddrs      = errors.New("noeq: no addresses provided")
	ErrInvalidToken = errors.New("noeq: token > 255 bytes in length")
)

type Client struct {
	mu    sync.Mutex
	cn    net.Conn
	addrs []string
	token string
}

func New(token string, addrs ...string) (*Client, error) {
	if len(addrs) == 0 {
		return nil, ErrNoAddrs
	}

	if len(token) > 255 {
		return nil, ErrInvalidToken
	}

	return &Client{token: token, addrs: addrs}, nil
}

func (c *Client) connect() (err error) {
	n := rand.Intn(len(c.addrs))
	c.cn, err = net.Dial("tcp", c.addrs[n])
	if err != nil {
		return
	}

	return c.auth()
}

func (c *Client) auth() (err error) {
	if c.token != "" {
		_, err = fmt.Fprintf(c.cn, "\000%c%s", len(c.token), c.token)
		return
	}
	return
}

func (c *Client) Gen(n uint8) (ids []uint64, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	defer func() {
		if err != nil {
			c.cn = nil
		}
	}()

	if c.cn == nil {
		err = c.connect()
		if err != nil {
			return
		}
	}

	_, err = c.cn.Write([]byte{n})
	if err != nil {
		return
	}

	ids = make([]uint64, n)
	err = binary.Read(c.cn, binary.BigEndian, &ids)
	return
}

func (c *Client) GenOne() (uint64, error) {
	ids, err := c.Gen(1)
	if len(ids) == 0 {
		return 0, err
	}
	return ids[0], nil
}


// The client attempts to give the user as much insight as possible
// It will *not* automaticly attempt a reconnect
// on error. However, it will attempt a reconnect if you ask for
// another id after an error has occured. This gives the user more
// control over how/when to attempt a reconnect.
// This function will automatically attempt a reconnect after sleeping 
// for 1 ms. It will attempt reconnection 3 times
func (c *Client) ReallyGenOne() (id uint64, err error) {
    for trys := 3; trys > 0; trys-- {
        id, err = c.GenOne()
        if err != nil {
            log.Println("noeq: Failed to GenOne, retrying after 1 ms sleep; error:", err)
            continue
        }

        return
    }

    return
}
