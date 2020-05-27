package main

var (
	VERSION    = "v0.5.0"
	COMMIT_SHA string
	BUILD_TIME string
)

func main() {
	s := new(Service)
	s.init()
}
