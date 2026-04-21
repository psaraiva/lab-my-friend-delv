package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
)

// FaceWeight defines the probabilistic weight of a dice face.
// All face weights for a die must sum to 100.
type FaceWeight struct {
	Face   int     `json:"face"`
	Weight float64 `json:"weight"`
}

type DiceConfig struct {
	Sides int          `json:"sides"`
	Faces []FaceWeight `json:"faces"`
}

type DiceRecord struct {
	Name  string `json:"name"`
	Sides int    `json:"sides"`
}

type RollResult struct {
	DiceName string `json:"dice_name"`
	Sides    int    `json:"sides"`
	Value    int    `json:"value"`
}

// diceConfigs contains the weights parameterized by the number of faces, the sum must be 100.
var diceConfigs = map[int]DiceConfig{
	6: {
		Sides: 6,
		Faces: []FaceWeight{
			{Face: 1, Weight: 15.0},
			{Face: 2, Weight: 20.0},
			{Face: 3, Weight: 15.0},
			{Face: 4, Weight: 15.0},
			{Face: 5, Weight: 20.0},
			{Face: 6, Weight: 15.0},
		},
	},
	12: {
		Sides: 12,
		Faces: []FaceWeight{
			{Face: 1, Weight: 8.33},
			{Face: 2, Weight: 8.33},
			{Face: 3, Weight: 8.33},
			{Face: 4, Weight: 8.33},
			{Face: 5, Weight: 8.33},
			{Face: 6, Weight: 8.33},
			{Face: 7, Weight: 8.33},
			{Face: 8, Weight: 8.33},
			{Face: 9, Weight: 8.33},
			{Face: 10, Weight: 8.33},
			{Face: 11, Weight: 8.33},
			{Face: 12, Weight: 8.37},
		},
	},
	16: {
		Sides: 16,
		Faces: []FaceWeight{
			{Face: 1, Weight: 6.25},
			{Face: 2, Weight: 6.25},
			{Face: 3, Weight: 6.25},
			{Face: 4, Weight: 8.25},
			{Face: 5, Weight: 6.25},
			{Face: 6, Weight: 6.25},
			{Face: 7, Weight: 6.25},
			{Face: 8, Weight: 6.25},
			{Face: 9, Weight: 6.25},
			{Face: 10, Weight: 7.25},
			{Face: 11, Weight: 4.25},
			{Face: 12, Weight: 6.25},
			{Face: 13, Weight: 6.25},
			{Face: 14, Weight: 6.25},
			{Face: 15, Weight: 6.25},
			{Face: 16, Weight: 5.25},
		},
	},
}

func app03URL() string {
	return "http://localhost:9023"
}

func appPort() string {
	return ":9022"
}

func main() {
	gin.SetMode("debug")
	r := gin.Default()
	r.GET("/roll/:name", rollDice)
	r.Run(appPort())
}

func rollDice(c *gin.Context) {
	name := c.Param("name")

	dice, err := fetchDice(name)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": fmt.Sprintf("error consulting App03: %s", err.Error())})
		return
	}
	if dice == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("dice '%s' not found in registry", name)})
		return
	}

	config, ok := diceConfigs[dice.Sides]
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("weight configuration not found for %d faces", dice.Sides)})
		return
	}

	value := weightedRoll(config.Faces)

	c.JSON(http.StatusOK, RollResult{
		DiceName: dice.Name,
		Sides:    dice.Sides,
		Value:    value,
	})
}

func fetchDice(name string) (*DiceRecord, error) {
	resp, err := http.Get(fmt.Sprintf("%s/dices/%s", app03URL(), name))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status from App03: %d", resp.StatusCode)
	}

	var d DiceRecord
	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return nil, fmt.Errorf("failed to decode response from App03: %w", err)
	}
	return &d, nil
}

func weightedRoll(faces []FaceWeight) int {
	total := 0.0
	for _, fw := range faces {
		total += fw.Weight
	}

	r := rand.Float64() * total
	cumulative := 0.0
	for _, fw := range faces {
		cumulative += fw.Weight
		if r < cumulative {
			return fw.Face
		}
	}
	// 100% ?
	return faces[len(faces)-1].Face
}
