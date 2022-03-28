package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

//variables used for command line parameters
var (
	Token string
)

const (
	//test
	testTrigger   = "!test"
	testTrigger2  = "!test2"
	testResponse  = "This is a statement you might disagree with."
	testResponse2 = "This is another statement you might disagree with."

	//for sqlite
	dbname = "scoreboardDB"

	//bot commands
	commandChallenge  = "!challenge"
	commandCheckScore = "!checkscore"

	//bot messages
	challengeMessage1 = " has challenged "
	challengeMessage2 = "Vote below to decide who's right!"
	challengeMessage3 = "\n\n🟦 = "
	challengeMessage4 = "\n🟨 = "
	challengeMessage5 = "\n🟥 = Abstain"

	//values
	maxIDLength = 18
)

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
}

type VotesStruct struct {
	ChallengerVotes int `db:"ChallengerVotes"`
	DefenderVotes   int `db:"DefenderVotes"`
	AbstainVotes    int `db:"AbstainVotes"`
}

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

//creates DB if it doesn't exist already
func dbConnection() (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", dbname)
	if err != nil {
		log.Printf("Error %s when opening database\n", err)
		return nil, err
	}
	return db, nil
}

//this table stores values for challenge votes
func createChallengeTable(db *sqlx.DB) error {
	query := "CREATE TABLE IF NOT EXISTS challengeTable(MessageID string primary key, ChallengerID text, ChallengerName text, DefenderID text, DefenderName text, ChallengerVotes int, DefenderVotes int, AbstainVotes int, Outcome int)"
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	res, err := db.ExecContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when creating challengeTable", err)
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when fetching rows affected during table creation", err)
		return err
	}
	log.Printf("%d rows affected when creating challengeTable", rows)
	return nil
}

func insertChallengeRow(db *sqlx.DB, row ChallengeTableEntryStruct) {
	query := "INSERT INTO challengeTable (MessageID, ChallengerID, ChallengerName, DefenderID, DefenderName, ChallengerVotes, DefenderVotes, AbstainVotes, Outcome) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Printf("Error %s while preparing insertChallengeRow query", err)
		return
	}
	res, err := stmt.Exec(row.MessageID, row.ChallengerID, row.ChallengerName, row.DefenderID, row.DefenderName, row.ChallengerVotes, row.DefenderVotes, row.AbstainVotes, row.Outcome)
	if err != nil {
		log.Printf("Error %s while executing insertChallengeRow query", err)
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when fetching rows affected during insertChallengeRow", err)
		return
	}
	log.Printf("%d rows affected when inserting challenge row", rows)
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
		Outcome:         0,
	}
	return ChallengeTableEntry
}

func selectChallengeRow(db *sqlx.DB, MessageID string) (ChallengeTableEntryStruct, error) {
	challengeRow := ChallengeTableEntryStruct{}
	err := db.Get(&challengeRow, "SELECT MessageID, ChallengerID, ChallengerName, DefenderID, DefenderName, ChallengerVotes, DefenderVotes, AbstainVotes, Outcome FROM challengeTable WHERE MessageID = ?", MessageID)
	return challengeRow, err
}

func selectVotes(db *sqlx.DB, MessageID string) (VotesStruct, error) {
	votes := VotesStruct{}
	err := db.Get(&votes, "SELECT ChallengerVotes, DefenderVotes, AbstainVotes FROM challengeTable WHERE MessageID = ?", MessageID)
	return votes, err
}

func updateVotes(db *sqlx.DB, MessageID string, votes VotesStruct) {
	query := "UPDATE challengeTable SET ChallengerVotes = ?, DefenderVotes = ?, AbstainVotes = ? WHERE MessageID = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Printf("Error %s while preparing updateVotes query", err)
		return
	}
	res, err := stmt.Exec(votes.ChallengerVotes, votes.DefenderVotes, votes.AbstainVotes, MessageID)
	if err != nil {
		log.Printf("Error %s while executing updateVotes query", err)
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when fetching rows affected during updateVotes", err)
		return
	}
	log.Printf("%d rows affected when updating votes", rows)
	return
}

func updateOutcome(db *sqlx.DB, MessageID string, votes VotesStruct) {
	query := "UPDATE challengeTable SET Outcome = ? WHERE MessageID = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Printf("Error %s while preparing updateOutcome query", err)
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
		log.Printf("Error %s while executing updateVotes query", err)
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when fetching rows affected during updateOutcome", err)
		return
	}
	log.Printf("%d rows affected when updating outcome", rows)
	return
}

//this table stores results of challenge votes
func createScoreboardTable(db *sqlx.DB) error {
	query := "CREATE TABLE IF NOT EXISTS scoreboardTable(UserID text primary key, Username text, TotalChallengeWins int, TotalChallengeLosses int, TotalChallengeTies int, TotalChallenges int, SuccessfulChallenges int, FailedChallenges int, SuccessfulDefenses int, FailedDefenses int)"
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	res, err := db.ExecContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when creating scoreboardTable", err)
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when fetching rows affected during table creation", err)
		return err
	}
	log.Printf("%d rows affected when creating scoreboardTable", rows)
	return nil
}

func insertScoreboardRow(db *sqlx.DB, row ScoreboardTableEntryStruct) {
	query := "INSERT OR IGNORE INTO scoreboardTable (UserID, Username, TotalChallengeWins, TotalChallengeLosses, TotalChallengeTies, TotalChallenges, SuccessfulChallenges, FailedChallenges, SuccessfulDefenses, FailedDefenses) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Printf("Error %s while preparing insertScoreboardRow query", err)
		return
	}
	res, err := stmt.Exec(row.UserID, row.Username, row.TotalChallengeWins, row.TotalChallengeLosses, row.TotalChallengeTies, row.TotalChallenges, row.SuccessfulChallenges, row.FailedChallenges, row.SuccessfulDefenses, row.FailedDefenses)
	if err != nil {
		log.Printf("Error %s while executing insertScoreboardRow query", err)
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when fetching rows affected during insertScoreboardRow", err)
		return
	}
	log.Printf("%d rows affected when inserting scoreboard row", rows)
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
		log.Printf("Error %s while preparing updateScoreboard query", err)
		return
	}
	res, err := stmt.Exec(scoreboardEntry.UserID, scoreboardEntry.Username, scoreboardEntry.TotalChallengeWins, scoreboardEntry.TotalChallengeLosses, scoreboardEntry.TotalChallengeTies, scoreboardEntry.TotalChallenges, scoreboardEntry.SuccessfulChallenges, scoreboardEntry.FailedChallenges, scoreboardEntry.SuccessfulDefenses, scoreboardEntry.FailedDefenses, scoreboardEntry.UserID)
	if err != nil {
		log.Printf("Error %s while executing updateScoreboard query", err)
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when fetching rows affected during updateScoreboard", err)
		return
	}
	log.Printf("%d rows affected when updating scoreboard", rows)
	printScoreboardRow(scoreboardEntry)
	return
}

func userInScoreboard(db *sqlx.DB, UserID string) bool {
	query := "SELECT UserID FROM scoreboardTable WHERE UserID = ?"
	row := db.QueryRow(query, UserID)
	temp := ""
	row.Scan(&temp)
	if temp != "" {
		return true
	}
	return false
}

func winnerID(db *sqlx.DB, score ChallengeTableEntryStruct) string {
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

//this table stores users' votes on each challenge
func createVotingRecord(db *sqlx.DB) error {
	query := "CREATE TABLE IF NOT EXISTS votingRecord(UserID text, MessageID text, ChallengerVotes int, DefenderVotes int, AbstainVotes int, PRIMARY KEY (UserID, MessageID))"
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	res, err := db.ExecContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when creating voting record", err)
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when fetching rows affected during table creation", err)
		return err
	}
	log.Printf("%d rows affected when creating votingRecord", rows)
	return nil
}

func insertVotingRecordRow(db *sqlx.DB, row VotingRecordEntryStruct) {
	query := "INSERT OR IGNORE INTO votingRecord (UserID, MessageID, ChallengerVotes, DefenderVotes, AbstainVotes) VALUES (?, ?, ?, ?, ?)"
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Printf("Error %s while preparing insertVotingRecordRow query", err)
		return
	}
	res, err := stmt.Exec(row.UserID, row.MessageID, row.ChallengerVotes, row.DefenderVotes, row.AbstainVotes)
	if err != nil {
		log.Printf("Error %s while executing insertVotingRecordRow query", err)
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when fetching rows affected during insertVotingRecordRow", err)
		return
	}
	log.Printf("%d rows affected when inserting votingRecord row", rows)
	return
}

func removeVotingRecordRow(db *sqlx.DB, row VotingRecordEntryStruct) {
	query := "DELETE FROM votingRecord WHERE MessageID = ? AND UserID = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Printf("Error %s while preparing removeVotingRecord query", err)
		return
	}
	res, err := stmt.Exec(row.MessageID, row.UserID)
	if err != nil {
		log.Printf("Error %s while executing removeVotingRecord query", err)
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when fetching rows affected during removeVotingRecord", err)
		return
	}
	log.Printf("%d rows affected when removing votingRecord", rows)
	return
}

func initVotingRecordRow(authorUserID string, messageID string) VotingRecordEntryStruct {
	VotingRecordEntry := VotingRecordEntryStruct{authorUserID, messageID, 0, 0, 0}
	return VotingRecordEntry
}

func selectVotingRecordRow(db *sqlx.DB, UserID string, MessageID string) (VotingRecordEntryStruct, error) {
	votingRecordRow := VotingRecordEntryStruct{}
	err := db.Get(&votingRecordRow, "SELECT UserID, MessageID, ChallengerVotes, DefenderVotes, AbstainVotes FROM votingRecord WHERE UserID = ? AND MessageID = ?", UserID, MessageID)
	return votingRecordRow, err
}

func updateVotingRecord(db *sqlx.DB, VotingRecordEntry VotingRecordEntryStruct) {
	query := "UPDATE votingRecord SET ChallengerVotes = ?, DefenderVotes = ?, AbstainVotes = ? WHERE MessageID = ? AND UserID = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Printf("Error %s while preparing updateVotingRecord query", err)
		return
	}
	res, err := stmt.Exec(VotingRecordEntry.ChallengerVotes, VotingRecordEntry.DefenderVotes, VotingRecordEntry.AbstainVotes, VotingRecordEntry.MessageID, VotingRecordEntry.UserID)
	if err != nil {
		log.Printf("Error %s while executing updateVotingRecord query", err)
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when fetching rows affected during updateVotingRecord", err)
		return
	}
	log.Printf("%d rows affected when updating votingRecord", rows)
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

func pushScore(db *sqlx.DB, challengeEntry ChallengeTableEntryStruct) {
	challengerID := challengeEntry.ChallengerID
	defenderID := challengeEntry.DefenderID
	challengerScoreboardRow, err := selectScoreboardRow(db, challengerID)
	if err != nil {
		log.Printf("Error %s while selecting challenger scoreboardRow for pushScore", err)
	}
	defenderScoreboardRow, err := selectScoreboardRow(db, defenderID)
	if err != nil {
		log.Printf("Error %s while selecting defender scoreboardRow for pushScore", err)
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

func printVotes(votes VotesStruct) {
	ChallengerVotes := strconv.Itoa(votes.ChallengerVotes)
	DefenderVotes := strconv.Itoa(votes.DefenderVotes)
	AbstainVotes := strconv.Itoa(votes.AbstainVotes)
	s := "-"
	log.Println(ChallengerVotes + s + DefenderVotes + s + AbstainVotes)
}

func printVotingRecordRow(v VotingRecordEntryStruct) {
	UserID := v.UserID
	MessageID := v.MessageID
	ChallengerVotes := strconv.Itoa(v.ChallengerVotes)
	DefenderVotes := strconv.Itoa(v.DefenderVotes)
	AbstainVotes := strconv.Itoa(v.AbstainVotes)
	s := "-"
	log.Println(UserID + s + MessageID + s + ChallengerVotes + s + DefenderVotes + s + AbstainVotes)
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

//trigger>response for messagecreate events
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	var messageContent = m.Content
	var messageType = m.Type

	//to send a message when m.Content == <whatever trigger you want>
	//follow this format (EqualFold compares strings, ignores case and returns True if they are equal):
	if strings.EqualFold(messageContent, testTrigger) {
		s.ChannelMessageSend(m.ChannelID, testResponse)
	}

	if strings.EqualFold(messageContent, testTrigger2) {
		s.ChannelMessageSend(m.ChannelID, testResponse2)
	}

	//!challenge
	if strings.EqualFold(messageContent, commandChallenge) && messageType == discordgo.MessageTypeReply {
		//connect to challengeDB
		db, err := dbConnection()
		if err != nil {
			log.Printf("Error %s when getting database connection", err)
			return
		}
		defer db.Close()

		authorUsername := m.Message.Author.Username
		authorUserID := m.Message.Author.ID
		referencedAuthorUsername := m.ReferencedMessage.Author.Username
		referencedAuthorID := m.ReferencedMessage.Author.ID

		challengerInfo := "<@" + authorUserID + ">" + challengeMessage1 + "<@" + referencedAuthorID + ">" + "!"
		debate := "\n\n<@" + referencedAuthorID + ">" + " says: `" + m.ReferencedMessage.Content + "`\n\n<@" + authorUserID + "> disagrees!\n"
		votingInfo := "\n" + challengeMessage2 + challengeMessage3 + "<@" + authorUserID + ">" + challengeMessage4 + "<@" + referencedAuthorID + ">" + challengeMessage5
		fullChallengeMessage := challengerInfo + debate + votingInfo
		announcementMessage, err := s.ChannelMessageSend(m.ChannelID, fullChallengeMessage)
		if err != nil {
			log.Printf("Error getting bot's message: %d", err)
			return
		}
		s.MessageReactionAdd(m.ChannelID, announcementMessage.ID, "🟦")
		s.MessageReactionAdd(m.ChannelID, announcementMessage.ID, "🟨")
		s.MessageReactionAdd(m.ChannelID, announcementMessage.ID, "🟥")
		s.MessageReactionAdd(m.ChannelID, announcementMessage.ID, "✋")
		announcementMessageID := announcementMessage.ID

		//create ChallengeTableEntry
		challengeTableEntry := initChallengeTableEntry(announcementMessageID, authorUserID, authorUsername, referencedAuthorID, referencedAuthorUsername)
		insertChallengeRow(db, challengeTableEntry)
		row, err := selectChallengeRow(db, announcementMessageID)
		if err != nil {
			log.Printf("Error %s while selecting challenge row", err)
		}
		printChallengeRow(row)

		//createScoreboardTableEntry x2 (one for challenger, one for defender)
		if !userInScoreboard(db, authorUserID) {
			scoreboardRow := initScoreBoardRow(authorUserID, authorUsername)
			insertScoreboardRow(db, scoreboardRow)
		}

		if !userInScoreboard(db, referencedAuthorID) {
			scoreboardRow := initScoreBoardRow(referencedAuthorID, referencedAuthorUsername)
			insertScoreboardRow(db, scoreboardRow)
		}
	}

	//!checkscore @username
	//connect to challengeDB
	db, err := dbConnection()
	if err != nil {
		log.Printf("Error %s when getting database connection", err)
		return
	}
	defer db.Close()
	var RegexUserPatternID *regexp.Regexp = regexp.MustCompile(fmt.Sprintf(`^(<@!(\d{%d,})>)$`, maxIDLength))
	parameters := strings.Split(messageContent, " ")
	if strings.EqualFold(parameters[0], commandCheckScore) && RegexUserPatternID.MatchString(parameters[1]) {
		mentionedUser := parameters[1]
		re, err := regexp.Compile(`[^\w]`)
		if err != nil {
			log.Fatal(err)
		}
		mentionedUser = re.ReplaceAllString(mentionedUser, "")
		mentionedScoreboard, err := selectScoreboardRow(db, mentionedUser)
		output := "<@" + mentionedUser + "> has the following challenge record:\n" + scoreboardToString(mentionedScoreboard)
		s.ChannelMessageSend(m.ChannelID, output)
	}

}

//trigger>response for messagereactionadd events
func messageReactionCreate(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	reactionEmoji := r.Emoji.Name
	messageID := r.MessageID
	reactionAuthorID := r.UserID

	//ignore all reactions created by the bot itself
	if r.UserID == s.State.User.ID {
		return
	}

	if reactionEmoji == "🛹" {
		log.Println("Skateboard detected")
	}

	if reactionEmoji == "🟦" {
		//connect to challengeDB
		db, err := dbConnection()
		if err != nil {
			log.Printf("Error %s when getting database connection", err)
			return
		}
		defer db.Close()
		votingRecordEntry, err := selectVotingRecordRow(db, reactionAuthorID, messageID)
		if hasVotedBlue(db, votingRecordEntry) || hasVotedYellow(db, votingRecordEntry) || hasVotedRed(db, votingRecordEntry) {
			log.Println("User has voted already")
			return
		}
		if !hasVotedBlue(db, votingRecordEntry) && !hasVotedRed(db, votingRecordEntry) && !hasVotedYellow(db, votingRecordEntry) {
			votingRecordEntry.UserID = reactionAuthorID
			votingRecordEntry.MessageID = messageID
			insertVotingRecordRow(db, votingRecordEntry)
		}
		votingRecordEntry.ChallengerVotes = 1
		updateVotingRecord(db, votingRecordEntry)
		votes, err := selectVotes(db, messageID)
		if err != nil {
			log.Printf("Error %s while selecting votes", err)
			return
		}
		ChallengerVotes := votes.ChallengerVotes + 1
		DefenderVotes := votes.DefenderVotes
		AbstainVotes := votes.AbstainVotes
		updatedVotes := VotesStruct{ChallengerVotes, DefenderVotes, AbstainVotes}
		updateVotes(db, messageID, updatedVotes)
		votes, err = selectVotes(db, messageID)
		if err != nil {
			log.Printf("Error %s while selecting votes", err)
			return
		}
		updateOutcome(db, messageID, votes)
		row, err := selectChallengeRow(db, messageID)
		printChallengeRow(row)
	}

	if reactionEmoji == "🟨" {
		//connect to challengeDB
		db, err := dbConnection()
		if err != nil {
			log.Printf("Error %s when getting database connection", err)
			return
		}
		defer db.Close()
		votingRecordEntry, err := selectVotingRecordRow(db, reactionAuthorID, messageID)
		if hasVotedYellow(db, votingRecordEntry) || hasVotedBlue(db, votingRecordEntry) || hasVotedRed(db, votingRecordEntry) {
			log.Println("User has voted already")
			return
		}
		if !hasVotedBlue(db, votingRecordEntry) && !hasVotedRed(db, votingRecordEntry) && !hasVotedYellow(db, votingRecordEntry) {
			votingRecordEntry.UserID = reactionAuthorID
			votingRecordEntry.MessageID = messageID
			insertVotingRecordRow(db, votingRecordEntry)
		}
		votingRecordEntry.DefenderVotes = 1
		updateVotingRecord(db, votingRecordEntry)
		votes, err := selectVotes(db, messageID)
		if err != nil {
			log.Printf("Error %s while selecting votes", err)
			return
		}
		ChallengerVotes := votes.ChallengerVotes
		DefenderVotes := votes.DefenderVotes + 1
		AbstainVotes := votes.AbstainVotes
		updatedVotes := VotesStruct{ChallengerVotes, DefenderVotes, AbstainVotes}
		updateVotes(db, messageID, updatedVotes)
		votes, err = selectVotes(db, messageID)
		if err != nil {
			log.Printf("Error %s while selecting votes", err)
			return
		}
		updateOutcome(db, messageID, votes)
		row, err := selectChallengeRow(db, messageID)
		printChallengeRow(row)
	}

	if reactionEmoji == "🟥" {
		//connect to challengeDB
		db, err := dbConnection()
		if err != nil {
			log.Printf("Error %s when getting database connection", err)
			return
		}
		defer db.Close()
		votingRecordEntry, err := selectVotingRecordRow(db, reactionAuthorID, messageID)
		if hasVotedRed(db, votingRecordEntry) || hasVotedBlue(db, votingRecordEntry) || hasVotedYellow(db, votingRecordEntry) {
			log.Println("User has voted already")
			return
		}
		if !hasVotedBlue(db, votingRecordEntry) && !hasVotedRed(db, votingRecordEntry) && !hasVotedYellow(db, votingRecordEntry) {
			votingRecordEntry.UserID = reactionAuthorID
			votingRecordEntry.MessageID = messageID
			insertVotingRecordRow(db, votingRecordEntry)
		}
		votingRecordEntry.AbstainVotes = 1
		updateVotingRecord(db, votingRecordEntry)
		votes, err := selectVotes(db, messageID)
		if err != nil {
			log.Printf("Error %s while selecting votes", err)
			return
		}
		ChallengerVotes := votes.ChallengerVotes
		DefenderVotes := votes.DefenderVotes
		AbstainVotes := votes.AbstainVotes + 1
		updatedVotes := VotesStruct{ChallengerVotes, DefenderVotes, AbstainVotes}
		updateVotes(db, messageID, updatedVotes)
		votes, err = selectVotes(db, messageID)
		if err != nil {
			log.Printf("Error %s while selecting votes", err)
			return
		}
		updateOutcome(db, messageID, votes)
		row, err := selectChallengeRow(db, messageID)
		printChallengeRow(row)
	}

	if reactionEmoji == "✋" {
		db, err := dbConnection()
		if err != nil {
			log.Printf("Error %s when getting database connection", err)
			return
		}
		defer db.Close()
		challengeEntry, err := selectChallengeRow(db, messageID)
		if err != nil {
			log.Printf("Error %s selecting challenge row in stop reaction", err)
			return
		}
		pushScore(db, challengeEntry)
		winnerIsChallenger := "\n<@" + challengeEntry.ChallengerID + "> has won the challenge!\n\nThe score was: " + strconv.Itoa(challengeEntry.ChallengerVotes) + " to " + strconv.Itoa(challengeEntry.DefenderVotes)
		winnerIsDefender := "\n<@" + challengeEntry.DefenderID + "> has won the challenge!\n\nThe score was: " + strconv.Itoa(challengeEntry.DefenderVotes) + " to " + strconv.Itoa(challengeEntry.ChallengerVotes)
		tie := "\nThe challenge between <@" + challengeEntry.ChallengerID + "> and <@" + challengeEntry.DefenderID + "> was a tie!"
		challengerRow, err := selectScoreboardRow(db, challengeEntry.ChallengerID)
		printScoreboardRow(challengerRow)
		defenderRow, err := selectScoreboardRow(db, challengeEntry.DefenderID)
		printScoreboardRow(defenderRow)
		if winnerID(db, challengeEntry) == "tie" {
			s.ChannelMessageSend(r.ChannelID, tie)
		}
		if winnerID(db, challengeEntry) == challengeEntry.ChallengerID {
			s.ChannelMessageSend(r.ChannelID, winnerIsChallenger)
		}
		if winnerID(db, challengeEntry) == challengeEntry.DefenderID {
			s.ChannelMessageSend(r.ChannelID, winnerIsDefender)
		}
	}
}

//trigger>response for messagereactionremove events
func messageReactionDelete(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
	reactionEmoji := r.Emoji.Name
	messageID := r.MessageID
	reactionAuthorID := r.UserID

	if reactionEmoji == "🛹" {
		log.Println("Skateboard removed")
	}

	if reactionEmoji == "🟦" {
		db, err := dbConnection()
		if err != nil {
			log.Printf("Error %s when getting database connection", err)
			return
		}
		defer db.Close()
		votingRecordEntry, err := selectVotingRecordRow(db, reactionAuthorID, messageID)
		if hasVotedBlue(db, votingRecordEntry) {
			removeVotingRecordRow(db, votingRecordEntry)
			votes, err := selectVotes(db, messageID)
			if err != nil {
				log.Printf("Error %s while selecting votes", err)
				return
			}
			ChallengerVotes := votes.ChallengerVotes - 1
			DefenderVotes := votes.DefenderVotes
			AbstainVotes := votes.AbstainVotes
			updatedVotes := VotesStruct{ChallengerVotes, DefenderVotes, AbstainVotes}
			updateVotes(db, messageID, updatedVotes)
			votes, err = selectVotes(db, messageID)
			if err != nil {
				log.Printf("Error %s while selecting votes", err)
				return
			}
			updateOutcome(db, messageID, votes)
			row, err := selectChallengeRow(db, messageID)
			printChallengeRow(row)
		}
	}

	if reactionEmoji == "🟨" {
		db, err := dbConnection()
		if err != nil {
			log.Printf("Error %s when getting database connection", err)
			return
		}
		defer db.Close()
		votingRecordEntry, err := selectVotingRecordRow(db, reactionAuthorID, messageID)
		if hasVotedYellow(db, votingRecordEntry) {
			removeVotingRecordRow(db, votingRecordEntry)
			votes, err := selectVotes(db, messageID)
			if err != nil {
				log.Printf("Error %s while selecting votes", err)
				return
			}
			ChallengerVotes := votes.ChallengerVotes
			DefenderVotes := votes.DefenderVotes - 1
			AbstainVotes := votes.AbstainVotes
			updatedVotes := VotesStruct{ChallengerVotes, DefenderVotes, AbstainVotes}
			updateVotes(db, messageID, updatedVotes)
			votes, err = selectVotes(db, messageID)
			if err != nil {
				log.Printf("Error %s while selecting votes", err)
				return
			}
			updateOutcome(db, messageID, votes)
			row, err := selectChallengeRow(db, messageID)
			printChallengeRow(row)
		}
	}

	if reactionEmoji == "🟥" {
		db, err := dbConnection()
		if err != nil {
			log.Printf("Error %s when getting database connection", err)
			return
		}
		defer db.Close()
		votingRecordEntry, err := selectVotingRecordRow(db, reactionAuthorID, messageID)
		if hasVotedRed(db, votingRecordEntry) {
			removeVotingRecordRow(db, votingRecordEntry)
			votes, err := selectVotes(db, messageID)
			if err != nil {
				log.Printf("Error %s while selecting votes", err)
				return
			}
			ChallengerVotes := votes.ChallengerVotes
			DefenderVotes := votes.DefenderVotes
			AbstainVotes := votes.AbstainVotes - 1
			updatedVotes := VotesStruct{ChallengerVotes, DefenderVotes, AbstainVotes}
			updateVotes(db, messageID, updatedVotes)
			votes, err = selectVotes(db, messageID)
			if err != nil {
				log.Printf("Error %s while selecting votes", err)
				return
			}
			updateOutcome(db, messageID, votes)
			row, err := selectChallengeRow(db, messageID)
			printChallengeRow(row)
		}
	}
}

func main() {
	//create a new Discord session using the provided bot token
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("Error creating Discord session,", err)
		return
	}
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentGuildMessageReactions
	//open a websocket connection to Discord and begin listening
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	//connect to scoreboardDB
	db, err := dbConnection()
	if err != nil {
		log.Printf("Error %s when getting database connection", err)
		return
	}
	defer db.Close()
	log.Printf("Successfully connected to database")
	//create tables
	err = createChallengeTable(db)
	if err != nil {
		log.Printf("createChallengeTable failed with error %s", err)
		return
	}
	err = createScoreboardTable(db)
	if err != nil {
		log.Printf("createScoreboardTable failed with error %s", err)
		return
	}
	err = createVotingRecord(db)
	if err != nil {
		log.Printf("createVotingRecord failed with error %s", err)
		return
	}

	//register messageCreate function as a callback for MessageCreate events
	dg.AddHandler(messageCreate)
	dg.AddHandler(messageReactionCreate)
	dg.AddHandler(messageReactionDelete)

	//everything runs here until one of the term signals is received
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	//close the Discord session
	dg.Close()
}
