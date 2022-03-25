// needs select & update functions
// then scoreboard initialization,
// voting process on challenge table

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
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
	testTrigger  = "Does the bot work?"
	testResponse = "Vinny is alive and well."

	//for sqlite
	dbname = "scoreboardDB"

	//bot commands
	commandChallenge = "!challenge"

	//bot messages
	challengeMessage1 = " has challenged "
	challengeMessage2 = "Vote below to decide who's right!"
	challengeMessage3 = "\n\n🟦 = "
	challengeMessage4 = "\n🟨 = "
	challengeMessage5 = "\n🟥 = Abstain"
)

//ChallengeTableEntryStruct fields
type ChallengeTableEntryStruct struct {
	ChallengeID     int    `db:"ChallengeID"`
	ChallengerID    string `db:"ChallengerID"`
	ChallengerName  string `db:"ChallengerName"`
	DefenderID      string `db:"DefenderID"`
	DefenderName    string `db:"DefenderName"`
	ChallengerVotes int    `db:"ChallengerVotes"`
	DefenderVotes   int    `db:"DefenderVotes"`
	AbstainVotes    int    `db:"AbstainVotes"`
}

//scoreboardTable fields
// type scoreboardTable struct {
// 	userID               string
// 	userName             string
// 	totalChallengeWins   int
// 	totalChallengeLosses int
// 	totalChallengeTies   int
// 	totalChallenges      int
// 	successfulChallenges int
// 	failedChallenges     int
// 	successfulDefenses   int
// 	failedDefenses       int
// }

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
	query := "CREATE TABLE IF NOT EXISTS challengeTable(ChallengeID int primary key, ChallengerID varchar(50), ChallengerName varchar(50), DefenderID varchar(50), DefenderName varchar(50), ChallengerVotes int, DefenderVotes int, AbstainVotes int)"
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

//this table stores results of challenge votes
func createScoreboardTable(db *sqlx.DB) error {
	query := "CREATE TABLE IF NOT EXISTS scoreboardTable(userID varchar(50) primary key, userName varchar(50), totalChallengeWins int, totalChallengeLosses int, totalChallenges int, successfulChallenges int, failedChallenges int, successfulDefenses int, failedDefenses int, created_at datetime default CURRENT_TIMESTAMP, updated_at datetime default CURRENT_TIMESTAMP)"
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

func insertChallengeRow(db *sqlx.DB, t ChallengeTableEntryStruct) error {
	query := "INSERT OR IGNORE INTO challengeTable(ChallengeID, ChallengerID, ChallengerName, DefenderID, DefenderName, ChallengerVotes, DefenderVotes, AbstainVotes) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL insertChallengeRow statement", err)
		return err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, t.ChallengeID, t.ChallengerID, t.ChallengerName, t.DefenderID, t.DefenderName, t.ChallengerVotes, t.DefenderVotes, t.AbstainVotes)
	if err != nil {
		log.Printf("Error %s when inserting row into ChallengeTableEntryStruct", err)
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error %s when fetching rows affected while inserting row", err)
		return err
	}
	log.Printf("%d entries created", rows)
	return nil
}

// func insertScoreboardRow(db *sqlx.DB, t scoreboardTable) error {

func selectChallengeRow(db *sqlx.DB, ChallengeID int) (ChallengeTableEntryStruct, error) {
	challengeRow := ChallengeTableEntryStruct{}
	err := db.Get(&challengeRow, "SELECT ChallengeID, ChallengerID, ChallengerName, DefenderID, DefenderName, ChallengerVotes, DefenderVotes, AbstainVotes FROM challengeTable WHERE ChallengeID = ?", ChallengeID)
	return challengeRow, err
}

var incrementingChallengeID = 0

func initChallengeTableEntry(authorUserID string, authorUsername string, referencedAuthorID string, referencedAuthorName string) ChallengeTableEntryStruct {
	ChallengeTableEntry := ChallengeTableEntryStruct{
		ChallengeID:     incrementingChallengeID,
		ChallengerID:    authorUserID,
		ChallengerName:  authorUsername,
		DefenderID:      referencedAuthorID,
		DefenderName:    referencedAuthorName,
		ChallengerVotes: 0,
		DefenderVotes:   0,
		AbstainVotes:    0,
	}
	incrementingChallengeID++
	return ChallengeTableEntry
}

//for testing purposes, prints in terminal
func printChallengeRow(row ChallengeTableEntryStruct) {
	ChallengeID := strconv.Itoa(row.ChallengeID)
	ChallengerID := row.ChallengerID
	ChallengerName := row.ChallengerName
	DefenderID := row.DefenderID
	DefenderName := row.DefenderName
	ChallengerVotes := strconv.Itoa(row.ChallengerVotes)
	DefenderVotes := strconv.Itoa(row.DefenderVotes)
	AbstainVotes := strconv.Itoa(row.AbstainVotes)
	log.Println(ChallengeID + ChallengerID + ChallengerName + DefenderID + DefenderName + ChallengerVotes + DefenderVotes + AbstainVotes)
}

//trigger>response
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	var messageContent = m.Content
	var messageType = m.Type

	//ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	//to send a message when m.Content == <whatever trigger you want>
	//follow this format (EqualFold compares strings, ignores case and returns True if they are equal):
	if strings.EqualFold(messageContent, testTrigger) {
		s.ChannelMessageSend(m.ChannelID, testResponse)
	}

	//!challenge
	if strings.EqualFold(messageContent, commandChallenge) && messageType == discordgo.MessageTypeReply {
		//connect to challengeDB
		log.Println("Connecting to db")
		db, err := dbConnection()
		if err != nil {
			log.Printf("Error %s when getting database connection", err)
			return
		}
		defer db.Close()
		log.Println("Connected to db")

		authorUsername := m.Message.Author.Username
		authorUserID := m.Message.Author.ID
		referencedAuthorUsername := m.ReferencedMessage.Author.Username
		referencedAuthorID := m.ReferencedMessage.Author.ID
		//create ChallengeTableEntry
		challengeTableEntry := initChallengeTableEntry(authorUserID, authorUsername, referencedAuthorID, referencedAuthorUsername)
		log.Println("ChallengeTableEntry initialized")
		err = insertChallengeRow(db, challengeTableEntry)
		if err != nil {
			log.Printf("Insert challenge row failed with error %s", err)
			return
		}
		log.Println("ChallengeTableEntry inserted")
		//announce challenge in the channel

		challengerInfo := "<@" + authorUserID + ">" + challengeMessage1 + "<@" + referencedAuthorID + ">" + "!"

		debate := "\n\n<@" + referencedAuthorID + ">" + " says: `" + m.ReferencedMessage.Content + "`\n\n<@" + authorUserID + "> disagrees!\n"

		votingInfo := "\n" + challengeMessage2 + challengeMessage3 + "<@" + authorUserID + ">" + challengeMessage4 + "<@" + referencedAuthorID + ">" + challengeMessage5

		fullChallengeMessage := challengerInfo + debate + votingInfo
		announcementMessage, err := s.ChannelMessageSend(m.ChannelID, fullChallengeMessage)
		if err != nil {
			log.Printf("Error getting bot's message: %d", err)
			return
		}
		//add voting reactions
		s.MessageReactionAdd(m.ChannelID, announcementMessage.ID, "🟦")
		s.MessageReactionAdd(m.ChannelID, announcementMessage.ID, "🟨")
		s.MessageReactionAdd(m.ChannelID, announcementMessage.ID, "🟥")
	}

}

func main() {
	//create a new Discord session using the provided bot token
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("Error creating Discord session,", err)
		return
	}
	dg.Identify.Intents = discordgo.IntentsGuildMessages
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

	//register messageCreate function as a callback for MessageCreate events
	dg.AddHandler(messageCreate)

	//everything runs here until one of the term signals is received
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	//close the Discord session
	dg.Close()
}
