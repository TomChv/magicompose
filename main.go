package main

import (
	"bytes"
	"context"
	"dagger/magicompose/internal/dagger"
	"fmt"
	"text/template"
	"strings"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/iancoleman/strcase"
	"gopkg.in/yaml.v2"
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
	compose, err := m.inspect(ctx)
	if err != nil {
		return nil, fmt.Errorf("Generate: failed to inspect compose file: %w", err)
	}

	for _, service := range compose.Services {
		service.Volumes, err = m.parseRawVolumes(ctx, service.Name)
		if err != nil {
			return nil, fmt.Errorf("Generate: failed to parse volumes: %w", err)
		}
	}

	templateDir := dag.CurrentModule().Source().Directory("template")
	templateFile, err := templateDir.File("main.go.tmpl").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("Generate: failed to read template file: %w", err)
	}

	tmpl := template.New("compose").Funcs(template.FuncMap{
		"ToUpper": strcase.ToCamel,
		"IsBind": func(volume *composeVolume) bool {
			return volume.Type == "bind"
		},
		"IsCache": func(volume *composeVolume) bool {
			return volume.Type == "volume"
		},
		"VolumeToFileOrDirectory": func(volume *composeVolume) string {
			if strings.HasSuffix(volume.Origin, "/") {
				return "*dagger.Directory"
			}

			return "*dagger.File"
		},
		"ConvertPathToContextDir": func(path string) string {
			if strings.HasPrefix(path, "./") {
				path = strings.Replace(path, "./", "/", 1)
			}

			if strings.HasSuffix(path, "/") {
				path = strings.TrimSuffix(path, "/")
			}

			return path
		},
		"IsDir": func(volume *composeVolume) bool {
			return strings.HasSuffix(volume.Origin, "/")
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

func (m *Magicompose) inspect(ctx context.Context) (*compose, error) {
	project, err := m.load(ctx)
	if err != nil {
		return nil, fmt.Errorf("Inspect: failed to load compose file: %w", err)
	}

	compose := &compose{}

	for _, service := range project.Services {
		_service := &composeService{
			Name:    service.Name,
			Image:   service.Image,
			Command: service.Command,
		}

		for _, volume := range service.Volumes {
			_service.Volumes = append(_service.Volumes, &composeVolume{
				Type:        volume.Type,
				Origin:      volume.Source,
				Destination: volume.Target,
			})
		}

		for name, value := range service.Environment {
			_service.Env = append(_service.Env, &composeEnv{
				Name:  name,
				Value: value,
			})
		}

		for _, port := range service.Ports {
			_service.Ports = append(_service.Ports, &composePort{
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

func (m *Magicompose) parseRawVolumes(ctx context.Context, service string) ([]*composeVolume, error) {
	content, err := m.File.Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("Load: failed to read compose file: %w", err)
	}

	var parsedContent rawCompose
	err = yaml.Unmarshal([]byte(content), &parsedContent)
	if err != nil {
		return nil, fmt.Errorf("Load: failed to parse compose file: %w", err)
	}

	volumes, ok := parsedContent.Services[service]
	if !ok {
		return nil, fmt.Errorf("Load: failed to find service %s in compose file", service)
	}

	var result []*composeVolume
	for _, volume := range volumes.Volumes {
		if strings.HasPrefix(volume, "./") {
			paths := strings.Split(volume, ":")
			result = append(result, &composeVolume{
				Type:        "bind",
				Origin:      paths[0],
				Destination: paths[1],
			})
		} else {
			paths := strings.Split(volume, ":")

			result = append(result, &composeVolume{
				Type:        "volume",
				Origin:      paths[0],
				Destination: paths[1],
			})
		}
	}

	return result, nil
}
