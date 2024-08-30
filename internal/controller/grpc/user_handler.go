package grpc

import (
	"context"
	gen "users/gen/proto/go"
	"users/internal/app"
	"users/internal/app/user"
	"users/internal/domain"

	"users/pkg/logger"

	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserHandler struct {
	gen.UnimplementedUserServiceServer
	l               logger.Interface
	serviceCommands app.UserServiceCommands
	serviceQueries  app.UserServiceQueries
	protoValidator  *protovalidate.Validator
}

func (us UserHandler) CreateUser(ctx context.Context, cur *gen.CreateUserRequest) (*gen.UserID, error) {
	if err := us.protoValidator.Validate(cur); err != nil {
		return nil, err
	}
	userID, err := us.serviceCommands.CreateUser(ctx, user.AddUserRequest{
		FirstName:      cur.GetFirstName(),
		LastName:       cur.GetLastName(),
		NickName:       cur.GetNickName(),
		CountryISOCode: cur.GetCountryIsoCode(),
		Email:          cur.GetEmail(),
		Password:       cur.GetPassword(),
	})
	return &gen.UserID{Id: userID}, err
}

func (us UserHandler) DeleteUser(ctx context.Context, dur *gen.UserID) (*gen.UserID, error) {
	if err := us.protoValidator.Validate(dur); err != nil {
		return nil, err
	}
	err := us.serviceCommands.DeleteUser(ctx, dur.GetId())
	return &gen.UserID{Id: dur.GetId()}, err
}

func (us UserHandler) UpdateUser(ctx context.Context, uur *gen.UpdateUserRequest) (*gen.UserID, error) {
	if err := us.protoValidator.Validate(uur); err != nil {
		return nil, err
	}
	pbUser := uur.GetUser()
	err := us.serviceCommands.UpdateUser(ctx, user.UpdateUserRequest{
		ID:             uur.GetId(),
		FirstName:      pbUser.GetFirstName(),
		LastName:       pbUser.GetLastName(),
		NickName:       pbUser.GetNickName(),
		CountryISOCode: pbUser.GetCountryIsoCode(),
		Email:          pbUser.GetEmail(),
	})
	return &gen.UserID{Id: uur.GetId()}, err
}

func (us UserHandler) GetUser(ctx context.Context, gur *gen.UserID) (*gen.UserResponse, error) {
	if err := us.protoValidator.Validate(gur); err != nil {
		return nil, err
	}
	user, err := us.serviceQueries.GetUser(ctx, gur.GetId())
	if err != nil || user == nil {
		return &gen.UserResponse{}, err
	}
	return &gen.UserResponse{User: &gen.ReadableUserFields{
		Id:             user.ID.String(),
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		NickName:       user.NickName,
		Email:          user.Email,
		CountryIsoCode: user.CountryISOCode,
		CreatedAt:      timestamppb.New(user.CreatedAt),
		UpdatedAt:      timestamppb.New(user.UpdatedAt)}}, err
}

func (us UserHandler) ListUsers(ctx context.Context, lur *gen.ListUsersRequest) (*gen.ListUsersResponse, error) {
	if lur == nil {
		return nil, domain.ErrEmptyRequest
	}
	if err := us.protoValidator.Validate(lur); err != nil {
		return nil, err
	}
	var resp = &gen.ListUsersResponse{}
	userList, nextCursor, err := us.serviceQueries.ListUsers(ctx,
		user.ListUsersRequest{
			Cursor: lur.GetCursor(),
			Limit:  lur.GetLimit(),
			Filters: domain.UserSearchFilters{
				FirstName:      lur.FirstName,
				LastName:       lur.LastName,
				NickName:       lur.NickName,
				CountryISOCode: lur.CountryIsoCode,
				Email:          lur.Email,
			},
		})
	if err != nil {
		return resp, err
	}
	resp.NextCursor = nextCursor
	resp.Users = make([]*gen.ReadableUserFields, 0, len(userList))
	for _, user := range userList {
		resp.Users = append(resp.Users, &gen.ReadableUserFields{
			Id:             user.ID.String(),
			FirstName:      user.FirstName,
			LastName:       user.LastName,
			NickName:       user.NickName,
			Email:          user.Email,
			CountryIsoCode: user.CountryISOCode,
			CreatedAt:      timestamppb.New(user.CreatedAt),
			UpdatedAt:      timestamppb.New(user.UpdatedAt),
		})
	}
	return resp, nil
}
