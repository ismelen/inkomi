package socialsignin

import (
	"context"
	"errors"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/domain/ports"
)

type SocialSignInUC struct {
	tx         ports.ITxManager
	users      ports.IUserRepository
	identities ports.IIdentityRepository
}

func NewSocialSignInUC(tx ports.ITxManager, users ports.IUserRepository, identities ports.IIdentityRepository) *SocialSignInUC {
	return &SocialSignInUC{tx: tx, users: users, identities: identities}
}

// Execute returns the user of a verified external account. An unknown account is linked to
// the user with the same verified email or, if there is none, to a new user.
func (s *SocialSignInUC) Execute(ctx context.Context, account models.ExternalAccount) (*models.User, error) {
	var user *models.User
	err := s.tx.ExecuteTx(ctx, func(ctx context.Context) error {
		var err error
		user, err = s.signIn(ctx, account)
		return err
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *SocialSignInUC) signIn(ctx context.Context, account models.ExternalAccount) (*models.User, error) {
	identity, err := s.identities.Get(ctx, account.Provider, account.Subject)
	if err == nil {
		return s.users.GetById(ctx, identity.UserId)
	}
	if !errors.Is(err, models.ErrIdentityNotFound) {
		return nil, err
	}

	user, err := s.userWithVerifiedEmail(ctx, account)
	if err != nil {
		return nil, err
	}

	if user != nil {
		user.ClaimEmail()
		err = s.users.Update(ctx, user)
	} else {
		user = models.NewSocialUser(account)
		err = s.users.Create(ctx, user)
	}
	if err != nil {
		return nil, err
	}

	identity = &models.Identity{Provider: account.Provider, Subject: account.Subject, UserId: user.Id}
	if err := s.identities.Create(ctx, *identity); err != nil {
		return nil, err
	}
	return user, nil
}

// userWithVerifiedEmail returns nil when account has no verified email or no user has it.
func (s *SocialSignInUC) userWithVerifiedEmail(ctx context.Context, account models.ExternalAccount) (*models.User, error) {
	if !account.EmailVerified {
		return nil, nil
	}
	email, err := models.NormalizeEmail(account.Email)
	if err != nil {
		return nil, nil
	}

	user, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, models.ErrUserNotFound) {
		return nil, nil
	}
	return user, err
}
