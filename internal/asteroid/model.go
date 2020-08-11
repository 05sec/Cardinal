package asteroid

import "github.com/vidar-team/Cardinal/internal/auth/team"

// greet will been sent when the client connect to the server firstly.
type greet struct {
	Title string
	Time  int
	Round int
	Team  []spaceShip
}

type spaceShip struct {
	Id    int
	Name  string
	Rank  int
	Image string
	Score int
}

type unityData struct {
	Type string
	Data interface{}
}

type attack struct {
	From int
	To   int
}

type rank struct {
	Team []team.Team
}

type status struct {
	Id     int
	Status string
}

type round struct {
	Round int
}

type clock struct {
	Time int
}

type clearStatus struct {
	Id int
}
