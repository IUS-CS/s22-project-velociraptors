package main

import (
	"testing"
)

func TestInitChallengeTableEntry(t *testing.T) {

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
	if expectedChallengeID != actualChallengeID {
		t.Errorf("got %q, wanted% q", expectedChallengeID, actualChallengeID)
	}
	if expectedChallengerID != actualChallengerID {
		t.Errorf("got %q, wanted% q", expectedChallengerID, actualChallengerID)
	}
	if expectedChallengerName != actualChallengerName {
		t.Errorf("got %q, wanted% q", expectedChallengerName, actualChallengerName)
	}
	if expectedDefenderID != actualDefenderID {
		t.Errorf("got %q, wanted% q", expectedDefenderID, actualDefenderID)
	}
	if expectedDefenderName != actualDefenderName {
		t.Errorf("got %q, wanted% q", expectedDefenderName, actualDefenderName)
	}
}
