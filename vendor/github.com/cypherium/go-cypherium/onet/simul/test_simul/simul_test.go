package main

import (
	"testing"

	"github.com/cypherium/go-cypherium/onet/simul"
)

func TestSimulation(t *testing.T) {
	simul.Start("count.toml")
}
