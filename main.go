package main

import (
	"github.com/gin-gonic/gin"

	"gin-try/controllers"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "gin-try/docs" // Importa il pacchetto docs generato da Swaggo per le docs sul brawser
)

func main() {
	r := gin.Default()
	// Middleware per servire la documentazione Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// route
	r.GET("/ping", controllers.GetPing)
	r.GET("/items", controllers.GetItems)
	r.GET("/items/:id", controllers.GetItemsByID)
	r.GET("/items/search", controllers.SearchItemsByName)
	r.POST("/items", controllers.CreateItem)
	r.DELETE("/items/:id", controllers.DeleteItem)
	r.PUT("/items/:id", controllers.UpdatedItem)
	r.Run(":8080") // listen and serve on localhost:8080
}
