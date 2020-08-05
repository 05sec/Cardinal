package main

var (
	VERSION    string
	COMMIT_SHA string
	BUILD_TIME string
)

func main() {
	s := new(Service)
	s.init()
}
