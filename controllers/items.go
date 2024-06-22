package controllers

import (
	"context"
	"encoding/json"
	"gin-try/schemas"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

var items = []schemas.Item{
	{ID: 1, Name: "item one"},
	{ID: 2, Name: "item two"},
}

var ctx = context.Background()
var rdb *redis.Client

// SetRedis imposta la connessione a Redis per i controllers
func SetRedis(redisClient *redis.Client) {
	rdb = redisClient
}

const (
	cacheDuration = 10 * time.Minute
	cachePrefix   = "items:"
)

// @Summary Get all items
// @Description Retrieve a list of all items
// @Produce json
// @Success 200 {array} schemas.Item
// @Router /items [get]
func GetItems(c *gin.Context) {
	cacheKey := cachePrefix + "all"

	// Controlla se gli items sono presenti nella cache
	val, err := rdb.Get(cacheKey).Result()
	if err == redis.Nil {
		// Se non sono nella cache, recuperali dalla sorgente
		var items []schemas.Item = getItemsFromSource()
		// Salva gli items nella cache
		jsonData, _ := json.Marshal(items)
		rdb.Set(cacheKey, jsonData, cacheDuration)
		c.JSON(http.StatusOK, items)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		// Se sono nella cache, restituiscili
		var items []schemas.Item
		json.Unmarshal([]byte(val), &items)
		c.JSON(http.StatusOK, items)
	}
}

// @Summary Get item by ID
// @Description Retrieve an item by its ID
// @Produce json
// @Param id path int true "Item ID"
// @Success 200 {object} schemas.Item
// @Router /items/{id} [get]
func GetItemsByID(c *gin.Context) {
	id := c.Param("id")
	cacheKey := cachePrefix + id

	// Controlla se l'item è presente nella cache
	val, err := rdb.Get(cacheKey).Result()
	if err == redis.Nil {
		// Se non è nella cache, recuperalo dalla sorgente
		item := getItemFromSourceByID(id)
		if item == nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Item not found"})
			return
		}
		// Salva l'item nella cache
		jsonData, _ := json.Marshal(item)
		rdb.Set(cacheKey, jsonData, cacheDuration)
		c.JSON(http.StatusOK, item)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		// Se è nella cache, restituiscilo
		var item schemas.Item
		json.Unmarshal([]byte(val), &item)
		c.JSON(http.StatusOK, item)
	}
}

// @Summary Search items by name
// @Description Retrieve items whose name contains the specified string
// @Produce json
// @Param name query string true "Item name to search"
// @Success 200 {array} schemas.Item
// @Router /items/search [get]
func SearchItemsByName(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing name query parameter"})
		return
	}

	cacheKey := cachePrefix + "search:" + name

	// Controlla se gli items sono presenti nella cache
	val, err := rdb.Get(cacheKey).Result()
	if err == redis.Nil {
		// Se non sono nella cache, recuperali dalla sorgente
		var foundItems []schemas.Item = searchItemsFromSourceByName(name)
		// Salva gli items nella cache
		jsonData, _ := json.Marshal(foundItems)
		rdb.Set(cacheKey, jsonData, cacheDuration)
		c.JSON(http.StatusOK, foundItems)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		// Se sono nella cache, restituiscili
		var foundItems []schemas.Item
		json.Unmarshal([]byte(val), &foundItems)
		c.JSON(http.StatusOK, foundItems)
	}
}

// @Summary Delete item by ID
// @Description Delete an item by its ID
// @Produce json
// @Param id path int true "Item ID"
// @Success 204 "No Content"
// @Router /items/{id} [delete]
func DeleteItem(c *gin.Context) {
	id := c.Param("id")

	// Elimina l'item dalla sorgente
	if !deleteItemFromSource(id) {
		c.JSON(http.StatusNotFound, gin.H{"message": "Item not found"})
		return
	}

	// Elimina l'item dalla cache
	cacheKey := cachePrefix + id
	rdb.Del(cacheKey)

	// Elimina la cache di tutti gli items
	rdb.Del(cachePrefix + "all")

	c.JSON(http.StatusNoContent, gin.H{"message": "Item deleted"})
}

// @Summary Create a new item
// @Description Create a new item with the provided JSON data
// @Accept json
// @Produce json
// @Param item body schemas.Item true "Item object"
// @Success 201 {object} schemas.Item
// @Router /items [post]
func CreateItem(c *gin.Context) {
	var newItem schemas.Item
	if err := c.BindJSON(&newItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	newItem.ID = len(getItemsFromSource()) + 1 // Genera un nuovo ID
	saveItemToSource(newItem)

	// Invalida la cache di tutti gli items
	rdb.Del(cachePrefix + "all")

	c.JSON(http.StatusCreated, newItem)
}

// @Summary Update an item by ID
// @Description Update an item by its ID with the provided JSON data
// @Accept json
// @Produce json
// @Param id path int true "Item ID"
// @Param item body schemas.Item true "Updated item object"
// @Success 200 {object} schemas.Item
// @Router /items/{id} [put]
func UpdatedItem(c *gin.Context) {
	id := c.Param("id")

	var updatedItem schemas.Item
	if err := c.BindJSON(&updatedItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if !updateItemInSource(id, updatedItem) {
		c.JSON(http.StatusNotFound, gin.H{"message": "Item not found"})
		return
	}

	// Invalida la cache del singolo item e di tutti gli items
	cacheKey := cachePrefix + id
	rdb.Del(cacheKey)
	rdb.Del(cachePrefix + "all")

	c.JSON(http.StatusOK, updatedItem)
}

// Funzioni fittizie per interagire con la sorgente dati
func getItemsFromSource() []schemas.Item {
	return []schemas.Item{
		{ID: 1, Name: "item one"},
		{ID: 2, Name: "item two"},
	}
}

func getItemFromSourceByID(id string) *schemas.Item {
	items := getItemsFromSource()
	for _, item := range items {
		if strconv.Itoa(item.ID) == id {
			return &item
		}
	}
	return nil
}

func searchItemsFromSourceByName(name string) []schemas.Item {
	items := getItemsFromSource()
	var foundItems []schemas.Item
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Name), strings.ToLower(name)) {
			foundItems = append(foundItems, item)
		}
	}
	return foundItems
}

func deleteItemFromSource(id string) bool {
	items := getItemsFromSource()
	for i, item := range items {
		if strconv.Itoa(item.ID) == id {
			items = append(items[:i], items[i+1:]...)
			return true
		}
	}
	return false
}

func saveItemToSource(item schemas.Item) {
	// Salva l'item nella sorgente (può essere un database, un file, ecc.)
}

func updateItemInSource(id string, updatedItem schemas.Item) bool {
	items := getItemsFromSource()
	for i, item := range items {
		if strconv.Itoa(item.ID) == id {
			items[i] = updatedItem
			return true
		}
	}
	return false
}
