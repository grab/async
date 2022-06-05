package travelcost

import (
	"testing"

	"github.com/grab/async/engine/sample/config"
	"github.com/stretchr/testify/assert"
)

func TestComputer_IsRegistered(t *testing.T) {
	assert.True(t, config.Engine.IsRegistered(TravelCost{}))
}
