/*
-on startup, the bot creates a database with two tables:
	-challengeTable
	-scoreboardTable

-each challengeTable row stores the following information needed to initiate a vote for a single challenge:
	-challengeID, a unique value to distinguish challenges from each other
	-challengerID, the ID of the user who replied '!challenge' to a message
	-challengerName, the username of the user who replied '!challenge'
	-defenderID, the ID of the user whose message was challenged
	-defenderName, the name of the user whose message was challenged
	-challengerVotes, the # of votes for the challenger
	-defenderVotes, the # of votes for the defender
	-abstainVotes, the # of abstain votes
	-outcome, the result of the votes (0=tie,1=challenger wins,2=defender wins)

-each scoreboardTable row stores the following information needed to track the results of challenges on the server for an individual user:
	-userID, a unique value of the user whose results are stored in this entry
	-username
	-totalChallengeWins, count of all this user's wins, whether defending or challenging
	-totalChallengeLosses, count of all this user's losses, " " " "
	-totalChallengeTies, count of all this user's ties, " " " "
	-totalChallenges, wins + losses + ties
	-successfulChallenges, # of challenges where this user initiated the challenge and won
	-failedChallenges, # of challenges where this user initiated the challenge and lost
	-successfulDefenses, # of challenges where this user was challenged by someone else and won
	-failedDefenses, # of challenges where this user was challenged by someone else and lost

-a challenge begins when a user replies '!challenge' to a message on the server
-a new challengeEntry row is inserted into the challenge table
	-challenger and defender info is stored and vote counts are set to 0
	-outcome starts as 0 (tie)
-the bot sends a message to the same channel announcing the start of the challenge

TO DO:

-if it is a user's first time participating in a challenge, a new scoreboardEntry row is inserted into the scoreboard table (one for each new participant)
	-user info is stored and initial values are 0 for all other fields
-when a vote reaction (blue, yellow or red square) is added to the message, the challengeEntry's vote counts are updated
-when the challengeEntry's vote counts are updated, the outcome field is updated for that entry
-when the outcome field of a challenge is updated, the participating users' scoreboardEntry counts are updated
-when a user sends a message that says '!checkScore <@username>', the tagged user's scoreboardEntry stats are sent in a message in the server


Potential features:
-function to set/announce time-limited votes



*/

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
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
	Outcome         int    `db:"Outcome"`
	//0=tie, 1=challenger wins, 2=defender wins
}

//scoreboardTableEntryStruct fields, ready to be implemented with scoreboardTableEntry

// type scoreboardTableEntryStruct struct {
// 	UserID               string
// 	Username             string
// 	TotalChallengeWins   int
// 	TotalChallengeLosses int
// 	TotalChallengeTies   int
// 	TotalChallenges      int
// 	SuccessfulChallenges int
// 	FailedChallenges     int
// 	SuccessfulDefenses   int
// 	FailedDefenses       int
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
	query := "CREATE TABLE IF NOT EXISTS challengeTable(ChallengeID int primary key, ChallengerID varchar(50), ChallengerName varchar(50), DefenderID varchar(50), DefenderName varchar(50), ChallengerVotes int, DefenderVotes int, AbstainVotes int, Outcome int)"
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
	query := "CREATE TABLE IF NOT EXISTS scoreboardTable(UserID varchar(50) primary key, Username varchar(50), TotalChallengeWins int, TotalChallengeLosses int, TotalChallenges int, SuccessfulChallenges int, FailedChallenges int, SuccessfulDefenses int, FailedDefenses int)"
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
	query := "INSERT OR IGNORE INTO challengeTable(ChallengeID, ChallengerID, ChallengerName, DefenderID, DefenderName, ChallengerVotes, DefenderVotes, AbstainVotes, Outcome) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"
	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelfunc()
	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		log.Printf("Error %s when preparing SQL insertChallengeRow statement", err)
		return err
	}
	defer stmt.Close()
	res, err := stmt.ExecContext(ctx, t.ChallengeID, t.ChallengerID, t.ChallengerName, t.DefenderID, t.DefenderName, t.ChallengerVotes, t.DefenderVotes, t.AbstainVotes, t.Outcome)
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

//insertScoreboardRow ready to be implemented with scoreboardUpdate function

// func insertScoreboardRow(db *sqlx.DB, t scoreboardTableEntryStruct) error {
// 	query := "INSERT OR IGNORE INTO scoreboardTable(UserID, Username, TotalChallengeWins, TotalChallengeLosses, TotalChallengeTies, TotalChallenges, SuccessfulChallenges, FailedChallenges, SuccessfulDefenses, FailedDefenses) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
// 	ctx, cancelfunc := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancelfunc()
// 	stmt, err := db.PrepareContext(ctx, query)
// 	if err != nil {
// 		log.Printf("Error %s when preparing SQL insertScoreboardRow statement", err)
// 		return err
// 	}
// 	defer stmt.Close()
// 	res, err := stmt.ExecContext(ctx, t.UserID, t.Username, t.TotalChallengeWins, t.TotalChallengeLosses, t.TotalChallengeTies, t.TotalChallenges, t.SuccessfulChallenges, t.FailedChallenges, t.SuccessfulDefenses, t.FailedDefenses)
// 	if err != nil {
// 		log.Printf("Error %s when inserting row into ScoreboardTableEntryStruct", err)
// 		return err
// 	}
// 	rows, err := res.RowsAffected()
// 	if err != nil {
// 		log.Printf("Error %s when fetching rows affected while inserting row", err)
// 		return err
// 	}
// 	log.Printf("%d entries created", rows)
// 	return nil
// }

//selectChallengeRow function ready to be implemented with scoreboardUpdate function

// func selectChallengeRow(db *sqlx.DB, ChallengeID int) (ChallengeTableEntryStruct, error) {
// 	challengeRow := ChallengeTableEntryStruct{}
// 	err := db.Get(&challengeRow, "SELECT ChallengeID, ChallengerID, ChallengerName, DefenderID, DefenderName, ChallengerVotes, DefenderVotes, AbstainVotes, Outcome FROM challengeTable WHERE ChallengeID = ?", ChallengeID)
// 	return challengeRow, err
// }

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
		Outcome:         0,
	}
	incrementingChallengeID++
	return ChallengeTableEntry
}

//for testing purposes, prints in terminal
// func printChallengeRow(row ChallengeTableEntryStruct) {
// 	ChallengeID := strconv.Itoa(row.ChallengeID)
// 	ChallengerID := row.ChallengerID
// 	ChallengerName := row.ChallengerName
// 	DefenderID := row.DefenderID
// 	DefenderName := row.DefenderName
// 	ChallengerVotes := strconv.Itoa(row.ChallengerVotes)
// 	DefenderVotes := strconv.Itoa(row.DefenderVotes)
// 	AbstainVotes := strconv.Itoa(row.AbstainVotes)
// 	Outcome := strconv.Itoa(row.Outcome)
// 	s := "-"
// 	log.Println(ChallengeID + s + ChallengerID + s + ChallengerName + s + DefenderID + s + DefenderName + s + ChallengerVotes + s + DefenderVotes + s + AbstainVotes + s + Outcome)
// }

//trigger>response for messagecreate events
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
	// dg.AddHandler(messageReactionCreate)

	//everything runs here until one of the term signals is received
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	//close the Discord session
	dg.Close()
}
