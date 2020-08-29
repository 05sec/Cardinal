package utils

// These variable will be assigned in CI.
var (
	VERSION    string = "v0.7.0"
	COMMIT_SHA string
	BUILD_TIME string
)

const (
	// Config
	TITLE_CONF          = "title"
	FLAG_PREFIX_CONF    = "flag_prefix"
	FLAG_SUFFIX_CONF    = "flag_suffix"
	ANIMATE_ASTEROID    = "animate_asteroid"
	SHOW_OTHERS_GAMEBOX = "show_others_gamebox"

	BOOLEAN_TRUE  = "true"
	BOOLEAN_FALSE = "false"
)

const (
	// Config type
	STRING = iota
	BOOLEAN
)
