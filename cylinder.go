package main

import (
	"log"
	"fmt"
	"os"
	"io/ioutil"
	"encoding/json"
	"math"
	"strconv"
)

// The purpose of this file is to store the base weight (tare)
// and full weight of the propane cylinder. The empty/full
// weight is used to calculate the percentage remaining as
// read from the scale

type Cylinder struct {
	TareWeight string `json:"tareweight"`
	FullWeight string `json:"fullweight"`
	// This is to take into consideration 
	// additional weight on the scale, like
	// the regulator, hose, and safety chain
	ExtraWeight string `json:"extraweight"`
}

var cylinder Cylinder

func loadCylinderData() {
	log.Println("Loading current cylinder info")
	jsonFile, err := os.Open("cylinder.json")	
	if err != nil {
    	log.Println(err)
	}		
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &cylinder)
}

func convert(val string) float64 {
	convVal, err := strconv.ParseFloat(val, 32) 
	if err != nil {
		log.Println("Drat, got", err)
	}

	return convVal
}

// Gives us the percentage remaining (as a string)
// for the given weight, taking into consideration
// the full and tare weight of the cylinder, plus
// any extra weight that might be on the scale
func calcRemaining(currentWeight string) string {
	cw := convert(currentWeight)
	tw := convert(cylinder.TareWeight)
	fw := convert(cylinder.FullWeight)
	ew := convert(cylinder.ExtraWeight)

	base := fw - tw + ew
	cur := cw - tw + ew
	delta := math.Round((cur/base) * 100)
	
	return fmt.Sprintf("%.0f%%", delta)
}