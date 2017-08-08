package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringArrayContains(t *testing.T) {
	arr := []string{"lorem", "ipsum", "dolor", "sit", "amet"}

	assert.True(t, StringArrayContains(arr, "ipsum"), "array should contain 'ipsum'")
	assert.False(t, StringArrayContains(arr, "engage"), "array should not contain 'engage'")
}

func TestRemoveFromSlice(t *testing.T) {
	arr := []string{"lorem", "ipsum", "dolor", "sit", "amet"}
	newArr := RemoveFromSlice(arr, "ipsum")

	assert.Equal(t, 4, len(newArr), "array should contain one less element")
	assert.False(t, StringArrayContains(newArr, "ipsum"))
}
