package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var addr = flag.String("addr", ":8080", "http service address")

//var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }} // local development
var upgrader = websocket.Upgrader{}

func pty2ws(ws *websocket.Conn, ptmx *os.File) {
	buffer := make([]byte, 1024)
	var payload, overflow []byte
	for {
		n, err := ptmx.Read(buffer)
		if err != nil {
			fmt.Println("[pty2ws] read from pty error: ", err)
			return
		}

		// Empty the overflow from the last read into the payload first.
		payload = append(payload[0:], overflow...)
		overflow = nil
		// Then empty the new buf read into the payload.
		payload = append(payload, buffer[:n]...)

		// Strip out any incomplete utf-8 from current payload into overflow.
		for !utf8.Valid(payload) {
			overflow = append(overflow[:0], append(payload[len(payload)-1:], overflow[0:]...)...)
			payload = payload[:len(payload)-1]
		}

		if len(payload) >= 1 {
			err = ws.WriteMessage(websocket.BinaryMessage, payload[:len(payload)])
			if err != nil {
				fmt.Println("[pty2ws] write to ws error: ", err)
				return
			}
		}

		// Empty the payload.
		payload = nil
	}
}

func ws2pty(ws *websocket.Conn, ptmx *os.File) {
	for {
		mt, message, err := ws.ReadMessage()
		if mt == -1 || err != nil {
			log.Println("[ws2pty] ws read error: ", err)
			return
		}
		msg := string(message)
		if strings.HasPrefix(msg, "<RESIZE>") {
			size := msg[len("<RESIZE>"):len(msg)]
			sizeArr := strings.Split(size, ",")
			rows, _ := strconv.Atoi(sizeArr[0])
			cols, _ := strconv.Atoi(sizeArr[1])

			ws := new(pty.Winsize)
			ws.Cols = uint16(cols)
			ws.Rows = uint16(rows)
			err = pty.Setsize(ptmx, ws)
			log.Printf("[ws2pty] pty resize window to %d, %d", cols, rows)
			if err != nil {
				log.Println("[ws2pty] pty resize error: ", err)
				return
			}
		} else {
			_, err = ptmx.Write(message)
			if err != nil {
				log.Println("[ws2pty] pty write error: ", err)
			}
		}
	}
}

func bash(context echo.Context) error {
	// Create arbitrary command.
	t, err := context.Cookie("token")
	if t == nil || err != nil || len(t.Value) == 0 || t.Value != token {
		return context.String(http.StatusUnauthorized, "session cookie is missing")
	}

	log.Println("begin ws upgrade")
	c, err := upgrader.Upgrade(context.Response(), context.Request(), nil)
	if err != nil {
		log.Print("ws upgrade failed: ", err)
		return err
	}
	defer func() { _ = c.Close() }()

	// can be bash or cmd.exe
	console := exec.Command("sh")
	defer func() { _ = console.Wait() }()

	// Start the command with a pty.
	ptmx, err := pty.Start(console)
	if err != nil {
		log.Print("open terminal err: ", err)
		return err
	}
	defer func() { _ = ptmx.Close() }() // Best effort.

	go func() { pty2ws(c, ptmx) }()
	// block from close
	ws2pty(c, ptmx)
	log.Println("ws closed")
	return nil
}

type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

var token = uuid.New().String()

func login(context echo.Context) error {
	var credential Credentials
	err := context.Bind(&credential)
	if err != nil {
		return context.String(http.StatusBadRequest, "invalid credentials")
	}
	//TODO: hardcoded username and password
	if credential.Username != "admin" || credential.Password != "admin" {
		return context.String(http.StatusUnauthorized, "unauthorized request")
	}

	token = uuid.New().String()
	// for https add secure flag into cookie too
	cookie := new(http.Cookie)
	cookie.Name = "token"
	cookie.Value = token
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Expires = time.Now().Add(24 * time.Hour)
	context.SetCookie(cookie)
	return context.String(http.StatusOK, "{\"token\":\"session_token\"}")
}

func refresh(context echo.Context) error {
	cookie, err := context.Cookie("token")
	if err != nil || len(cookie.Value) == 0 || cookie.Value != token {
		return context.String(http.StatusUnauthorized, "invalid token")
	}

	token = uuid.New().String()
	// for https add secure flag into cookie too
	c := new(http.Cookie)
	c.Name = "token"
	c.Value = token
	c.Path = "/"
	c.Expires = time.Now().Add(24 * time.Hour)
	context.SetCookie(c)
	return context.String(http.StatusOK, "{\"token\":\"session_token\"}")
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	v1 := e.Group("/api/v1")
	v1.POST("/login", login)
	v1.GET("/refresh", refresh)
	v1.GET("/ws", bash)
	e.Logger.Fatal(e.Start(*addr))
}
