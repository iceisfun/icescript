package main

import (
	"fmt"
	"time"
)

type Gotype struct {
	internalCreate time.Time
	internalState  int
}

func NewGotype() *Gotype {
	return &Gotype{
		internalCreate: time.Now(),
		internalState:  0,
	}
}

func (g *Gotype) String() string {
	return fmt.Sprintf("Gotype{Create: %v, State: %d}", g.internalCreate, g.internalState)
}
