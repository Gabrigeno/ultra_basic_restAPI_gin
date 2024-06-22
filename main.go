package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"gin-try/controllers"

	_ "gin-try/docs" // Importa il pacchetto docs generato da Swaggo per le docs sul brawser
)

func main() {
	// Carica le variabili d'ambiente dal file .env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Errore nel caricare il file .env: %v", err)
	}

	// Esempio di utilizzo di una variabile d'ambiente
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	// Connessione a Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	// Verifica la connessione a Redis
	pong, err := rdb.Ping().Result()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	log.Printf("Success!: %s", pong)
	controllers.SetRedis(rdb)

	router := gin.Default()

	// Imposta le rotte
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/ping", controllers.GetPing)
	router.GET("/items", controllers.GetItems)
	router.POST("/items", controllers.CreateItem)
	router.GET("/items/search", controllers.SearchItemsByName)
	router.GET("/items/:id", controllers.GetItemsByID)
	router.DELETE("/items/:id", controllers.DeleteItem)
	router.PUT("/items/:id", controllers.UpdatedItem)
	router.Run(":8080")
	// Avvia il server
	if err := router.Run(); err != nil {
		log.Fatalf("Errore nell'avvio del server: %v", err)
	}
}
