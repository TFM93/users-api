package notification

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	loggerMocks "users/gen/mocks/users/pkg/logger"
	"users/internal/domain"
	log "users/pkg/logger"

	"github.com/stretchr/testify/assert"
)

func TestNotificationService_Publish(t *testing.T) {
	mockedLogger := loggerMocks.NewInterface(t)

	eventPayload, err := json.Marshal(map[string]interface{}{
		"id": "abc",
	})
	assert.NoError(t, err)

	type fields struct {
		l log.Interface
	}
	type args struct {
		event domain.Event
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		expectedMocks func(*loggerMocks.Interface)
		wantErr       error
	}{
		{
			name: "success",
			fields: fields{
				l: mockedLogger,
			},
			args: args{
				event: domain.Event{
					Type:    "as",
					Payload: eventPayload,
				},
			},
			expectedMocks: func(i *loggerMocks.Interface) {
				i.On("Info", "Published: Type: %s | Payload: %s", "as", "{\"id\":\"abc\"}").Return().Once()
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &loggerNotifier{
				l: tt.fields.l,
			}
			tt.expectedMocks(mockedLogger)
			err := n.Publish(context.TODO(), &tt.args.event)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error())
			}
		})
	}
}

func TestNewNotificationService(t *testing.T) {
	mockedLogger := loggerMocks.NewInterface(t)
	type args struct {
		logger log.Interface
	}
	tests := []struct {
		name string
		args args
		want *loggerNotifier
	}{
		{name: "success", args: args{mockedLogger}, want: &loggerNotifier{mockedLogger}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewLoggerNotifierService(tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNotificationService() = %v, want %v", got, tt.want)
			}
		})
	}
}
