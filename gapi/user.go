package gapi

import (
	"context"
	"errors"
	"time"

	db "github.com/andreanpradanaa/simple-bank-app/db/sqlc"
	proto "github.com/andreanpradanaa/simple-bank-app/pb"
	"github.com/andreanpradanaa/simple-bank-app/utils"
	"github.com/andreanpradanaa/simple-bank-app/val"
	"github.com/andreanpradanaa/simple-bank-app/worker"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) CreateUser(ctx context.Context, request *proto.CreateUserRequest) (*proto.CreateUserResponse, error) {
	violations := validateCreateUserRequest(request)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	hashedPassword, err := utils.HashPassword(request.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot hash password: %s", err.Error())
	}

	arg := db.CreateUserTxParams{
		CreateUserParams: db.CreateUserParams{
			Username:       request.GetUsername(),
			HashedPassword: hashedPassword,
			FullName:       request.GetFullName(),
			Email:          request.GetEmail(),
		},

		AfterCreateUser: func(user db.User) error {
			taskPayload := &worker.PayloadSendVerifyEmail{
				Username: user.Username,
			}
			opts := []asynq.Option{
				asynq.MaxRetry(10),
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(worker.CriticalQueue),
			}
			return server.taskDistributor.DistributeTaskSendVerifyEmail(ctx, taskPayload, opts...)
		},
	}

	txResult, err := server.store.CreateUserTx(ctx, arg)
	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolation {
			return nil, status.Errorf(codes.AlreadyExists, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "cannot create user: %s", err.Error())
	}

	response := proto.CreateUserResponse{
		User: &proto.User{
			Username:          txResult.User.Username,
			FullName:          txResult.User.FullName,
			Email:             txResult.User.Email,
			PasswordChangedAt: timestamppb.New(txResult.User.PasswordChangedAt),
			CreatedAt:         timestamppb.New(txResult.User.CreatedAt),
		},
	}

	return &response, nil
}

func (server *Server) LoginUser(ctx context.Context, req *proto.LoginUserRequest) (*proto.LoginUserResponse, error) {
	violations := validateLoginUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	user, err := server.store.GetUser(ctx, req.GetUsername())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "user not found: %s", err.Error())
		}
		return nil, status.Errorf(codes.Internal, "cannot get user: %s", err.Error())
	}

	err = utils.CheckPassword(req.GetPassword(), user.HashedPassword)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "password salah: %s", err.Error())
	}

	token, payload, err := server.tokenMaker.CreateToken(req.GetUsername(), user.Role, server.config.AccessTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create token: %s", err.Error())
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(payload.Username, user.Role, server.config.RefreshTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot create refresh token: %s", err.Error())
	}

	mdtd := server.extractMetadata(ctx)
	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     refreshPayload.Username,
		RefreshToken: refreshToken,
		UserAgent:    mdtd.UserAgent,
		ClientIp:     mdtd.ClientIP,
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	res := &proto.LoginUserResponse{
		SessionId:             session.ID.String(),
		AccessToken:           token,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  timestamppb.New(payload.ExpiredAt),
		RefreshTokenExpiresAt: timestamppb.New(refreshPayload.ExpiredAt),
		User: &proto.User{
			Username:          user.Username,
			FullName:          user.FullName,
			Email:             user.Email,
			PasswordChangedAt: &timestamppb.Timestamp{},
			CreatedAt:         &timestamppb.Timestamp{},
		},
	}

	return res, nil
}

func (server *Server) UpdateUser(ctx context.Context, req *proto.UpdateUserRequest) (*proto.UpdateUserResponse, error) {
	authPayload, err := server.AuthorizeUser(ctx, []string{utils.BankerRole, utils.DepositorRole})
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())

	}

	violations := validateUpdateUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	if authPayload.Role != utils.BankerRole && authPayload.Username != req.GetUsername() {
		return nil, status.Errorf(codes.PermissionDenied, "cannot update other user's info")
	}

	args := db.UpdateUserParams{
		FullName: pgtype.Text{
			String: req.GetFullName(),
			Valid:  req.FullName != nil,
		},
		Email: pgtype.Text{
			String: req.GetEmail(),
			Valid:  req.Email != nil,
		},
		Username: req.Username,
	}

	if req.HashedPassword != nil {
		hahedPassword, err := utils.HashPassword(req.GetHashedPassword())
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		args.HashedPassword = pgtype.Text{
			String: hahedPassword,
			Valid:  true,
		}

		args.PasswordChangedAt = pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		}
	}

	updatedUser, err := server.store.UpdateUser(ctx, args)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &proto.UpdateUserResponse{
		User: &proto.User{
			Username:          updatedUser.Username,
			FullName:          updatedUser.FullName,
			Email:             updatedUser.Email,
			PasswordChangedAt: timestamppb.New(updatedUser.PasswordChangedAt),
			CreatedAt:         timestamppb.New(updatedUser.CreatedAt),
		},
	}

	return res, nil
}

func (server *Server) VerifyEmail(ctx context.Context, request *proto.VerifyEmailRequest) (*proto.VerifyEmailResponse, error) {
	violations := validateVerifyEmailRequest(request)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	txResult, err := server.store.VerifyEmailTx(ctx, db.VerifyEmailTxParams{
		EmailId:    request.EmailId,
		SecretCode: request.SecretCode,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to verify email")
	}

	response := proto.VerifyEmailResponse{
		IsVerified: txResult.User.IsEmailVerified,
	}

	return &response, nil
}

func validateCreateUserRequest(req *proto.CreateUserRequest) (violation []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violation = append(violation, fieldViolation("username", err))
	}

	if err := val.ValidatePassword(req.GetPassword()); err != nil {
		violation = append(violation, fieldViolation("password", err))
	}

	if err := val.ValidateFullname(req.GetFullName()); err != nil {
		violation = append(violation, fieldViolation("full_name", err))
	}

	if err := val.ValidateEmail(req.GetEmail()); err != nil {
		violation = append(violation, fieldViolation("email", err))
	}

	return violation
}

func validateUpdateUserRequest(req *proto.UpdateUserRequest) (violation []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violation = append(violation, fieldViolation("username", err))
	}

	if req.HashedPassword != nil {
		if err := val.ValidatePassword(req.GetHashedPassword()); err != nil {
			violation = append(violation, fieldViolation("password", err))
		}
	}

	if req.FullName != nil {
		if err := val.ValidateFullname(req.GetFullName()); err != nil {
			violation = append(violation, fieldViolation("full_name", err))
		}
	}

	if req.Email != nil {
		if err := val.ValidateEmail(req.GetEmail()); err != nil {
			violation = append(violation, fieldViolation("email", err))
		}
	}

	return violation
}

func validateLoginUserRequest(req *proto.LoginUserRequest) (violation []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violation = append(violation, fieldViolation("username", err))
	}

	if err := val.ValidatePassword(req.GetPassword()); err != nil {
		violation = append(violation, fieldViolation("password", err))
	}

	return violation
}

func validateVerifyEmailRequest(req *proto.VerifyEmailRequest) (violation []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateEmailId(req.GetEmailId()); err != nil {
		violation = append(violation, fieldViolation("email_id", err))
	}

	if err := val.ValidateSecretCode(req.GetSecretCode()); err != nil {
		violation = append(violation, fieldViolation("secret_code", err))
	}

	return violation
}
