package main

type ComposeVolume struct {
	Type string
	Origin string
	Destination string
}

type ComposeEnv struct {
	Name string
	Value *string
}

type ComposePort struct {
	Protocol string
	Port string
	Target int
}

type ComposeService struct {
	Name string
	Image string
	Command []string
	Volumes []*ComposeVolume
	Env []*ComposeEnv
	Ports []*ComposePort
}

type Compose struct {
	Services []*ComposeService
}