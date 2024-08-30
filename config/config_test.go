package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	// create temp file
	invalidConfig := []byte("invalid_yaml: -")
	validConfig := []byte(`
app:
  name: 'users'
  version: '1.0.0'
  log_level: 'debug'

http:
  port: 8080

grpc:
  port: 8081

postgres:
  pool_max: 2
  dsn: something

pubsub:
  enabled: true
  project_id: users-project
  users_topic: users

notifications:
  interval: 30
  batch_size_max: 50
  `)
	invalidTmpFile, err := os.CreateTemp("", "invalid_config.yaml")
	assert.NoError(t, err)
	defer os.Remove(invalidTmpFile.Name())

	validTmpFile, err := os.CreateTemp("", "config*.yaml")
	assert.NoError(t, err)
	defer os.Remove(validTmpFile.Name())

	_, err = invalidTmpFile.Write(invalidConfig)
	assert.NoError(t, err)
	_, err = validTmpFile.Write(validConfig)
	assert.NoError(t, err)

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr error
	}{
		{
			name: "no path provided",
			args: args{
				path: "",
			},
			want:    nil,
			wantErr: fmt.Errorf("config path not provided"),
		}, {
			name: "config error",
			args: args{
				path: invalidTmpFile.Name(),
			},
			want:    nil,
			wantErr: fmt.Errorf("config error: file format '%s' doesn't supported by the parser", filepath.Ext(invalidTmpFile.Name())),
		},
		{
			name: "config success",
			args: args{
				path: validTmpFile.Name(),
			},
			want: &Config{
				App:           App{Name: "users", Version: "1.0.0", LogLevel: "debug"},
				HTTP:          HTTP{Port: 8080},
				GRPC:          GRPC{Port: 8081},
				PG:            PG{PoolMax: 2, DSN: "something"},
				PubSub:        PubSub{Enabled: true, ProjectID: "users-project", UsersTopic: "users"},
				Notifications: Notifications{MaxBatchSize: 50, Interval: 30},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConfig(tt.args.path)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
