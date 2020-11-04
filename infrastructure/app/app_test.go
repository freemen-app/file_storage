package app_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/freemen-app/file_storage/config"
	"github.com/freemen-app/file_storage/infrastructure/app"
)

var (
	conf *config.Config
)

func TestMain(m *testing.M) {
	conf = config.New(config.DefaultConfig)
	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	type fields struct {
		conf *config.Config
	}
	tests := []struct {
		name      string
		fields    fields
		wantPanic bool
	}{
		{
			name:      "succeed",
			fields:    fields{conf: conf},
			wantPanic: false,
		},
		{
			name:      "invalid config",
			fields:    fields{conf: &config.Config{}},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test := func() {
				application := app.New(tt.fields.conf)
				assert.NotNil(t, application.Repos())
				assert.NotNil(t, application.UseCases())
			}
			if tt.wantPanic {
				assert.Panics(t, test)
			} else {
				assert.NotPanics(t, test)
			}
		})
	}
}

func TestApp_Start_Shutdown(t *testing.T) {
	application := app.New(conf)

	assert.NoError(t, application.Start())
	assert.True(t, application.IsRunning())

	application.Shutdown()
	assert.False(t, application.IsRunning())
}
