package main

import (
	"net/http"
	"regexp"
	"sort"
	"sync"

	"github.com/gin-gonic/gin"
)

type Dice struct {
	Name  string `json:"name"`
	Sides int    `json:"sides"`
}

type CreateDiceRequest struct {
	Name  string `json:"name"  binding:"required"`
	Sides int    `json:"sides" binding:"required"`
}

var (
	store      = map[string]Dice{}
	storeMu    sync.RWMutex
	nameRe     = regexp.MustCompile(`^[a-zA-Z0-9]{1,25}$`)
	validSides = map[int]bool{6: true, 12: true, 16: true}
)

func main() {
	gin.SetMode("debug")
	r := gin.Default()

	r.POST("/dices", createDice)
	r.GET("/dices", listDice)
	r.GET("/dices/:name", getDice)
	r.DELETE("/dices/:name", deleteDice)

	r.Run(":9023")
}

func createDice(c *gin.Context) {
	var req CreateDiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !nameRe.MatchString(req.Name) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name must have 1-25 alphanumeric characters (a-Z, 0-9)"})
		return
	}

	if !validSides[req.Sides] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "number of sides must be 6, 12 or 16"})
		return
	}

	storeMu.Lock()
	defer storeMu.Unlock()

	if _, exists := store[req.Name]; exists {
		c.JSON(http.StatusConflict, gin.H{"error": "dice already registered with this name"})
		return
	}

	d := Dice{Name: req.Name, Sides: req.Sides}
	store[req.Name] = d
	c.JSON(http.StatusCreated, d)
}

func listDice(c *gin.Context) {
	storeMu.RLock()
	defer storeMu.RUnlock()

	list := make([]Dice, 0, len(store))
	for _, d := range store {
		list = append(list, d)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })

	c.JSON(http.StatusOK, list)
}

func getDice(c *gin.Context) {
	name := c.Param("name")

	storeMu.RLock()
	defer storeMu.RUnlock()

	d, exists := store[name]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "dice not found"})
		return
	}
	c.JSON(http.StatusOK, d)
}

func deleteDice(c *gin.Context) {
	name := c.Param("name")

	storeMu.Lock()
	defer storeMu.Unlock()

	if _, exists := store[name]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "dice not found"})
		return
	}

	delete(store, name)
	c.JSON(http.StatusOK, gin.H{"message": "dice removed successfully"})
}
