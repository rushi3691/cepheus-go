package main

import (
	"cephgo/database"
	"cephgo/routes"
	"log"
	"os"
	"os/signal"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("0")
		log.Panic("couldn't load .env file")
	}

	app := fiber.New()

	c := make(chan os.Signal, 1)   // Create channel to signify a signal being sent
	signal.Notify(c, os.Interrupt) // When an interrupt is sent, notify the channel

	// Goroutine to monitor the channel and run app.Shutdown when an interrupt is recieved
	// This should cause app.Listen to return nil, then allowing the cleanup tasks to be
	// run.
	go func() {
		<-c
		log.Println("Gracefully shutting down...")
		_ = app.Shutdown()
	}()

	// apply middlewares
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000, https://cepheustest.netlify.app",
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowCredentials: true,
	}))
	app.Use(compress.New())

	routes.SetupRoutes(app)
	err := database.CreateDBPool(os.Getenv("RENDER_DB"))
	if err != nil {
		log.Println("1")
		log.Panic(err)
	}
	if err := app.Listen(":8000"); err != nil {
		log.Println("2")
		log.Panic(err)
	}

	log.Println("Running cleanup tasks...")
	database.DB_STRUCT.Pool.Close()
	log.Println("database closed")

}
