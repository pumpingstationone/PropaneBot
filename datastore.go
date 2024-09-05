package main

import (
	"fmt"
	"sync"
	"time"
)

type CurrentData struct {
	Weight    float64
	TimeStamp time.Time
	// We'll calculate this when setting so the
	// bot doesn't have to
	Remaining float64
}

type Datastore struct {
	data CurrentData
	lock *sync.RWMutex
}

func NewDatastore() *Datastore {
	return &Datastore{
		data: CurrentData{},
		lock: &sync.RWMutex{},
	}
}

func (d *Datastore) Get() CurrentData {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return d.data
}

func (d *Datastore) GetString() string {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return fmt.Sprintf(
		"Well, as of %s the cylinder weighs %.0f lbs which kinda translates into %.0f%% remaining",
		d.data.TimeStamp.Format("Mon Jan _2 03:04PM 2006"),
		d.data.Weight,
		d.data.Remaining,
	)
}

func (d *Datastore) Set(weight float64, timestamp time.Time, remaining float64) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.data.Weight = weight
	d.data.TimeStamp = timestamp
	d.data.Remaining = remaining
}
