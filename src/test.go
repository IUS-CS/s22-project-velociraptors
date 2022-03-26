package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestinitChallengeTableEntry(t *testing.T) {
	test := initChallengeTableEntry("1", "Gabe", "2", "Miia")
	expectedChallengeID := 0
	expectedChallengerID := "1"
	expectedChallengerName := "Gabe"
	expectedDefenderID := "2"
	expectedDefenderName := "Miia"
	actualChallengeID := test.ChallengeID
	actualChallengerID := test.ChallengerID
	actualChallengerName := test.ChallengerName
	actualDefenderID := test.DefenderID
	actualDefenderName := test.DefenderName
	assert.Equal(t, expectedChallengeID, actualChallengeID)
	assert.Equal(t, expectedChallengerID, actualChallengerID)
	assert.Equal(t, expectedChallengerName, actualChallengerName)
	assert.Equal(t, expectedDefenderID, actualDefenderID)
	assert.Equal(t, expectedDefenderName, actualDefenderName)
}
