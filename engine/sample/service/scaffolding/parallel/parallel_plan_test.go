package parallel

import (
	"testing"

	"github.com/grab/async/engine/sample/config"
	"github.com/stretchr/testify/assert"
)

func TestParallelPlan_IsAnalyzed(t *testing.T) {
	assert.True(t, config.Engine.IsAnalyzed(&ParallelPlan{}))
}

func TestParallelPlan_IsExecutable(t *testing.T) {
	assert.Nil(t, config.Engine.IsExecutable(&ParallelPlan{}))
}
