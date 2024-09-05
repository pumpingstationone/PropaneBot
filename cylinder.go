package main

import (
	"encoding/json"
	"io"
	"log"
	"math"
	"os"
)

// The purpose of this file is to store the base weight (tare)
// and full weight of the propane cylinder. The empty/full
// weight is used to calculate the percentage remaining as
// read from the scale

type Cylinder struct {
	TareWeight float64 `json:"tareweight"`
	FullWeight float64 `json:"fullweight"`
	// This is to take into consideration
	// additional weight on the scale, like
	// the regulator, hose, and safety chain
	ExtraWeight float64 `json:"extraweight"`
}

var cylinder Cylinder

func LoadCylinderData() {
	log.Println("Loading current cylinder info")
	jsonFile, err := os.Open("cylinder.json")
	if err != nil {
		log.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
	if err = json.Unmarshal(byteValue, &cylinder); err != nil {
		log.Panicf("Failed to read cylinder data: %s\n", err)
	}
}

// Gives us the percentage remaining for the given weight, taking into
// consideration the full and tare weight of the cylinder, plus any extra
// weight that might be on the scale
func (c Cylinder) CalcRemaining(currentWeight float64) float64 {
	base := cylinder.FullWeight - cylinder.TareWeight + cylinder.ExtraWeight
	cur := currentWeight - cylinder.TareWeight + cylinder.ExtraWeight
	delta := math.Round((cur / base) * 100)

	return delta
}
