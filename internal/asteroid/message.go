package asteroid

const (
	INIT      = "init"
	ATTACK    = "attack"
	RANK      = "rank"
	STATUS    = "status"
	ROUND     = "round"
	EGG       = "easterEgg"
	TIME      = "time"
	CLEAR     = "clear"
	CLEAR_ALL = "clearAll"
)

var hub *Hub
var refresh func() Greet // Used to get title, team, score, time data.

// Init is used to init the asteroid. A function will be given to get the team rank data.
func Init(function func() Greet) {
	refresh = function
	hub = newHub()

	// Start to handle the request.
	go hub.run()
}

// NewRoundAction runs in the new round begin.
// Refresh rank, clean all gameboxes' status, set round text, set time text.
func NewRoundAction() {
	sendRank()
	sendClearAll()
	sendRound(refresh().Round)
	sendTime(refresh().Time)
}

// SendStatus sends the teams' status message.
func SendStatus(team int, statusString string) {
	sendStatus(team, statusString)
}

// SendAttack sends an attack action message.
func SendAttack(from int, to int) {
	sendAttack(from, to)
}

// sendAttack sends an attack action message.
func sendAttack(from int, to int) {
	hub.sendMessage(ATTACK, attack{
		From: from,
		To:   to,
	})
}

// sendRank sends the team rank list message.
func sendRank() {
	hub.sendMessage(RANK, rank{Team: refresh().Team})
}

// sendStatus sends the teams' status message.
func sendStatus(team int, statusString string) {
	hub.sendMessage(STATUS, status{
		Id:     team,
		Status: statusString,
	})
}

// sendRound sends now round.
func sendRound(roundNumber int) {
	hub.sendMessage(ROUND, round{Round: roundNumber})
}

// sendEasterEgg can send a meteorite!!
func sendEasterEgg() {
	hub.sendMessage(EGG, nil)
}

// sendTime sends time message.
func sendTime(time int) {
	hub.sendMessage(TIME, clock{Time: time})
}

// sendClear removes the status of the team.
func sendClear(team int) {
	hub.sendMessage(CLEAR, clearStatus{Id: team})
}

// sendClearAll removes all the teams' status.
func sendClearAll() {
	hub.sendMessage(CLEAR_ALL, nil)
}
