# Architecture

## Discord Bot
The Discord bot creates two different databases; one for holding challenge data and one for holding user points. A challenger will mention the bot and someone else to challenge. If accepted, the bot will then announce the challenge and the participants and ask for votes. The votes are inserted into a database. Upon voting ending, a winner is determined and the challenge data database is updated.
![Capture](https://user-images.githubusercontent.com/98437411/160295372-2969502c-0940-404b-ab91-61c402d7ff70.PNG)
