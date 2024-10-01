package main

import (
	"bytes"
	"context"
	"dagger/magicompose/internal/dagger"
	"fmt"
	"text/template"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/iancoleman/strcase"
)

type Magicompose struct {
	File *dagger.File
}

func New(
	file *dagger.File,
) *Magicompose {
	return &Magicompose{
		File: file,
	}
}

func (m *Magicompose) Generate(ctx context.Context) (*dagger.Directory, error) {
	compose, err := m.Inspect(ctx)
	if err != nil {
		return nil, fmt.Errorf("Generate: failed to inspect compose file: %w", err)
	}

	templateDir := dag.CurrentModule().Source().Directory("template")
	templateFile, err := templateDir.File("main.go.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("Generate: failed to read template file: %w", err)
	}

	tmpl := template.New("compose").Funcs(template.FuncMap{
		"ToUpper": strcase.ToCamel,
		"IsBind": func(volume *ComposeVolume) bool {
			return volume.Type == "bind"
		},
		"IsCache": func(volume *ComposeVolume) bool {
			return volume.Type == "volume"
		},
	})


	tmpl, err = tmpl.Parse(templateFile)
	if err != nil {
		return nil, fmt.Errorf("Magicinit.initGo: failed to parse go template: %w", err)
	}


	var buf bytes.Buffer
	err = tmpl.Execute(&buf, compose)
	if err != nil {
		return nil, fmt.Errorf("Magicinit.initGo: failed to execute go template: %w", err)
	}

	return templateDir.WithNewFile("main.go", buf.String()).WithoutFile("main.go.tmpl"), nil
}

func (m *Magicompose) Inspect(ctx context.Context) (*Compose, error) {
	project, err := m.load(ctx)
	if err != nil {
		return nil, fmt.Errorf("Inspect: failed to load compose file: %w", err)
	}

	compose := &Compose{}

	for _, service := range project.Services {
		_service := &ComposeService{
			Name:    service.Name,
			Image:   service.Image,
			Command: service.Command,
		}

		for _, volume := range service.Volumes {
			_service.Volumes = append(_service.Volumes, &ComposeVolume{
				Type:        volume.Type,
				Origin:      volume.Source,
				Destination: volume.Target,
			})
		}

		for name, value := range service.Environment {
			_service.Env = append(_service.Env, &ComposeEnv{
				Name:  name,
				Value: value,
			})
		}

		for _, port := range service.Ports {
			_service.Ports = append(_service.Ports, &ComposePort{
				Protocol: port.Protocol,
				Port:     port.HostIP,
				Target:   int(port.Target),
			})
		}

		compose.Services = append(compose.Services, _service)
	}

	return compose, nil
}

func (m *Magicompose) load(ctx context.Context) (*types.Project, error) {
	content, err := m.File.Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("Load: failed to read compose file: %w", err)
	}

	// Define a config details object with the parsed YAML
	configDetails := types.ConfigDetails{
		ConfigFiles: []types.ConfigFile{
			{
				Content: []byte(content),
			},
			{
				Config: map[string]interface{}{
					"name": "magic",
				},
			},
		},
	}

	// Load the Compose project from the in-memory YAML
	project, err := loader.LoadWithContext(ctx, configDetails, loader.WithSkipValidation)
	if err != nil {
		return nil, fmt.Errorf("Load: failed to load compose file: %w", err)
	}

	return project, nil
}
