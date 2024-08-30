package grpc

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"
	appmocks "users/gen/mocks/users/app"
	loggermocks "users/gen/mocks/users/pkg/logger"
	gen "users/gen/proto/go"
	"users/internal/app/user"
	"users/internal/domain"

	"github.com/bufbuild/protovalidate-go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestUserServerImpl_CreateUser(t *testing.T) {
	expectedUserID := "0f913f6a-497b-4305-b3d1-3f53657e3a25"

	mockServiceCommands := appmocks.NewUserServiceCommands(t)
	mockLogger := loggermocks.NewInterface(t)
	protoValidator, err := protovalidate.New()
	assert.NoError(t, err, "creating protovalidator instance")

	server := &UserHandler{
		l:               mockLogger,
		serviceCommands: mockServiceCommands,
		protoValidator:  protoValidator,
	}

	type args struct {
		ctx context.Context
		cur *gen.CreateUserRequest
	}
	tests := []struct {
		name          string
		args          args
		expectedMocks func(ctx context.Context)
		want          *gen.UserID
		wantErr       error
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				cur: &gen.CreateUserRequest{
					FirstName:      "first",
					LastName:       "last",
					NickName:       "nick",
					CountryIsoCode: "UK",
					Email:          "something@something.pt",
					Password:       "serverKnows",
				},
			},
			expectedMocks: func(ctx context.Context) {
				mockServiceCommands.On("CreateUser", ctx, user.AddUserRequest{
					FirstName: "first", LastName: "last",
					NickName: "nick", Email: "something@something.pt",
					Password: "serverKnows", CountryISOCode: "UK"}).Return(expectedUserID, nil).Once()
			},
			want: &gen.UserID{
				Id: expectedUserID,
			},
			wantErr: nil,
		},
		{
			name: "invalid email",
			args: args{
				ctx: context.Background(),
				cur: &gen.CreateUserRequest{
					FirstName:      "first",
					LastName:       "last",
					NickName:       "nick",
					CountryIsoCode: "UK",
					Email:          "something@",
					Password:       "serverKnows",
				},
			},
			want:    nil,
			wantErr: fmt.Errorf("validation error:\n - email: value must be a valid email address [string.email]"),
		},
		{
			name: "invalid password",
			args: args{
				ctx: context.Background(),
				cur: &gen.CreateUserRequest{
					FirstName:      "first",
					LastName:       "last",
					NickName:       "nick",
					CountryIsoCode: "UK",
					Email:          "something@xpto.pt",
					Password:       "p",
				},
			},
			want:    nil,
			wantErr: fmt.Errorf("validation error:\n - password: value length must be at least 6 characters [string.min_len]"),
		},
		{
			name: "invalid nickname",
			args: args{
				ctx: context.Background(),
				cur: &gen.CreateUserRequest{
					FirstName:      "first",
					LastName:       "last",
					NickName:       "1",
					CountryIsoCode: "UK",
					Email:          "something@xpto.pt",
					Password:       "serverKnows",
				},
			},
			want:    nil,
			wantErr: fmt.Errorf("validation error:\n - nick_name: value length must be at least 3 characters [string.min_len]"),
		},
		{
			name: "invalid first name",
			args: args{
				ctx: context.Background(),
				cur: &gen.CreateUserRequest{
					FirstName:      "1",
					LastName:       "last",
					NickName:       "nick",
					CountryIsoCode: "UK",
					Email:          "something@xpto.pt",
					Password:       "serverKnows",
				},
			},
			want:    nil,
			wantErr: fmt.Errorf("validation error:\n - first_name: value length must be at least 3 characters [string.min_len]"),
		},
		{
			name: "invalid last name",
			args: args{
				ctx: context.Background(),
				cur: &gen.CreateUserRequest{
					FirstName:      "first",
					LastName:       "1",
					NickName:       "nick",
					CountryIsoCode: "UK",
					Email:          "something@xpto.pt",
					Password:       "serverKnows",
				},
			},
			want:    nil,
			wantErr: fmt.Errorf("validation error:\n - last_name: value length must be at least 3 characters [string.min_len]"),
		},
		{
			name: "invalid country code",
			args: args{
				ctx: context.Background(),
				cur: &gen.CreateUserRequest{
					FirstName:      "first",
					LastName:       "last",
					NickName:       "nick",
					CountryIsoCode: "UK1",
					Email:          "something@xpto.pt",
					Password:       "serverKnows",
				},
			},
			want:    nil,
			wantErr: fmt.Errorf("validation error:\n - country_iso_code: value length must be 2 characters [string.len]"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedMocks != nil {
				tt.expectedMocks(tt.args.ctx)
			}
			got, err := server.CreateUser(tt.args.ctx, tt.args.cur)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserServerImpl.CreateUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserServerImpl_DeleteUser(t *testing.T) {
	expectedUserID := "0f913f6a-497b-4305-b3d1-3f53657e3a25"

	mockServiceCommands := appmocks.NewUserServiceCommands(t)
	mockLogger := loggermocks.NewInterface(t)
	protoValidator, err := protovalidate.New()
	assert.NoError(t, err, "creating protovalidator instance")

	server := &UserHandler{
		l:               mockLogger,
		serviceCommands: mockServiceCommands,
		protoValidator:  protoValidator,
	}

	type args struct {
		ctx context.Context
		cur *gen.UserID
	}
	tests := []struct {
		name          string
		args          args
		expectedMocks func(ctx context.Context)
		want          *gen.UserID
		wantErr       error
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				cur: &gen.UserID{
					Id: expectedUserID,
				},
			},
			expectedMocks: func(ctx context.Context) {
				mockServiceCommands.On("DeleteUser", ctx, expectedUserID).Return(nil).Once()
			},
			want: &gen.UserID{
				Id: expectedUserID,
			},
			wantErr: nil,
		},
		{
			name: "empty id",
			args: args{
				ctx: context.Background(),
				cur: &gen.UserID{
					Id: "",
				},
			},
			want:    nil,
			wantErr: fmt.Errorf("validation error:\n - id: value is empty, which is not a valid UUID [string.uuid_empty]"),
		},
		{
			name: "invalid id",
			args: args{
				ctx: context.Background(),
				cur: &gen.UserID{
					Id: "something wrong",
				},
			},
			want:    nil,
			wantErr: fmt.Errorf("validation error:\n - id: value must be a valid UUID [string.uuid]"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedMocks != nil {
				tt.expectedMocks(tt.args.ctx)
			}
			got, err := server.DeleteUser(tt.args.ctx, tt.args.cur)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserServerImpl.DeleteUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserServerImpl_UpdateUser(t *testing.T) {
	expectedUserID := "0f913f6a-497b-4305-b3d1-3f53657e3a25"

	mockServiceCommands := appmocks.NewUserServiceCommands(t)
	mockLogger := loggermocks.NewInterface(t)
	protoValidator, err := protovalidate.New()
	assert.NoError(t, err, "creating protovalidator instance")

	server := &UserHandler{
		l:               mockLogger,
		serviceCommands: mockServiceCommands,
		protoValidator:  protoValidator,
	}

	type args struct {
		ctx context.Context
		cur *gen.UpdateUserRequest
	}
	tests := []struct {
		name          string
		args          args
		expectedMocks func(ctx context.Context)
		want          *gen.UserID
		wantErr       error
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				cur: &gen.UpdateUserRequest{Id: expectedUserID, User: &gen.EditableUserFields{
					FirstName:      "first",
					LastName:       "last",
					NickName:       "nick",
					CountryIsoCode: "UK",
					Email:          "something@something.pt",
				},
				},
			},
			expectedMocks: func(ctx context.Context) {
				mockServiceCommands.On("UpdateUser", ctx, user.UpdateUserRequest{
					ID:        "0f913f6a-497b-4305-b3d1-3f53657e3a25",
					FirstName: "first", LastName: "last",
					NickName: "nick", Email: "something@something.pt",
					CountryISOCode: "UK"}).Return(nil).Once()
			},
			want: &gen.UserID{
				Id: expectedUserID,
			},
			wantErr: nil,
		},
		{
			name: "invalid email",
			args: args{
				ctx: context.Background(),
				cur: &gen.UpdateUserRequest{Id: expectedUserID, User: &gen.EditableUserFields{
					FirstName:      "first",
					LastName:       "last",
					NickName:       "nick",
					CountryIsoCode: "UK",
					Email:          "something@.pt",
				},
				},
			},
			want:    nil,
			wantErr: fmt.Errorf("validation error:\n - user.email: value must be a valid email address [string.email]"),
		},
		{
			name: "invalid first name",
			args: args{
				ctx: context.Background(),
				cur: &gen.UpdateUserRequest{Id: expectedUserID, User: &gen.EditableUserFields{
					FirstName:      "f",
					LastName:       "last",
					NickName:       "nick",
					CountryIsoCode: "UK",
					Email:          "something@something.pt",
				},
				},
			},
			want:    nil,
			wantErr: fmt.Errorf("validation error:\n - user.first_name: value length must be at least 3 characters [string.min_len]"),
		},
		{
			name: "invalid last name",
			args: args{
				ctx: context.Background(),
				cur: &gen.UpdateUserRequest{Id: expectedUserID, User: &gen.EditableUserFields{
					FirstName:      "first",
					LastName:       "l",
					NickName:       "nick",
					CountryIsoCode: "UK",
					Email:          "something@something.pt",
				},
				},
			},
			want:    nil,
			wantErr: fmt.Errorf("validation error:\n - user.last_name: value length must be at least 3 characters [string.min_len]"),
		},
		{
			name: "invalid nick name",
			args: args{
				ctx: context.Background(),
				cur: &gen.UpdateUserRequest{Id: expectedUserID, User: &gen.EditableUserFields{
					FirstName:      "first",
					LastName:       "last",
					NickName:       "n",
					CountryIsoCode: "UK",
					Email:          "something@something.pt",
				},
				},
			},
			want:    nil,
			wantErr: fmt.Errorf("validation error:\n - user.nick_name: value length must be at least 3 characters [string.min_len]"),
		},
		{
			name: "invalid country code",
			args: args{
				ctx: context.Background(),
				cur: &gen.UpdateUserRequest{Id: expectedUserID, User: &gen.EditableUserFields{
					FirstName:      "first",
					LastName:       "last",
					NickName:       "nick",
					CountryIsoCode: "U",
					Email:          "something@something.pt",
				},
				},
			},
			want:    nil,
			wantErr: fmt.Errorf("validation error:\n - user.country_iso_code: value length must be 2 characters [string.len]"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedMocks != nil {
				tt.expectedMocks(tt.args.ctx)
			}
			got, err := server.UpdateUser(tt.args.ctx, tt.args.cur)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserServerImpl.UpdateUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserServerImpl_GetUser(t *testing.T) {
	timeFreeze := time.Now()
	expectedUserID := "0f913f6a-497b-4305-b3d1-3f53657e3a25"

	mockServiceQueries := appmocks.NewUserServiceQueries(t)
	mockLogger := loggermocks.NewInterface(t)
	protoValidator, err := protovalidate.New()
	assert.NoError(t, err, "creating protovalidator instance")

	server := &UserHandler{
		l:              mockLogger,
		serviceQueries: mockServiceQueries,
		protoValidator: protoValidator,
	}

	type args struct {
		ctx context.Context
		cur *gen.UserID
	}
	tests := []struct {
		name          string
		args          args
		expectedMocks func(ctx context.Context)
		want          *gen.UserResponse
		wantErr       error
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				cur: &gen.UserID{Id: expectedUserID},
			},
			expectedMocks: func(ctx context.Context) {
				mockServiceQueries.On("GetUser", ctx, expectedUserID).Return(&domain.User{
					ID:             uuid.MustParse(expectedUserID),
					FirstName:      "first",
					LastName:       "last",
					NickName:       "nick",
					CountryISOCode: "UK",
					Email:          "something",
					Password:       "",
					CreatedAt:      timeFreeze,
					UpdatedAt:      timeFreeze,
				}, nil).Once()
			},
			want: &gen.UserResponse{
				User: &gen.ReadableUserFields{
					Id:             expectedUserID,
					FirstName:      "first",
					LastName:       "last",
					NickName:       "nick",
					CountryIsoCode: "UK",
					Email:          "something",
					CreatedAt:      timestamppb.New(timeFreeze),
					UpdatedAt:      timestamppb.New(timeFreeze),
				},
			},
			wantErr: nil,
		},
		{
			name: "service layer error",
			args: args{
				ctx: context.Background(),
				cur: &gen.UserID{Id: expectedUserID},
			},
			expectedMocks: func(ctx context.Context) {
				mockServiceQueries.On("GetUser", ctx, expectedUserID).Return(nil, domain.ErrInvalidUserID).Once()
			},
			want:    &gen.UserResponse{},
			wantErr: domain.ErrInvalidUserID,
		},
		{
			name: "empty user returned",
			args: args{
				ctx: context.Background(),
				cur: &gen.UserID{Id: expectedUserID},
			},
			expectedMocks: func(ctx context.Context) {
				mockServiceQueries.On("GetUser", ctx, expectedUserID).Return(nil, nil).Once()
			},
			want:    &gen.UserResponse{},
			wantErr: nil,
		},
		{
			name: "empty id",
			args: args{
				ctx: context.Background(),
				cur: &gen.UserID{
					Id: "",
				},
			},
			want:    nil,
			wantErr: fmt.Errorf("validation error:\n - id: value is empty, which is not a valid UUID [string.uuid_empty]"),
		},
		{
			name: "invalid id",
			args: args{
				ctx: context.Background(),
				cur: &gen.UserID{
					Id: "something wrong",
				},
			},
			want:    nil,
			wantErr: fmt.Errorf("validation error:\n - id: value must be a valid UUID [string.uuid]"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedMocks != nil {
				tt.expectedMocks(tt.args.ctx)
			}
			got, err := server.GetUser(tt.args.ctx, tt.args.cur)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserServerImpl.GetUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserServerImpl_ListUsers(t *testing.T) {
	timeFreeze := time.Now()
	expectedCursor := "MjAyNC0wOC0xOFQxMTozNjo1MS45OTcyNjRafGFiZTA3NWQ2LWViMDQtNGQyNS04ZDlmLTk5YTdkN2M2YWY0ZA=="
	mockServiceQueries := appmocks.NewUserServiceQueries(t)
	mockLogger := loggermocks.NewInterface(t)
	protoValidator, err := protovalidate.New()
	assert.NoError(t, err, "creating protovalidator instance")

	server := &UserHandler{
		l:              mockLogger,
		serviceQueries: mockServiceQueries,
		protoValidator: protoValidator,
	}

	type args struct {
		ctx context.Context
		cur *gen.ListUsersRequest
	}
	tests := []struct {
		name          string
		args          args
		expectedMocks func(ctx context.Context)
		want          *gen.ListUsersResponse
		wantErr       error
	}{
		{
			name: "success - no cursor",
			args: args{
				ctx: context.Background(),
				cur: &gen.ListUsersRequest{
					Limit: 2,
				},
			},
			expectedMocks: func(ctx context.Context) {
				mockServiceQueries.On("ListUsers", ctx, user.ListUsersRequest{
					Cursor: "", Limit: 2,
				}).Return([]*domain.User{
					{
						ID:             uuid.MustParse("0f913f6a-497b-4305-b3d1-3f53657e3a25"),
						FirstName:      "first",
						LastName:       "last",
						NickName:       "nick",
						CountryISOCode: "UK",
						Email:          "something",
						Password:       "",
						CreatedAt:      timeFreeze,
						UpdatedAt:      timeFreeze,
					}, {
						ID:             uuid.MustParse("0f913f6a-497b-4305-b3d1-3f53657e3a27"),
						FirstName:      "first2",
						LastName:       "last2",
						NickName:       "nick2",
						CountryISOCode: "PT",
						Email:          "something2",
						Password:       "",
						CreatedAt:      timeFreeze,
						UpdatedAt:      timeFreeze,
					},
				}, expectedCursor, nil).Once()
			},
			want: &gen.ListUsersResponse{
				NextCursor: expectedCursor,
				Users: []*gen.ReadableUserFields{
					{
						Id:             "0f913f6a-497b-4305-b3d1-3f53657e3a25",
						FirstName:      "first",
						LastName:       "last",
						NickName:       "nick",
						CountryIsoCode: "UK",
						Email:          "something",
						CreatedAt:      timestamppb.New(timeFreeze),
						UpdatedAt:      timestamppb.New(timeFreeze),
					},
					{
						Id:             "0f913f6a-497b-4305-b3d1-3f53657e3a27",
						FirstName:      "first2",
						LastName:       "last2",
						NickName:       "nick2",
						CountryIsoCode: "PT",
						Email:          "something2",
						CreatedAt:      timestamppb.New(timeFreeze),
						UpdatedAt:      timestamppb.New(timeFreeze),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "success - with cursor",
			args: args{
				ctx: context.Background(),
				cur: &gen.ListUsersRequest{
					Limit:  2,
					Cursor: &expectedCursor,
				},
			},
			expectedMocks: func(ctx context.Context) {
				mockServiceQueries.On("ListUsers", ctx, user.ListUsersRequest{
					Cursor: expectedCursor, Limit: 2,
				}).Return([]*domain.User{
					{
						ID:             uuid.MustParse("0f913f6a-497b-4305-b3d1-3f53657e3a25"),
						FirstName:      "first",
						LastName:       "last",
						NickName:       "nick",
						CountryISOCode: "UK",
						Email:          "something",
						Password:       "",
						CreatedAt:      timeFreeze,
						UpdatedAt:      timeFreeze,
					},
				}, "", nil).Once()
			},
			want: &gen.ListUsersResponse{
				NextCursor: "",
				Users: []*gen.ReadableUserFields{
					{
						Id:             "0f913f6a-497b-4305-b3d1-3f53657e3a25",
						FirstName:      "first",
						LastName:       "last",
						NickName:       "nick",
						CountryIsoCode: "UK",
						Email:          "something",
						CreatedAt:      timestamppb.New(timeFreeze),
						UpdatedAt:      timestamppb.New(timeFreeze),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "error",
			args: args{
				ctx: context.Background(),
				cur: &gen.ListUsersRequest{
					Limit:  2,
					Cursor: &expectedCursor,
				},
			},
			expectedMocks: func(ctx context.Context) {
				mockServiceQueries.On("ListUsers", ctx, user.ListUsersRequest{
					Cursor: expectedCursor, Limit: 2,
				}).Return([]*domain.User{}, "", fmt.Errorf("something went wrong")).Once()
			},
			want:    &gen.ListUsersResponse{},
			wantErr: fmt.Errorf("something went wrong"),
		},
		{
			name: "nil payload",
			args: args{
				ctx: context.Background(),
				cur: nil,
			},
			want:    nil,
			wantErr: domain.ErrEmptyRequest,
		},
		{
			name: "invalid limit param",
			args: args{
				ctx: context.Background(),
				cur: &gen.ListUsersRequest{
					Limit: -1,
				},
			},
			want:    nil,
			wantErr: fmt.Errorf("validation error:\n - limit: value must be greater than or equal to 1 [int32.gte]"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedMocks != nil {
				tt.expectedMocks(tt.args.ctx)
			}
			got, err := server.ListUsers(tt.args.ctx, tt.args.cur)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error())
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UserServerImpl.ListUsers() = %v, want %v", got, tt.want)
			}
		})
	}
}
