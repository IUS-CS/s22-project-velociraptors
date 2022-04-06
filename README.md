# Challenge Accepted

## Initial setup:

Initialize go modules:

    go mod init github.com/IUS-CS/s22-project-velociraptors

Install DiscordGo

    go get github.com/bwmarrin/discordgo


## To build:

    
## To run:

Assign token variable

    For Mac:
    export BOT_TOKEN=<TOKEN GOES HERE!!!!!>

    For Windows:
    Set-Variable -Name "BOT_TOKEN" -Value "<TOKEN GOES HERE!!!!!>"

Windows users download C compiler:

    https://jmeubank.github.io/tdm-gcc/


Run main.go with & pass it the token variable

    go run main.go -t $BOT_TOKEN

## What is it?
Challenge Accepted is a Discord bot with a scoreboard to keep track of who in the server is right/wrong most often.

## How does I use it?
You reply to a message in the channel with !challenge to start a challenge. Users then use emojis reactions to vote for the winner of the challenge.

The winner is chosen/updated in real time once at least 2 votes have been cast.

Other commands include !leaderboard to display the leaderboard and !score '@user' to display the mentioned user's score.

## How does the code work?
On startup, the bot creates a database with three tables:
	-challengeTable
	-scoreboardTable
    -votingRecord

Each challengeTable row stores the following information needed to initiate a vote for a single challenge:
	-messageID, the ID of the "!challenge" reply, a unique value to distinguish challenges from each other
	-challengerID, the ID of the user who replied '!challenge' to a message
	-challengerName, the username of the user who replied '!challenge'
	-defenderID, the ID of the user whose message was challenged
	-defenderName, the name of the user whose message was challenged
	-challengerVotes, the # of votes for the challenger
	-defenderVotes, the # of votes for the defender
	-abstainVotes, the # of abstain votes
	-outcome, the result of the votes (0=tie,1=challenger wins,2=defender wins)

Each scoreboardTable row stores the following information needed to track the results of challenges on the server for an individual user:
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

Each votingRecord row stores the following information needed to track who has already voted:
    -userID
    -messageID, the challenge that they voted in

A challenge begins when a user replies '!challenge' to a message on the server
A new challengeEntry row is inserted into the challenge table
	-challenger and defender info is stored and vote counts are set to 0
	-outcome starts as 0 (tie)
The bot sends a message to the same channel announcing the start of the challenge
When a vote reaction (blue, yellow or red square) is added to the message:
    -the user is added to the voteRecord for that challenge
    -the challengeEntry's vote counts are updated


TO DO:

If it is a user's first time participating in a challenge, a new scoreboardEntry row is inserted into the scoreboard table (one for each new participant)
	-user info is stored and initial values are 0 for all other fields

When the challengeEntry's vote counts are updated, the outcome field is updated for that entry
When the outcome field of a challenge is updated, the participating users' scoreboardEntry counts are updated
When a user sends a message that says '!checkScore <@username>', the tagged user's scoreboardEntry stats are sent in a message in the server


Potential features:
-function to set/announce time-limited votes
