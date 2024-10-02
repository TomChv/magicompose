package main

import (
	"context"
	"dagger/services/internal/dagger"
)

type Services struct{}


func (m *Services) Postgres(
	ctx context.Context,
	
	// +defaultPath="/postgres/postgres.conf"
	volume0 *dagger.File,
	
	// +defaultPath="/postgres"
	volume1 *dagger.Directory,
) (*dagger.Service, error) {
	return dag.Container().
		From("postgres:12").
		WithEnvVariable("POSTGRES_USER", "medplum").
		WithEnvVariable("POSTGRES_PASSWORD", "medplum").
		WithExposedPort(5432).
		WithMountedFile(
			"/usr/local/etc/postgres/postgres.conf", 
			volume0,
		).
		WithMountedDirectory(
			"/docker-entrypoint-initdb.d/", 
			volume1,
		).
		WithMountedCache("/var/lib/postgresql/data", dag.CacheVolume("postgres-postgres-data")).
		WithEntrypoint([]string{
			"postgres",
			"-c",
			"config_file=/usr/local/etc/postgres/postgres.conf",
		}).
	AsService().
	Start(ctx)
}
func (m *Services) Redis(
	ctx context.Context,
) (*dagger.Service, error) {
	return dag.Container().
		From("redis:7").
		WithExposedPort(6379).
		WithEntrypoint([]string{
			"redis-server",
			"--requirepass",
			"medplum",
		}).
	AsService().
	Start(ctx)
}
