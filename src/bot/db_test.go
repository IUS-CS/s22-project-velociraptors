package db

import (
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func ConnectToTestDB() (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", "testDB")
	if err != nil {
		oops(err, "Open()")
		return nil, err
	}
	return db, nil
}

func TestConnectToDB(t *testing.T) {
	db, err := ConnectToDB()
	if err != nil {
		t.Errorf("got %t, wanted an nil", err)
	}
	db.Close()
}

func TestInitChallengeTableEntryCorrect(t *testing.T) {

	actual := initChallengeTableEntry("0", "1", "Gabe", "2", "Miia")
	expectedChallengeID := "0"
	expectedChallengerID := "1"
	expectedChallengerName := "Gabe"
	expectedDefenderID := "2"
	expectedDefenderName := "Miia"
	if expectedChallengeID != actual.MessageID {
		t.Errorf("got %q, wanted% q", expectedChallengeID, actual.MessageID)
	}
	if expectedChallengerID != actual.ChallengerID {
		t.Errorf("got %q, wanted% q", expectedChallengerID, actual.ChallengerID)
	}
	if expectedChallengerName != actual.ChallengerName {
		t.Errorf("got %q, wanted% q", expectedChallengerName, actual.ChallengerName)
	}
	if expectedDefenderID != actual.DefenderID {
		t.Errorf("got %q, wanted% q", expectedDefenderID, actual.DefenderID)
	}
	if expectedDefenderName != actual.DefenderName {
		t.Errorf("got %q, wanted% q", expectedDefenderName, actual.DefenderName)
	}
}

func TestInitChallengeTableEntryError(t *testing.T) {

	actual := initChallengeTableEntry("0", "1", "Gabe", "Miia", "2")
	expectedChallengeID := "0"
	expectedChallengerID := "1"
	expectedChallengerName := "Gabe"
	expectedDefenderID := "2"
	expectedDefenderName := "Miia"
	if expectedChallengeID != actual.MessageID {
		t.Errorf("got %q, wanted% q", expectedChallengeID, actual.MessageID)
	}
	if expectedChallengerID != actual.ChallengerID {
		t.Errorf("got %q, wanted% q", expectedChallengerID, actual.ChallengerID)
	}
	if expectedChallengerName != actual.ChallengerName {
		t.Errorf("got %q, wanted% q", expectedChallengerName, actual.ChallengerName)
	}
	if expectedDefenderID == actual.DefenderID {
		t.Errorf("got %q, wanted% q", expectedDefenderID, actual.DefenderID)
	}
	if expectedDefenderName == actual.DefenderName {
		t.Errorf("got %q, wanted% q", expectedDefenderName, actual.DefenderName)
	}
}

func TestInitScoreBoardRowPass(t *testing.T) {
	actual := initScoreBoardRow("1", "Gabe")
	expectedUserID := "1"
	expectedUserName := "Gabe"
	if expectedUserID != actual.UserID {
		t.Errorf("got %q, wanted% q", expectedUserID, actual.UserID)
	}
	if expectedUserName != actual.Username {
		t.Errorf("got %q, wanted% q", expectedUserName, actual.Username)
	}
}

func TestInitScoreBoardRowFail(t *testing.T) {
	actual := initScoreBoardRow("Gabe", "1")
	expectedUserID := "1"
	expectedUserName := "Gabe"
	if expectedUserID == actual.UserID {
		t.Errorf("got %q, wanted% q", expectedUserID, actual.UserID)
	}
	if expectedUserName == actual.Username {
		t.Errorf("got %q, wanted% q", expectedUserName, actual.Username)
	}
}

func TestInsertChallengeRow(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	err = CreateChallengeTable(db)
	if err != nil {
		oops(err, "CreateChallengeTable")
		return
	}
	test := initChallengeTableEntry("0", "1", "Gabe", "2", "Miia")
	insertChallengeRow(db, test)
	actual, err := selectChallengeRow(db, "0")
	if err != nil {
		oops(err, "Selecting challenge row")
		return
	}
	expectedChallengeID := test.MessageID
	expectedChallengerID := test.ChallengerID
	expectedChallengerName := test.ChallengerName
	expectedDefenderID := test.DefenderID
	expectedDefenderName := test.DefenderName
	if expectedChallengeID != actual.MessageID {
		t.Errorf("got %q, wanted% q", expectedChallengeID, actual.MessageID)
	}
	if expectedChallengerID != actual.ChallengerID {
		t.Errorf("got %q, wanted% q", expectedChallengerID, actual.ChallengerID)
	}
	if expectedChallengerName != actual.ChallengerName {
		t.Errorf("got %q, wanted% q", expectedChallengerName, actual.ChallengerName)
	}
	if expectedDefenderID != actual.DefenderID {
		t.Errorf("got %q, wanted% q", expectedDefenderID, actual.DefenderID)
	}
	if expectedDefenderName != actual.DefenderName {
		t.Errorf("got %q, wanted% q", expectedDefenderName, actual.DefenderName)
	}
	db.Close()
}

func TestSelectVotes(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	actual, err := selectVotes(db, "0")
	if err != nil {
		oops(err, "Selecting votes")
		return
	}
	expectedChallengerVotes := 0
	expectedDefenderVotes := 0
	expectedAbstainVotes := 0
	expectedStopVotes := 0
	if expectedChallengerVotes != actual.ChallengerVotes {
		t.Errorf("got %q, wanted% q", expectedChallengerVotes, actual.ChallengerVotes)
	}
	if expectedDefenderVotes != actual.DefenderVotes {
		t.Errorf("got %q, wanted% q", expectedDefenderVotes, actual.DefenderVotes)
	}
	if expectedAbstainVotes != actual.AbstainVotes {
		t.Errorf("got %q, wanted% q", expectedAbstainVotes, actual.AbstainVotes)
	}
	if expectedStopVotes != actual.StopVotes {
		t.Errorf("got %q, wanted% q", expectedStopVotes, actual.StopVotes)
	}
	db.Close()
}

func TestUpdateVotes(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	votes, err := selectVotes(db, "0")
	if err != nil {
		oops(err, "Selecting votes")
		return
	}
	votes.ChallengerVotes = 2
	votes.DefenderVotes = 1
	votes.AbstainVotes = 3
	votes.StopVotes = 1
	expectedChallengerVotes := 2
	expectedDefenderVotes := 1
	expectedAbstainVotes := 3
	expectedStopVotes := 1
	updateVotes(db, "0", votes)
	actual, err := selectVotes(db, "0")
	if err != nil {
		oops(err, "Selecting votes")
		return
	}
	if expectedChallengerVotes != actual.ChallengerVotes {
		t.Errorf("got %q, wanted% q", expectedChallengerVotes, actual.ChallengerVotes)
	}
	if expectedDefenderVotes != actual.DefenderVotes {
		t.Errorf("got %q, wanted% q", expectedDefenderVotes, actual.DefenderVotes)
	}
	if expectedAbstainVotes != actual.AbstainVotes {
		t.Errorf("got %q, wanted% q", expectedAbstainVotes, actual.AbstainVotes)
	}
	if expectedStopVotes != actual.StopVotes {
		t.Errorf("got %q, wanted% q", expectedStopVotes, actual.StopVotes)
	}
	db.Close()
}

func TestUpdateOutcomeChallengerWin(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	votes, err := selectVotes(db, "0")
	if err != nil {
		oops(err, "Selecting votes")
		return
	}
	updateOutcome(db, "0", votes)
	actual, err := selectChallengeRow(db, "0")
	if err != nil {
		oops(err, "Selecting challenge row")
		return
	}
	expected := 1
	if actual.Outcome != expected {
		t.Errorf("got %q, wanted% q", actual.Outcome, expected)
	}
	db.Close()
}
func TestUpdateOutcomeDefenderWin(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	test := initChallengeTableEntry("1", "1", "Gabe", "2", "Miia")
	insertChallengeRow(db, test)

	votes, err := selectVotes(db, "1")
	if err != nil {
		oops(err, "Selecting votes")
		return
	}
	votes.ChallengerVotes = 1
	votes.DefenderVotes = 2
	votes.AbstainVotes = 3
	votes.StopVotes = 1
	updateOutcome(db, "1", votes)
	actual, err := selectChallengeRow(db, "1")
	if err != nil {
		oops(err, "Selecting challenge row")
		return
	}
	expected := 2
	if actual.Outcome != expected {
		t.Errorf("got %q, wanted% q", actual.Outcome, expected)
	}
	db.Close()
}

func TestUpdateOutcomeTie(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	test := initChallengeTableEntry("2", "1", "Gabe", "2", "Miia")
	insertChallengeRow(db, test)

	votes, err := selectVotes(db, "2")
	if err != nil {
		oops(err, "Selecting votes")
		return
	}
	votes.ChallengerVotes = 1
	votes.DefenderVotes = 1
	votes.AbstainVotes = 3
	votes.StopVotes = 1
	updateOutcome(db, "2", votes)
	actual, err := selectChallengeRow(db, "2")
	if err != nil {
		oops(err, "Selecting challenge row")
		return
	}
	expected := 0
	if actual.Outcome != expected {
		t.Errorf("got %q, wanted% q", actual.Outcome, expected)
	}
	db.Close()
}

func TestInitScoreBoardRow(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	CreateScoreboardTable(db)
	actual := initScoreBoardRow("1", "Gabe")
	expectedUserID := "1"
	expectedUsername := "Gabe"
	if actual.UserID != expectedUserID {
		t.Errorf("got %q, wanted% q", actual.UserID, expectedUserID)
	}
	if actual.Username != expectedUsername {
		t.Errorf("got %q, wanted% q", actual.Username, expectedUsername)
	}
	db.Close()
}

func TestInsertScoreboardRow(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	score := initScoreBoardRow("1", "Gabe")
	insertScoreboardRow(db, score)
	actual, err := selectScoreboardRow(db, "1")
	if err != nil {
		oops(err, "Selecting scoreboard row")
		return
	}
	expectedUserID := "1"
	expectedUsername := "Gabe"
	if actual.UserID != expectedUserID {
		t.Errorf("got %q, wanted% q", actual.UserID, expectedUserID)
	}
	if actual.Username != expectedUsername {
		t.Errorf("got %q, wanted% q", actual.Username, expectedUsername)
	}
	db.Close()
}

func TestUpdateScoreboard(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	score, err := selectScoreboardRow(db, "1")
	if err != nil {
		t.Errorf("Selecting scoreboard row 1")
		return
	}
	score.TotalChallengeWins = 1
	score.TotalChallengeLosses = 2
	score.TotalChallengeTies = 3
	score.TotalChallenges = 6
	score.SuccessfulChallenges = 1
	score.FailedChallenges = 2
	score.SuccessfulDefenses = 2
	score.FailedDefenses = 2
	expectedTotalChallengeWins := 1
	expectedTotalChallengeLosses := 2
	expectedTotalChallengeTies := 3
	expectedTotalChallenges := 6
	expectedSuccessfulChallenges := 1
	expectedFailedChallenges := 2
	expectedSuccessfulDefenses := 2
	expectedFailedDefenses := 2
	updateScoreboard(db, score)
	actual, err := selectScoreboardRow(db, "1")
	if err != nil {
		t.Errorf("Selecting scoreboard row 2")
		return
	}
	if expectedTotalChallengeWins != actual.TotalChallengeWins {
		t.Errorf("got %q, wanted% q", actual.TotalChallengeWins, expectedTotalChallengeWins)
	}
	if expectedTotalChallengeLosses != actual.TotalChallengeLosses {
		t.Errorf("got %q, wanted% q", actual.TotalChallengeLosses, expectedTotalChallengeLosses)
	}
	if expectedTotalChallengeTies != actual.TotalChallengeTies {
		t.Errorf("got %q, wanted% q", actual.TotalChallengeTies, expectedTotalChallengeTies)
	}
	if expectedTotalChallenges != actual.TotalChallenges {
		t.Errorf("got %q, wanted% q", actual.TotalChallenges, expectedTotalChallenges)
	}
	if expectedSuccessfulChallenges != actual.SuccessfulChallenges {
		t.Errorf("got %q, wanted% q", actual.SuccessfulChallenges, expectedSuccessfulChallenges)
	}
	if expectedFailedChallenges != actual.FailedChallenges {
		t.Errorf("got %q, wanted% q", actual.FailedChallenges, expectedFailedChallenges)
	}
	if expectedSuccessfulDefenses != actual.SuccessfulDefenses {
		t.Errorf("got %q, wanted% q", actual.SuccessfulDefenses, expectedSuccessfulDefenses)
	}
	if expectedFailedDefenses != actual.FailedDefenses {
		t.Errorf("got %q, wanted% q", actual.FailedDefenses, expectedFailedDefenses)
	}
	db.Close()
}

func TestUserInScoreboardTrue(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	actual := userInScoreboard(db, "1")
	expected := true
	if expected != actual {
		t.Errorf("got %t, wanted%t", actual, expected)
	}
	db.Close()
}

func TestUserInScoreboardFalse(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	actual := userInScoreboard(db, "3")
	expected := true
	if expected == actual {
		t.Errorf("got %t, wanted %t", actual, expected)
	}
	db.Close()
}

func TestWinnerID1(t *testing.T) {
	test := initChallengeTableEntry("2", "1", "Gabe", "2", "Miia")
	test.Outcome = 1
	actual := winnerID(test)
	expected := "1"
	if expected != actual {
		t.Errorf("got %q, wanted%q", actual, expected)
	}
}

func TestWinnerID2(t *testing.T) {
	test := initChallengeTableEntry("2", "1", "Gabe", "2", "Miia")
	test.Outcome = 2
	actual := winnerID(test)
	expected := "2"
	if expected != actual {
		t.Errorf("got %q, wanted%q", actual, expected)
	}
}

func TestWinnerIDtie(t *testing.T) {
	test := initChallengeTableEntry("2", "1", "Gabe", "2", "Miia")
	test.Outcome = 0
	actual := winnerID(test)
	expected := "tie"
	if expected != actual {
		t.Errorf("got %q, wanted%q", actual, expected)
	}
}

func TestScoreBoardToString(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	test, err := selectScoreboardRow(db, "1")
	if err != nil {
		t.Errorf("Selecting scoreboard row 1")
		return
	}
	actual := scoreboardToString(test)
	expected := "`Gabe\nTotal challenge wins: 1\nTotal challenge losses: 2\nTotal challenge ties: 3\nTotal challenges: 6\nWins as challenger: 1\nLosses as challenger: 2\nWins as defender: 2\nLosses as defender: 2`"
	if expected != actual {
		t.Errorf("got %q, wanted%q", actual, expected)
	}
	db.Close()
}

func TestSelectVotingRecordRowError(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	test, err := selectVotingRecordRow(db, "10", "10")
	if err == nil {
		t.Errorf("got %t, wanted an error", err)
	}
	test.DefenderVotes = 0
	db.Close()
}

func TestInsertVotingRecordRow(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	CreateVotingRecord(db)
	votingRecord := VotingRecordEntryStruct{"1", "0", 1, 0, 0, 1}
	insertVotingRecordRow(db, votingRecord)
	actual, err := selectVotingRecordRow(db, "1", "0")
	if err != nil {
		t.Errorf("select ")
		return
	}
	expectedUserID := "1"
	expectedChallengerVotes := 1
	expectedDefenderVotes := 0
	expectedAbstainVotes := 0
	expectedStopVotes := 1
	if actual.UserID != expectedUserID {
		t.Errorf("got %q, wanted% q", actual.UserID, expectedUserID)
	}
	if expectedChallengerVotes != actual.ChallengerVotes {
		t.Errorf("got %q, wanted% q", actual.ChallengerVotes, expectedChallengerVotes)
	}
	if expectedDefenderVotes != actual.DefenderVotes {
		t.Errorf("got %q, wanted% q", actual.DefenderVotes, expectedDefenderVotes)
	}
	if expectedAbstainVotes != actual.AbstainVotes {
		t.Errorf("got %q, wanted% q", actual.AbstainVotes, expectedAbstainVotes)
	}
	if expectedStopVotes != actual.StopVotes {
		t.Errorf("got %q, wanted% q", actual.StopVotes, expectedStopVotes)
	}
	db.Close()
}

func TestUpdateVotingRecord(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	test, err := selectVotingRecordRow(db, "1", "0")
	test.ChallengerVotes = 0
	test.DefenderVotes = 1
	expectedChallengerVotes := 0
	expectedDefenderVotes := 1
	updateVotingRecord(db, test)
	actual, err := selectVotingRecordRow(db, "1", "0")
	if expectedChallengerVotes != actual.ChallengerVotes {
		t.Errorf("got %q, wanted% q", actual.ChallengerVotes, expectedChallengerVotes)
	}
	if expectedDefenderVotes != actual.DefenderVotes {
		t.Errorf("got %q, wanted% q", actual.DefenderVotes, expectedDefenderVotes)
	}
	db.Close()
}

func TestHasVotedBlueTrue(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	votingRecord := VotingRecordEntryStruct{"1", "0", 1, 0, 0, 0}
	updateVotingRecord(db, votingRecord)
	if hasVotedBlue(db, votingRecord) != true {
		t.Errorf("got %t, wanted %t", hasVotedBlue(db, votingRecord), true)
	}
	db.Close()
}

func TestHasVotedBlueFalse(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	votingRecord := VotingRecordEntryStruct{"1", "0", 0, 0, 0, 0}
	updateVotingRecord(db, votingRecord)
	if hasVotedBlue(db, votingRecord) == true {
		t.Errorf("got %t, wanted %t", hasVotedBlue(db, votingRecord), false)
	}
	db.Close()
}

func TestHasVotedYellowTrue(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	votingRecord := VotingRecordEntryStruct{"1", "0", 0, 1, 0, 0}
	updateVotingRecord(db, votingRecord)
	if hasVotedYellow(db, votingRecord) != true {
		t.Errorf("got %t, wanted %t", hasVotedBlue(db, votingRecord), true)
	}
	db.Close()
}

func TestHasVotedYellowFalse(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	votingRecord := VotingRecordEntryStruct{"1", "0", 0, 0, 0, 0}
	updateVotingRecord(db, votingRecord)
	if hasVotedYellow(db, votingRecord) == true {
		t.Errorf("got %t, wanted %t", hasVotedBlue(db, votingRecord), false)
	}
	db.Close()
}

func TestHasVotedRedTrue(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	votingRecord := VotingRecordEntryStruct{"1", "0", 0, 0, 1, 0}
	updateVotingRecord(db, votingRecord)
	if hasVotedRed(db, votingRecord) != true {
		t.Errorf("got %t, wanted %t", hasVotedBlue(db, votingRecord), true)
	}
	db.Close()
}

func TestHasVotedRedFalse(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	votingRecord := VotingRecordEntryStruct{"1", "0", 0, 0, 0, 0}
	updateVotingRecord(db, votingRecord)
	if hasVotedRed(db, votingRecord) == true {
		t.Errorf("got %t, wanted %t", hasVotedBlue(db, votingRecord), false)
	}
	db.Close()
}

func TestHasVotedStopTrue(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	votingRecord := VotingRecordEntryStruct{"1", "0", 0, 0, 0, 1}
	updateVotingRecord(db, votingRecord)
	if hasVotedStop(db, votingRecord) != true {
		t.Errorf("got %t, wanted %t", hasVotedBlue(db, votingRecord), true)
	}
	db.Close()
}

func TestHasVotedStopFalse(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	votingRecord := VotingRecordEntryStruct{"1", "0", 0, 0, 0, 0}
	updateVotingRecord(db, votingRecord)
	if hasVotedStop(db, votingRecord) == true {
		t.Errorf("got %t, wanted %t", hasVotedBlue(db, votingRecord), false)
	}
	db.Close()
}

func TestRemoveVotingRecordRow(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	test, err := selectVotingRecordRow(db, "1", "0")
	removeVotingRecordRow(db, test)
	expected := ""
	actual, err := selectVotingRecordRow(db, "1", "0")
	if err == nil {
		t.Errorf("got %q, wanted% q", actual.UserID, expected)
	}
	db.Close()
}

func TestCheckStopVotes(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	expected := 1
	checkStopVotes(db, "0")
	if checkStopVotes(db, "0") != expected {
		t.Errorf("got %q, wanted %q", checkStopVotes(db, "0"), expected)
	}
	db.Close()
}

func TestPushScore1(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	insertScoreboardRow(db, initScoreBoardRow("2", "Miia"))
	challengeTable := ChallengeTableEntryStruct{"10", "1", "Gabe", "2", "Miia", 0, 0, 0, 0, 1}
	pushScore(db, challengeTable)
	challenger, err := selectScoreboardRow(db, "1")
	if err != nil {
		t.Errorf("selecting scoreboard row")
	}
	defender, err := selectScoreboardRow(db, "2")
	if err != nil {
		t.Errorf("selecting scoreboard row")
	}
	expectedSuccessfulChallenges := challenger.SuccessfulChallenges + 1
	expectedTotalChallengeWins := challenger.TotalChallengeWins + 1
	expectedCTotalChallenges := challenger.TotalChallenges + 1
	expectedTotalChallengeLosses := defender.TotalChallengeLosses + 1
	expectedFailedDefenses := defender.FailedDefenses + 1
	expectedDTotalCHallenges := defender.TotalChallenges + 1
	pushScore(db, challengeTable)
	challenger, err = selectScoreboardRow(db, "1")
	if err != nil {
		t.Errorf("selecting scoreboard row")
	}
	defender, err = selectScoreboardRow(db, "2")
	if err != nil {
		t.Errorf("selecting scoreboard row")
	}
	if expectedSuccessfulChallenges != challenger.SuccessfulChallenges {
		t.Errorf("got %q, wanted %q", challenger.SuccessfulChallenges, expectedSuccessfulChallenges)
	}
	if expectedTotalChallengeWins != challenger.TotalChallengeWins {
		t.Errorf("got %q, wanted %q", challenger.TotalChallengeWins, expectedTotalChallengeWins)
	}
	if expectedCTotalChallenges != challenger.TotalChallenges {
		t.Errorf("got %q, wanted %q", challenger.TotalChallenges, expectedCTotalChallenges)
	}
	if expectedTotalChallengeLosses != defender.TotalChallengeLosses {
		t.Errorf("got %q, wanted %q", defender.TotalChallengeLosses, expectedTotalChallengeLosses)
	}
	if expectedFailedDefenses != defender.FailedDefenses {
		t.Errorf("got %q, wanted %q", defender.FailedChallenges, expectedFailedDefenses)
	}
	if expectedDTotalCHallenges != defender.TotalChallenges {
		t.Errorf("got %q, wanted %q", defender.TotalChallenges, expectedDTotalCHallenges)
	}
	db.Close()
}

func TestPushScore2(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	challengeTable := ChallengeTableEntryStruct{"10", "1", "Gabe", "2", "Miia", 0, 0, 0, 0, 2}
	pushScore(db, challengeTable)
	challenger, err := selectScoreboardRow(db, "1")
	if err != nil {
		t.Errorf("selecting scoreboard row")
	}
	defender, err := selectScoreboardRow(db, "2")
	if err != nil {
		t.Errorf("selecting scoreboard row")
	}
	expectedFailedChallenges := challenger.FailedChallenges + 1
	expectedTotalChallengeLosses := challenger.TotalChallengeLosses + 1
	expectedCTotalChallenges := challenger.TotalChallenges + 1
	expectedTotalChallengeWins := defender.TotalChallengeWins + 1
	expectedSuccesfulDefenses := defender.SuccessfulDefenses + 1
	expectedDTotalCHallenges := defender.TotalChallenges + 1
	pushScore(db, challengeTable)
	challenger, err = selectScoreboardRow(db, "1")
	if err != nil {
		t.Errorf("selecting scoreboard row")
	}
	defender, err = selectScoreboardRow(db, "2")
	if err != nil {
		t.Errorf("selecting scoreboard row")
	}
	if expectedFailedChallenges != challenger.FailedChallenges {
		t.Errorf("got %q, wanted %q", challenger.FailedChallenges, expectedFailedChallenges)
	}
	if expectedTotalChallengeLosses != challenger.TotalChallengeLosses {
		t.Errorf("got %q, wanted %q", challenger.TotalChallengeWins, expectedTotalChallengeLosses)
	}
	if expectedCTotalChallenges != challenger.TotalChallenges {
		t.Errorf("got %q, wanted %q", challenger.TotalChallenges, expectedCTotalChallenges)
	}
	if expectedTotalChallengeWins != defender.TotalChallengeWins {
		t.Errorf("got %q, wanted %q", defender.TotalChallengeLosses, expectedTotalChallengeWins)
	}
	if expectedSuccesfulDefenses != defender.SuccessfulDefenses {
		t.Errorf("got %q, wanted %q", defender.SuccessfulChallenges, expectedSuccesfulDefenses)
	}
	if expectedDTotalCHallenges != defender.TotalChallenges {
		t.Errorf("got %q, wanted %q", defender.TotalChallenges, expectedDTotalCHallenges)
	}
	db.Close()
}

func TestPushScoretie(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	challengeTable := ChallengeTableEntryStruct{"10", "1", "Gabe", "2", "Miia", 0, 0, 0, 0, 0}
	challenger, err := selectScoreboardRow(db, "1")
	if err != nil {
		t.Errorf("selecting scoreboard row")
	}
	defender, err := selectScoreboardRow(db, "2")
	if err != nil {
		t.Errorf("selecting scoreboard row")
	}
	expectedCTotalChallengeTies := challenger.TotalChallengeTies + 1
	expectedCTotalChallenges := challenger.TotalChallenges + 1
	expectedDTotalChallengeTies := defender.TotalChallengeTies + 1
	expectedDTotalCHallenges := defender.TotalChallenges + 1
	pushScore(db, challengeTable)
	challenger, err = selectScoreboardRow(db, "1")
	if err != nil {
		t.Errorf("selecting scoreboard row")
	}
	defender, err = selectScoreboardRow(db, "2")
	if err != nil {
		t.Errorf("selecting scoreboard row")
	}
	if expectedCTotalChallengeTies != challenger.TotalChallengeTies {
		t.Errorf("got %q, wanted %q", challenger.TotalChallengeTies, expectedCTotalChallengeTies)
	}
	if expectedCTotalChallenges != challenger.TotalChallenges {
		t.Errorf("got %q, wanted %q", challenger.TotalChallenges, expectedCTotalChallenges)
	}
	if expectedDTotalChallengeTies != defender.TotalChallengeTies {
		t.Errorf("got %q, wanted %q", defender.TotalChallengeTies, expectedDTotalChallengeTies)
	}
	if expectedDTotalCHallenges != defender.TotalChallenges {
		t.Errorf("got %q, wanted %q", defender.TotalChallenges, expectedDTotalCHallenges)
	}
	db.Close()
}

func TestPushScoreChallengerError(t *testing.T) {
	db, err := ConnectToTestDB()
	if err != nil {
		t.Errorf("database not open")
		return
	}
	challengeTable := ChallengeTableEntryStruct{"10", "7", "Gabe", "2", "Miia", 0, 0, 0, 0, 0}
	pushScore(db, challengeTable)
	db.Close()
}
