package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dravog7/GameBox/connection"
	"github.com/dravog7/GameBox/room"
	"github.com/dravog7/paper-ink-backend/rooms"

	"github.com/gofiber/fiber"
	"github.com/gofiber/websocket"
)

func main() {
	app := fiber.New()

	app.Static("/", "./statics")
	app.Use(func(c *fiber.Ctx) {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			c.Next()
		}
	})

	chatroom := &rooms.Entry{Name: "Entry"}
	chatroom.Init()
	manager := room.DefaultManager{}
	factory := &connection.WebSocketConnectionFactory{}

	manager.Register(chatroom)
	manager.AddFactory(factory, func(err error) {
		fmt.Println(err)
	})

	app.Get("/ws", factory.Setup(func(c *websocket.Conn) map[string]string {
		params := map[string]string{
			"id": "Entry",
		}
		fmt.Println("in factory")
		return params
	}))
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}
	app.Listen(port)
}
