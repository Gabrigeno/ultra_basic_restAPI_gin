package controllers

import (
	"gin-try/schemas"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var items = []schemas.Item{
	{ID: 1, Name: "item one"},
	{ID: 2, Name: "item two"},
}

// @Summary Get all items
// @Description Retrieve a list of all items
// @Produce json
// @Success 200 {array} schemas.Item
// @Router /items [get]
func GetItems(c *gin.Context) {
	c.JSON(http.StatusOK, items)
}

// @Summary Get item by ID
// @Description Retrieve an item by its ID
// @Produce json
// @Param id path int true "Item ID"
// @Success 200 {object} schemas.Item
// @Router /items/{id} [get]
func GetItemsByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}

	for _, item := range items {
		if item.ID == id {
			c.JSON(http.StatusOK, item)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"message": "Item not found"})
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

	var foundItems []schemas.Item
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Name), strings.ToLower(name)) {
			foundItems = append(foundItems, item)
		}
	}

	c.JSON(http.StatusOK, foundItems)
}

// @Summary Delete item by ID
// @Description Delete an item by its ID
// @Produce json
// @Param id path int true "Item ID"
// @Success 204 "No Content"
// @Router /items/{id} [delete]
func DeleteItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}

	for i, item := range items {
		if item.ID == id {
			items = append(items[:i], items[i+1:]...)
			c.JSON(http.StatusNoContent, gin.H{"message": "Item deleted"})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"message": "Item not found"})
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
	newItem.ID = len(items) + 1
	items = append(items, newItem)
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
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}

	var updatedItem schemas.Item
	if err := c.BindJSON(&updatedItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	for i, item := range items {
		if item.ID == id {
			items[i].Name = updatedItem.Name
			c.JSON(http.StatusOK, items[i])
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"message": "Item not found"})
}
