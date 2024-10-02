package main

type composeVolume struct {
	Type string
	Origin string
	Destination string
}

type composeEnv struct {
	Name string
	Value *string
}

type composePort struct {
	Protocol string
	Port string
	Target int
}

type composeService struct {
	Name string
	Image string
	Command []string
	Volumes []*composeVolume
	Env []*composeEnv
	Ports []*composePort
}

type compose struct {
	Services []*composeService
}

type rawCompose struct {
	Services map[string]struct {
		Volumes []string `yaml:"volumes"`
	} `yaml:"services"`
}