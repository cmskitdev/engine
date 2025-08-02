package engine

import (
	"fmt"
	"testing"
)

func TestEngine(t *testing.T) {
	engine := NewEngine[any](NewEngineOptions{
		Plugins: []string{"redis"},
	})

	fmt.Printf("Engine: %v\n", engine)
}
