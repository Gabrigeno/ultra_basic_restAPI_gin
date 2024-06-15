package tests

import (
	"encoding/json"
	"gin-try/controllers"
	"gin-try/schemas"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
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

func TestGetItems(t *testing.T) {
	// Setup -> inizializzo il router, creo la richiesta e registro la risposta
	router := setupRouter()
	req, _ := http.NewRequest("GET", "/items", nil)
	w := httptest.NewRecorder()

	// Perform request -> lancio la richiesta contro il router
	router.ServeHTTP(w, req)

	// Asserts -> controllo la risposta
	assert.Equal(t, http.StatusOK, w.Code)

	var responseItems []schemas.Item
	err := json.Unmarshal(w.Body.Bytes(), &responseItems) // Deserializza il corpo della risposta JSON nell'array responseItems.
	assert.Nil(t, err)                                    //  Verifica che non ci siano errori durante la deserializzazione.
	assert.Len(t, responseItems, 2)
}

func TestGetItemsByID(t *testing.T) {
	// Setup
	router := setupRouter()
	req, _ := http.NewRequest("GET", "/items/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var responseItem schemas.Item // Utilizzo di una variabile singola per deserializzare l'oggetto Item non metto []
	err := json.Unmarshal(w.Body.Bytes(), &responseItem)
	assert.Nil(t, err)
	assert.Equal(t, 1, responseItem.ID)
}

func TestSearchItemsByName(t *testing.T) {
	// Setup
	router := setupRouter()

	// una GET request a /items/search with query parameter "name"
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
}

func TestCreateItem(t *testing.T) {
	// Setup
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
}

func TestUpdateItem(t *testing.T) {
	// Setup
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
}

func TestDeleteItem(t *testing.T) {
	// Setup
	router := setupRouter()
	req, _ := http.NewRequest("DELETE", "/items/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	req, _ = http.NewRequest("GET", "/items/1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
