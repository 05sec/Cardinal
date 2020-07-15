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

// InitAsteroid is used to init the asteroid. A function will be given to get the team rank data.
func InitAsteroid(function func() Greet) {
	refresh = function
	hub = newHub()

	// Start to handle the request.
	go hub.run()
}

func Attack(from int, to int) {
	hub.sendMessage(ATTACK, attack{
		From: from,
		To:   to,
	})
}

func Rank() {
	hub.sendMessage(RANK, rank{Team: refresh().Team})
}

func Status(team int, statusString string) {
	hub.sendMessage(STATUS, status{
		Id:     team,
		Status: statusString,
	})
}

func Round(roundNumber int) {
	hub.sendMessage(ROUND, round{Round: roundNumber})
}

func EasterEgg() {
	hub.sendMessage(EGG, nil)
}

func Time(time int) {
	hub.sendMessage(TIME, clock{Time: time})
}

func Clear(team int) {
	hub.sendMessage(CLEAR, clearStatus{Id: team})
}

func ClearAll() {
	hub.sendMessage(CLEAR_ALL, nil)
}
