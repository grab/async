package service

import (
    "testing"

    "github.com/grab/async/engine/sample/config"
    "github.com/stretchr/testify/assert"
)

func TestConcretePlan_IsAnalyzed(t *testing.T) {
    assert.True(t, config.Engine.IsAnalyzed(ConcretePlan{}))
}

func TestConcretePlan_IsExecutable(t *testing.T) {
    assert.Nil(t, config.Engine.IsPlanExecutable(&ConcretePlan{}))
}
