package db

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"log"
	"strconv"
	"time"
)

func rowsAffected(rows int64, task string) {
	log.Printf("%d rows affected while %s", rows, task)
}

// ConnectToDB creates DB if it doesn't exist already
func ConnectToDB() (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", dbname)
	if err != nil {
		oops(err, "Open()")
		return nil, err
	}
	return db, nil
}

//ChallengeTableEntryStruct fields
type ChallengeTableEntryStruct struct {
	MessageID       string `db:"MessageID"`
	ChallengerID    string `db:"ChallengerID"`
	ChallengerName  string `db:"ChallengerName"`
	DefenderID      string `db:"DefenderID"`
	DefenderName    string `db:"DefenderName"`
	ChallengerVotes int    `db:"ChallengerVotes"`
	DefenderVotes   int    `db:"DefenderVotes"`
	AbstainVotes    int    `db:"AbstainVotes"`
	StopVotes       int    `db:"StopVotes"`
	Outcome         int    `db:"Outcome"`
	//0=tie, 1=challenger wins, 2=defender wins
}

type ScoreboardTableEntryStruct struct {
	UserID               string `db:"UserID"`
	Username             string `db:"Username"`
	TotalChallengeWins   int    `db:"TotalChallengeWins"`
	TotalChallengeLosses int    `db:"TotalChallengeLosses"`
	TotalChallengeTies   int    `db:"TotalChallengeTies"`
	TotalChallenges      int    `db:"TotalChallenges"`
	SuccessfulChallenges int    `db:"SuccessfulChallenges"`
	FailedChallenges     int    `db:"FailedChallenges"`
	SuccessfulDefenses   int    `db:"SuccessfulDefenses"`
	FailedDefenses       int    `db:"FailedDefenses"`
}

type VotingRecordEntryStruct struct {
	UserID          string `db:"UserID"`
	MessageID       string `db:"MessageID"`
	ChallengerVotes int    `db:"ChallengerVotes"`
	DefenderVotes   int    `db:"DefenderVotes"`
	AbstainVotes    int    `db:"AbstainVotes"`
	StopVotes       int    `db:"StopVotes"`
}

type VotesStruct struct {
	ChallengerVotes int `db:"ChallengerVotes"`
	DefenderVotes   int `db:"DefenderVotes"`
	AbstainVotes    int `db:"AbstainVotes"`
	StopVotes       int `db:"StopVotes"`
}

// CreateChallengeTable this table stores values for challenge votes
func CreateChallengeTable(db *sqlx.DB) error {
	query := "CREATE TABLE IF NOT EXISTS challengeTable(MessageID string primary key, ChallengerID text, ChallengerName text, DefenderID text, DefenderName text, ChallengerVotes int, DefenderVotes int, AbstainVotes int, StopVotes int, Outcome int)"
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	res, err := db.ExecContext(ctx, query)
	if err != nil {
		oops(err, "ExecContext")
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		oops(err, "RowsAffected")
		return err
	}
	rowsAffected(rows, "creating challenge table")
	return nil
}

func insertChallengeRow(db *sqlx.DB, row ChallengeTableEntryStruct) {
	query := "INSERT INTO challengeTable (MessageID, ChallengerID, ChallengerName, DefenderID, DefenderName, ChallengerVotes, DefenderVotes, AbstainVotes, StopVotes, Outcome) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	stmt, err := db.Prepare(query)
	if err != nil {
		oops(err, "prepare insertChallengeRow")
		return
	}
	res, err := stmt.Exec(row.MessageID, row.ChallengerID, row.ChallengerName, row.DefenderID, row.DefenderName, row.ChallengerVotes, row.DefenderVotes, row.AbstainVotes, row.StopVotes, row.Outcome)
	if err != nil {
		oops(err, "execute insertChallengeRow")
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		oops(err, "RowsAffected")
		return
	}
	rowsAffected(rows, "inserting challenge row")
	return
}

func initChallengeTableEntry(messageID string, authorUserID string, authorUsername string, referencedAuthorID string, referencedAuthorName string) ChallengeTableEntryStruct {
	ChallengeTableEntry := ChallengeTableEntryStruct{
		MessageID:       messageID,
		ChallengerID:    authorUserID,
		ChallengerName:  authorUsername,
		DefenderID:      referencedAuthorID,
		DefenderName:    referencedAuthorName,
		ChallengerVotes: 0,
		DefenderVotes:   0,
		AbstainVotes:    0,
		StopVotes:       0,
		Outcome:         0,
	}
	return ChallengeTableEntry
}

func selectChallengeRow(db *sqlx.DB, MessageID string) (ChallengeTableEntryStruct, error) {
	challengeRow := ChallengeTableEntryStruct{}
	err := db.Get(&challengeRow, "SELECT MessageID, ChallengerID, ChallengerName, DefenderID, DefenderName, ChallengerVotes, DefenderVotes, AbstainVotes, StopVotes, Outcome FROM challengeTable WHERE MessageID = ?", MessageID)
	return challengeRow, err
}

func selectVotes(db *sqlx.DB, MessageID string) (VotesStruct, error) {
	votes := VotesStruct{}
	err := db.Get(&votes, "SELECT ChallengerVotes, DefenderVotes, AbstainVotes, StopVotes FROM challengeTable WHERE MessageID = ?", MessageID)
	return votes, err
}

func updateVotes(db *sqlx.DB, MessageID string, votes VotesStruct) {
	query := "UPDATE challengeTable SET ChallengerVotes = ?, DefenderVotes = ?, AbstainVotes = ?, StopVotes = ? WHERE MessageID = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		oops(err, "prepare updateVotes")
		return
	}
	res, err := stmt.Exec(votes.ChallengerVotes, votes.DefenderVotes, votes.AbstainVotes, votes.StopVotes, MessageID)
	if err != nil {
		oops(err, "execute updateVotes")
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		oops(err, "RowsAffected")
		return
	}
	rowsAffected(rows, "updating votes")
	return
}

func updateOutcome(db *sqlx.DB, MessageID string, votes VotesStruct) {
	query := "UPDATE challengeTable SET Outcome = ? WHERE MessageID = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		oops(err, "prepare updateOutcome")
		return
	}
	var res sql.Result
	if votes.ChallengerVotes > votes.DefenderVotes {
		res, err = stmt.Exec(1, MessageID)
	}
	if votes.ChallengerVotes < votes.DefenderVotes {
		res, err = stmt.Exec(2, MessageID)
	}
	if votes.ChallengerVotes == votes.DefenderVotes {
		res, err = stmt.Exec(0, MessageID)
	}
	if err != nil {
		oops(err, "execute updateOutcome")
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		oops(err, "RowsAffected")
		return
	}
	rowsAffected(rows, "updating outcome")
	return
}

// CreateScoreboardTable this table stores results of challenge votes
func CreateScoreboardTable(db *sqlx.DB) error {
	query := "CREATE TABLE IF NOT EXISTS scoreboardTable(UserID text primary key, Username text, TotalChallengeWins int, TotalChallengeLosses int, TotalChallengeTies int, TotalChallenges int, SuccessfulChallenges int, FailedChallenges int, SuccessfulDefenses int, FailedDefenses int)"
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	res, err := db.ExecContext(ctx, query)
	if err != nil {
		oops(err, "CreateScoreboardTable")
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		oops(err, "RowsAffected")
		return err
	}
	rowsAffected(rows, "creating scoreboard table")
	return nil
}

func insertScoreboardRow(db *sqlx.DB, row ScoreboardTableEntryStruct) {
	query := "INSERT OR IGNORE INTO scoreboardTable (UserID, Username, TotalChallengeWins, TotalChallengeLosses, TotalChallengeTies, TotalChallenges, SuccessfulChallenges, FailedChallenges, SuccessfulDefenses, FailedDefenses) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	stmt, err := db.Prepare(query)
	if err != nil {
		oops(err, "prepare insertScoreboardRow")
		return
	}
	res, err := stmt.Exec(row.UserID, row.Username, row.TotalChallengeWins, row.TotalChallengeLosses, row.TotalChallengeTies, row.TotalChallenges, row.SuccessfulChallenges, row.FailedChallenges, row.SuccessfulDefenses, row.FailedDefenses)
	if err != nil {
		oops(err, "execute insertScoreboardRow")
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		oops(err, "RowsAffected")
		return
	}
	rowsAffected(rows, "inserting scoreboard row")
	return
}

func initScoreBoardRow(userID string, username string) ScoreboardTableEntryStruct {
	scoreboardTableEntry := ScoreboardTableEntryStruct{userID, username, 0, 0, 0, 0, 0, 0, 0, 0}
	return scoreboardTableEntry
}

func selectScoreboardRow(db *sqlx.DB, UserID string) (ScoreboardTableEntryStruct, error) {
	scoreboardRow := ScoreboardTableEntryStruct{}
	err := db.Get(&scoreboardRow, "SELECT UserID, Username, TotalChallengeWins, TotalChallengeLosses, TotalChallengeTies, TotalChallenges, SuccessfulChallenges, FailedChallenges, SuccessfulDefenses, FailedDefenses FROM scoreboardTable WHERE UserID = ?", UserID)
	return scoreboardRow, err
}

func updateScoreboard(db *sqlx.DB, scoreboardEntry ScoreboardTableEntryStruct) {
	query := "UPDATE scoreboardTable SET UserID = ?, Username = ?, TotalChallengeWins = ?, TotalChallengeLosses = ?, TotalChallengeTies = ?, TotalChallenges = ?, SuccessfulChallenges = ?, FailedChallenges = ?, SuccessfulDefenses = ?, FailedDefenses = ? WHERE UserID = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		oops(err, "prepare updateScoreboard")
		return
	}
	res, err := stmt.Exec(scoreboardEntry.UserID, scoreboardEntry.Username, scoreboardEntry.TotalChallengeWins, scoreboardEntry.TotalChallengeLosses, scoreboardEntry.TotalChallengeTies, scoreboardEntry.TotalChallenges, scoreboardEntry.SuccessfulChallenges, scoreboardEntry.FailedChallenges, scoreboardEntry.SuccessfulDefenses, scoreboardEntry.FailedDefenses, scoreboardEntry.UserID)
	if err != nil {
		oops(err, "execute updateScoreboard")
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		oops(err, "RowsAffected")
		return
	}
	rowsAffected(rows, "updating scoreboard")
	return
}

func userInScoreboard(db *sqlx.DB, UserID string) bool {
	query := "SELECT UserID FROM scoreboardTable WHERE UserID = ?"
	row := db.QueryRow(query, UserID)
	temp := ""
	err := row.Scan(&temp)
	if err != nil {
		oops(err, "scan")
	}
	if temp != "" {
		return true
	}
	return false
}

func winnerID(score ChallengeTableEntryStruct) string {
	if score.Outcome == 1 {
		return score.ChallengerID
	}
	if score.Outcome == 2 {
		return score.DefenderID
	}
	return "tie"
}

func scoreboardToString(s ScoreboardTableEntryStruct) string {
	score := "`" + s.Username + "\nTotal challenge wins: " + strconv.Itoa(s.TotalChallengeWins) + "\nTotal challenge losses: " + strconv.Itoa(s.TotalChallengeLosses) + "\nTotal challenge ties: " + strconv.Itoa(s.TotalChallengeTies) + "\nTotal challenges: " + strconv.Itoa(s.TotalChallenges) + "\nWins as challenger: " + strconv.Itoa(s.SuccessfulChallenges) + "\nLosses as challenger: " + strconv.Itoa(s.FailedChallenges) + "\nWins as defender: " + strconv.Itoa(s.SuccessfulDefenses) + "\nLosses as defender: " + strconv.Itoa(s.FailedDefenses) + "`"
	return score
}

// CreateVotingRecord this table stores users' votes on each challenge
func CreateVotingRecord(db *sqlx.DB) error {
	query := "CREATE TABLE IF NOT EXISTS votingRecord(UserID text, MessageID text, ChallengerVotes int, DefenderVotes int, AbstainVotes int, StopVotes int, PRIMARY KEY (UserID, MessageID))"
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	res, err := db.ExecContext(ctx, query)
	if err != nil {
		oops(err, "execute createVotingRecord")
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		oops(err, "RowsAffected")
		return err
	}
	rowsAffected(rows, "creating voting record")
	return nil
}

func removeVotingRecordRow(db *sqlx.DB, row VotingRecordEntryStruct) {
	query := "DELETE FROM votingRecord WHERE MessageID = ? AND UserID = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		oops(err, "prepare removeVotingRecordRow")
		return
	}
	res, err := stmt.Exec(row.MessageID, row.UserID)
	if err != nil {
		oops(err, "execute removeVotingRecordRow")
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		oops(err, "RowsAffected")
		return
	}
	rowsAffected(rows, "removing voting record row")
	return
}

func insertVotingRecordRow(db *sqlx.DB, row VotingRecordEntryStruct) {
	query := "INSERT OR IGNORE INTO votingRecord (UserID, MessageID, ChallengerVotes, DefenderVotes, AbstainVotes, StopVotes) VALUES (?, ?, ?, ?, ?, ?)"
	stmt, err := db.Prepare(query)
	if err != nil {
		oops(err, "prepare insertVotingRecordRow")
		return
	}
	res, err := stmt.Exec(row.UserID, row.MessageID, row.ChallengerVotes, row.DefenderVotes, row.AbstainVotes, row.StopVotes)
	if err != nil {
		oops(err, "execute insertVotingRecordRow")
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		oops(err, "RowsAffected")
		return
	}
	rowsAffected(rows, "inserting voting record row")
	return
}

func selectVotingRecordRow(db *sqlx.DB, UserID string, MessageID string) (VotingRecordEntryStruct, error) {
	votingRecordRow := VotingRecordEntryStruct{}
	err := db.Get(&votingRecordRow, "SELECT UserID, MessageID, ChallengerVotes, DefenderVotes, AbstainVotes, StopVotes FROM votingRecord WHERE UserID = ? AND MessageID = ?", UserID, MessageID)
	return votingRecordRow, err
}

func updateVotingRecord(db *sqlx.DB, VotingRecordEntry VotingRecordEntryStruct) {
	query := "UPDATE votingRecord SET ChallengerVotes = ?, DefenderVotes = ?, AbstainVotes = ?, StopVotes = ? WHERE MessageID = ? AND UserID = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		oops(err, "prepare updateVotingRecord")
		return
	}
	res, err := stmt.Exec(VotingRecordEntry.ChallengerVotes, VotingRecordEntry.DefenderVotes, VotingRecordEntry.AbstainVotes, VotingRecordEntry.StopVotes, VotingRecordEntry.MessageID, VotingRecordEntry.UserID)
	if err != nil {
		oops(err, "execute updateVotingRecord")
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		oops(err, "RowsAffected")
		return
	}
	rowsAffected(rows, "updating voting record")
	return
}

func hasVotedBlue(db *sqlx.DB, VotingRecordEntry VotingRecordEntryStruct) bool {
	rec, err := selectVotingRecordRow(db, VotingRecordEntry.UserID, VotingRecordEntry.MessageID)
	if err != nil {
		return false
	}
	if rec.ChallengerVotes > 0 {
		return true
	}
	return false
}

func hasVotedYellow(db *sqlx.DB, VotingRecordEntry VotingRecordEntryStruct) bool {
	rec, err := selectVotingRecordRow(db, VotingRecordEntry.UserID, VotingRecordEntry.MessageID)
	if err != nil {
		return false
	}
	if rec.DefenderVotes > 0 {
		return true
	}
	return false
}

func hasVotedRed(db *sqlx.DB, VotingRecordEntry VotingRecordEntryStruct) bool {
	rec, err := selectVotingRecordRow(db, VotingRecordEntry.UserID, VotingRecordEntry.MessageID)
	if err != nil {
		return false
	}
	if rec.AbstainVotes > 0 {
		return true
	}
	return false
}

func hasVotedStop(db *sqlx.DB, VotingRecordEntry VotingRecordEntryStruct) bool {
	rec, err := selectVotingRecordRow(db, VotingRecordEntry.UserID, VotingRecordEntry.MessageID)
	if err != nil {
		return false
	}
	if rec.StopVotes > 0 {
		return true
	}
	return false
}

func checkStopVotes(db *sqlx.DB, MessageID string) int {
	stopVotes := -1
	challengeRow, err := selectChallengeRow(db, MessageID)
	if err != nil {
		oops(err, "checkStopVotes")
		return stopVotes
	}
	stopVotes = challengeRow.StopVotes
	return stopVotes
}

func pushScore(db *sqlx.DB, challengeEntry ChallengeTableEntryStruct) {
	challengerID := challengeEntry.ChallengerID
	defenderID := challengeEntry.DefenderID
	challengerScoreboardRow, err := selectScoreboardRow(db, challengerID)
	if err != nil {
		oops(err, "selectScoreboardRow")
	}
	defenderScoreboardRow, err := selectScoreboardRow(db, defenderID)
	if err != nil {
		oops(err, "selectScoreboardRow")
	}
	if challengeEntry.Outcome == 1 {
		challengerScoreboardRow.SuccessfulChallenges += 1
		challengerScoreboardRow.TotalChallengeWins += 1
		challengerScoreboardRow.TotalChallenges += 1
		defenderScoreboardRow.TotalChallengeLosses += 1
		defenderScoreboardRow.FailedDefenses += 1
		defenderScoreboardRow.TotalChallenges += 1
		updateScoreboard(db, challengerScoreboardRow)
		updateScoreboard(db, defenderScoreboardRow)
	}
	if challengeEntry.Outcome == 2 {
		defenderScoreboardRow.SuccessfulDefenses += 1
		defenderScoreboardRow.TotalChallengeWins += 1
		defenderScoreboardRow.TotalChallenges += 1
		challengerScoreboardRow.FailedChallenges += 1
		challengerScoreboardRow.TotalChallengeLosses += 1
		challengerScoreboardRow.TotalChallenges += 1
		updateScoreboard(db, challengerScoreboardRow)
		updateScoreboard(db, defenderScoreboardRow)
	}
	if challengeEntry.Outcome == 0 {
		challengerScoreboardRow.TotalChallengeTies += 1
		challengerScoreboardRow.TotalChallenges += 1
		defenderScoreboardRow.TotalChallengeTies += 1
		defenderScoreboardRow.TotalChallenges += 1
		updateScoreboard(db, challengerScoreboardRow)
		updateScoreboard(db, defenderScoreboardRow)
	}
	return
}

//print in terminal

func printChallengeRow(row ChallengeTableEntryStruct) {
	MessageID := row.MessageID
	ChallengerID := row.ChallengerID
	ChallengerName := row.ChallengerName
	DefenderID := row.DefenderID
	DefenderName := row.DefenderName
	ChallengerVotes := strconv.Itoa(row.ChallengerVotes)
	DefenderVotes := strconv.Itoa(row.DefenderVotes)
	AbstainVotes := strconv.Itoa(row.AbstainVotes)
	Outcome := strconv.Itoa(row.Outcome)
	s := "-"
	log.Println("Challenge row values: " + MessageID + s + ChallengerID + s + ChallengerName + s + DefenderID + s + DefenderName + s + ChallengerVotes + s + DefenderVotes + s + AbstainVotes + s + Outcome)
}

func printScoreboardRow(r ScoreboardTableEntryStruct) {
	UserID := r.UserID
	Username := r.Username
	TotalChallengeWins := strconv.Itoa(r.TotalChallengeWins)
	TotalChallengeLosses := strconv.Itoa(r.TotalChallengeLosses)
	TotalChallengeTies := strconv.Itoa(r.TotalChallengeTies)
	TotalChallenges := strconv.Itoa(r.TotalChallenges)
	SuccessfulChallenges := strconv.Itoa(r.SuccessfulChallenges)
	FailedChallenges := strconv.Itoa(r.FailedChallenges)
	SuccessfulDefenses := strconv.Itoa(r.SuccessfulDefenses)
	FailedDefenses := strconv.Itoa(r.FailedDefenses)
	s := "-"
	log.Println(UserID + s + Username + s + TotalChallengeWins + s + TotalChallengeLosses + s + TotalChallengeTies + s + TotalChallenges + s + SuccessfulChallenges + s + FailedChallenges + s + SuccessfulDefenses + s + FailedDefenses)
}
