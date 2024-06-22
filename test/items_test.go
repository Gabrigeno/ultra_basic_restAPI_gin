package tests

import (
	"encoding/json"
	"gin-try/controllers"
	"gin-try/schemas"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/items", controllers.GetItems)
	r.GET("/items/:id", controllers.GetItemsByID)
	r.GET("/items/search", controllers.SearchItemsByName)
	r.DELETE("/items/:id", controllers.DeleteItem)
	r.POST("/items", controllers.CreateItem)
	r.PUT("/items/:id", controllers.UpdatedItem)

	return r
}

func setupRedis() (*miniredis.Miniredis, *redis.Client) {
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	controllers.SetRedis(client)

	return mr, client
}

func TestGetItems(t *testing.T) {
	// Setup
	mr, client := setupRedis()
	defer mr.Close()
	defer client.Close()

	router := setupRouter()
	req, _ := http.NewRequest("GET", "/items", nil)
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Asserts
	assert.Equal(t, http.StatusOK, w.Code)

	var responseItems []schemas.Item
	err := json.Unmarshal(w.Body.Bytes(), &responseItems)
	assert.Nil(t, err)
	assert.Len(t, responseItems, 2)

	// Check if items are cached
	cacheKey := "items:all"
	cachedData, err := client.Get(cacheKey).Result()
	assert.Nil(t, err)

	var cachedItems []schemas.Item
	err = json.Unmarshal([]byte(cachedData), &cachedItems)
	assert.Nil(t, err)
	assert.Len(t, cachedItems, 2)
}

func TestGetItemsByID(t *testing.T) {
	// Setup
	mr, client := setupRedis()
	defer mr.Close()
	defer client.Close()

	router := setupRouter()
	req, _ := http.NewRequest("GET", "/items/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var responseItem schemas.Item
	err := json.Unmarshal(w.Body.Bytes(), &responseItem)
	assert.Nil(t, err)
	assert.Equal(t, 1, responseItem.ID)

	// Check if item is cached
	cacheKey := "items:1"
	cachedData, err := client.Get(cacheKey).Result()
	assert.Nil(t, err)

	var cachedItem schemas.Item
	err = json.Unmarshal([]byte(cachedData), &cachedItem)
	assert.Nil(t, err)
	assert.Equal(t, 1, cachedItem.ID)
}

func TestSearchItemsByName(t *testing.T) {
	// Setup
	mr, client := setupRedis()
	defer mr.Close()
	defer client.Close()

	router := setupRouter()
	req, _ := http.NewRequest("GET", "/items/search?name=item", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var responseItems []schemas.Item
	err := json.Unmarshal(w.Body.Bytes(), &responseItems)
	assert.Nil(t, err)

	assert.True(t, len(responseItems) > 0)

	for _, item := range responseItems {
		assert.Contains(t, strings.ToLower(item.Name), "item")
	}

	// Check if search results are cached
	cacheKey := "items:search:item"
	cachedData, err := client.Get(cacheKey).Result()
	assert.Nil(t, err)

	var cachedItems []schemas.Item
	err = json.Unmarshal([]byte(cachedData), &cachedItems)
	assert.Nil(t, err)
	assert.True(t, len(cachedItems) > 0)

	for _, item := range cachedItems {
		assert.Contains(t, strings.ToLower(item.Name), "item")
	}
}

func TestCreateItem(t *testing.T) {
	// Setup
	mr, client := setupRedis()
	defer mr.Close()
	defer client.Close()

	router := setupRouter()

	testNewItem := schemas.Item{
		ID:   3,
		Name: "New Item",
	}
	payload, _ := json.Marshal(testNewItem)
	req, _ := http.NewRequest("POST", "/items", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var responseItem schemas.Item
	err := json.Unmarshal(w.Body.Bytes(), &responseItem)
	assert.Nil(t, err)
	assert.Equal(t, testNewItem.ID, responseItem.ID)
	assert.Equal(t, testNewItem.Name, responseItem.Name)

	// Check if cache for all items is invalidated
	_, err = client.Get("items:all").Result()
	assert.Equal(t, redis.Nil, err)
}

func TestUpdateItem(t *testing.T) {
	// Setup
	mr, client := setupRedis()
	defer mr.Close()
	defer client.Close()

	router := setupRouter()

	updatedItem := schemas.Item{
		ID:   1,
		Name: "Updated Item",
	}
	payload, _ := json.Marshal(updatedItem)
	req, _ := http.NewRequest("PUT", "/items/1", strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var responseItem schemas.Item
	err := json.Unmarshal(w.Body.Bytes(), &responseItem)
	assert.Nil(t, err)
	assert.Equal(t, updatedItem.ID, responseItem.ID)
	assert.Equal(t, updatedItem.Name, responseItem.Name)

	// Check if cache for the item and all items is invalidated
	_, err = client.Get("items:1").Result()
	assert.Equal(t, redis.Nil, err)
	_, err = client.Get("items:all").Result()
	assert.Equal(t, redis.Nil, err)
}

func TestDeleteItem(t *testing.T) {
	// Setup
	mr, client := setupRedis()
	defer mr.Close()
	defer client.Close()

	router := setupRouter()
	req, _ := http.NewRequest("DELETE", "/items/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Check if cache for the item is invalidated
	_, err := client.Get("items:1").Result()
	assert.Equal(t, redis.Nil, err)

	// Check if cache for all items is invalidated
	_, err = client.Get("items:all").Result()
	assert.Equal(t, redis.Nil, err)

}
