package app

import (
	"context"
	"testing"
	dbmocks "users/gen/mocks/users/pkg/postgresql"
	pubsubmocks "users/gen/mocks/users/pkg/pubsub"
	"users/internal/domain"

	"github.com/stretchr/testify/mock"
)

func Test_healthCheckQueries_Check(t *testing.T) {
	type fields struct {
		repo   domain.MonitoringRepoQueries
		pubsub domain.MonitoringRepoQueries
	}
	type args struct {
		ctx context.Context
	}
	mockDB := dbmocks.NewInterface(t)
	mockPubSub := pubsubmocks.NewInterface(t)

	tests := []struct {
		name          string
		fields        fields
		expectedMocks func()
		args          args
		want          bool
	}{
		{
			name: "success",
			fields: fields{
				repo:   mockDB,
				pubsub: mockPubSub,
			},
			expectedMocks: func() {
				mockDB.On("Ping", mock.Anything).Return(true).Once()
				mockPubSub.On("Ping", mock.Anything).Return(true).Once()
				mockPubSub.On("IsEnabled").Return(true).Once()
			},
			args: args{
				ctx: context.Background(),
			},
			want: true,
		},
		{
			name: "unhealthy repo",
			fields: fields{
				repo:   mockDB,
				pubsub: mockPubSub,
			},
			expectedMocks: func() {
				mockDB.On("Ping", mock.Anything).Return(false).Once()
			},
			args: args{
				ctx: context.Background(),
			},
			want: false,
		},
		{
			name: "unhealthy pubsub",
			fields: fields{
				repo:   mockDB,
				pubsub: mockPubSub,
			},
			expectedMocks: func() {
				mockDB.On("Ping", mock.Anything).Return(true).Once()
				mockPubSub.On("Ping", mock.Anything).Return(false).Once()
				mockPubSub.On("IsEnabled").Return(true).Once()
			},
			args: args{
				ctx: context.Background(),
			},
			want: false,
		},
		{
			name: "disabled pubsub",
			fields: fields{
				repo:   mockDB,
				pubsub: mockPubSub,
			},
			expectedMocks: func() {
				mockDB.On("Ping", mock.Anything).Return(true).Once()
				mockPubSub.On("IsEnabled").Return(false).Once()
			},
			args: args{
				ctx: context.Background(),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedMocks != nil {
				tt.expectedMocks()
			}
			h := NewHealthCheckQueries(tt.fields.repo, tt.fields.pubsub)
			if got := h.Check(tt.args.ctx); got != tt.want {
				t.Errorf("healthCheckQueries.Check() = %v, want %v", got, tt.want)
			}

		})
	}
}
