package ascanius

// A source is a config values container
// This can be the environment, a file, a map, a struct ecc...
// anything that contains values that will be loaded inside of a config struct
type Source interface {
	// load the source values inside of a map
	Load() (map[string]any, error)

	// return the name
	Name() string

	// return the priority of the source
	Priority() int

	// set the priority of the source
	SetPriority(int)

	// set the name of the source
	SetName(string)

	// return the filetype or string type
	Type() string
}
