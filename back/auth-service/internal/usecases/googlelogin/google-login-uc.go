package googlelogin

import (
	"context"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/domain/ports"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/socialsignin"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/startsession"
)

type GoogleLoginUC struct {
	verifier     ports.IGoogleVerifier
	signIn       *socialsignin.SocialSignInUC
	startSession *startsession.StartSessionUC
}

func NewGoogleLoginUC(verifier ports.IGoogleVerifier, signIn *socialsignin.SocialSignInUC, startSession *startsession.StartSessionUC) *GoogleLoginUC {
	return &GoogleLoginUC{verifier: verifier, signIn: signIn, startSession: startSession}
}

func (g *GoogleLoginUC) Execute(ctx context.Context, idToken, deviceId string) (*models.TokenPair, error) {
	if err := models.ValidateDeviceId(deviceId); err != nil {
		return nil, err
	}

	account, err := g.verifier.Verify(ctx, idToken)
	if err != nil {
		return nil, err
	}

	user, err := g.signIn.Execute(ctx, *account)
	if err != nil {
		return nil, err
	}
	return g.startSession.Execute(ctx, user, deviceId)
}
