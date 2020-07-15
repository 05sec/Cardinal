package asteroid

// Team is the team status struct.
type Team struct {
	Id    int
	Name  string
	Rank  int
	Score int
}

// Greet will been sent when the client connect to the server firstly.
type Greet struct {
	Title string
	Time  int
	Round int
	Team  []Team
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
	Team []Team
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
