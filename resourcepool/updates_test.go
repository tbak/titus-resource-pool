package resourcepool

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestButResourceShapeLabels(t *testing.T) {
	require.Equal(t,
		map[string]string{"keyA": "valueA", "keyB": "valueB"},
		ButResourceShapeLabels(EmptyResourcePool(), "keyA", "valueA", "keyB", "valueB").Spec.ResourceShape.Labels,
	)
	require.Equal(t,
		map[string]string{"keyA": "valueA", "keyB": ""},
		ButResourceShapeLabels(EmptyResourcePool(), "keyA", "valueA", "keyB", "").Spec.ResourceShape.Labels,
	)
}
