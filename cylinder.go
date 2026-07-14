package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"math"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// The purpose of this file is to store the base weight (tare)
// and full weight of the propane cylinder. The empty/full
// weight is used to calculate the percentage remaining as
// read from the scale

const cylinderFile = "cylinder.json"

type Cylinder struct {
	TareWeight float64 `json:"tareweight"`
	FullWeight float64 `json:"fullweight"`
	// This is to take into consideration
	// additional weight on the scale, like
	// the regulator, hose, and safety chain
	ExtraWeight float64 `json:"extraweight"`
}

var (
	cylinder   Cylinder
	cylinderMu sync.RWMutex
)

func LoadCylinderData() {
	log.Println("Loading current cylinder info")
	jsonFile, err := os.Open(cylinderFile)
	if err != nil {
		log.Println(err)
		return
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Printf("Failed to read cylinder data: %s\n", err)
		return
	}

	var c Cylinder
	if err := json.Unmarshal(byteValue, &c); err != nil {
		log.Printf("Failed to parse cylinder data: %s\n", err)
		return
	}

	cylinderMu.Lock()
	cylinder = c
	cylinderMu.Unlock()
}

// GetCylinderData returns a copy of the currently loaded cylinder settings
func GetCylinderData() Cylinder {
	cylinderMu.RLock()
	defer cylinderMu.RUnlock()
	return cylinder
}

// SaveCylinderData writes the given cylinder settings to cylinder.json and
// updates the in-memory copy used for calculations
func SaveCylinderData(c Cylinder) error {
	data, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(cylinderFile, data, 0644); err != nil {
		return err
	}

	cylinderMu.Lock()
	cylinder = c
	cylinderMu.Unlock()

	return nil
}

// WatchCylinderData watches cylinder.json on disk and reloads it into memory
// whenever it changes, so edits made outside the web page (or by the web
// page's handler writing the file directly) are picked up automatically.
func WatchCylinderData(ctx context.Context) func() error {
	return func() error {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}
		defer watcher.Close()

		if err := watcher.Add(cylinderFile); err != nil {
			return err
		}

		for {
			select {
			case <-ctx.Done():
				return nil
			case event, ok := <-watcher.Events:
				if !ok {
					return nil
				}
				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					LoadCylinderData()
				}
				// Some editors/writers replace the file instead of writing
				// in place, which drops the watch and needs it re-added.
				if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
					_ = watcher.Add(cylinderFile)
					LoadCylinderData()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return nil
				}
				log.Printf("cylinder.json watcher error: %v", err)
			}
		}
	}
}

// Gives us the percentage remaining for the given weight, taking into
// consideration the full and tare weight of the cylinder, plus any extra
// weight that might be on the scale
func (c Cylinder) CalcRemaining(currentWeight float64) float64 {
	cur := GetCylinderData()
	base := cur.FullWeight - cur.TareWeight + cur.ExtraWeight
	adjusted := currentWeight - cur.TareWeight + cur.ExtraWeight
	delta := math.Round((adjusted / base) * 100)

	return delta
}
