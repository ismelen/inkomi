package main

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"

	"github.com/ismelen/inkomi/back/monolith/internal/config"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/infra/api/handlers"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/infra/api/routes"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/infra/datasources"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/infra/dropbox"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/infra/google"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/infra/repositories"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/infra/security"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/dropboxlogin"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/getdropboxaccess"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/getfcmtoken"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/getprofile"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/googlelogin"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/linkdropbox"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/login"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/logout"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/refreshsession"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/register"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/socialsignin"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/startsession"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/unlinkdropbox"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/updatedropboxfolder"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/updatefcmtoken"
	"github.com/ismelen/inkomi/back/monolith/migrations"
)

// authAPI are the auth use cases other modules call in-process.
type authAPI struct {
	getDropboxAccess *getdropboxaccess.GetDropboxAccessUC
	getFCMToken      *getfcmtoken.GetFCMTokenUC
}

func wireAuth(d *deps, r chi.Router) (func(), *authAPI) {
	db, err := datasources.NewSQLiteDatasource(filepath.Join(config.Env.DataDir, "users.db"), migrations.Users)
	if err != nil {
		log.Fatal(err)
	}

	cipher, err := security.NewAESTokenCipher(config.Env.TokenEncryptionKey)
	if err != nil {
		log.Fatal(err)
	}
	issuer, err := security.NewJWTTokenIssuer(config.Env.TokenIssuer, config.Env.JwtPrivateKey, config.Env.AccessTokenTTL)
	if err != nil {
		log.Fatal(err)
	}
	hasher := security.NewArgon2PasswordHasher()

	externalClient := &http.Client{Timeout: config.Env.ExternalTimeout}
	googleVerifier, err := google.NewIdTokenGoogleVerifier(d.ctx, externalClient, config.Env.GoogleClientIds)
	if err != nil {
		log.Fatal(err)
	}
	dropboxOAuth := dropbox.NewHTTPDropboxOAuth(externalClient, dropbox.APIURL, config.Env.DropboxAppKey)

	users := repositories.NewSQLiteUserRepository(db.DB)
	identities := repositories.NewSQLiteIdentityRepository(db.DB)
	sessions := repositories.NewSQLiteSessionRepository(db.DB)
	dropboxLinks := repositories.NewSQLiteDropboxLinkRepository(db.DB, cipher)

	startSessionUC := startsession.NewStartSessionUC(sessions, issuer, config.Env.RefreshTokenTTL)
	socialSignInUC := socialsignin.NewSocialSignInUC(db, users, identities)
	authHandler := handlers.NewAuthHandler(
		register.NewRegisterUC(users, hasher, startSessionUC),
		login.NewLoginUC(users, hasher, startSessionUC),
		googlelogin.NewGoogleLoginUC(googleVerifier, socialSignInUC, startSessionUC),
		dropboxlogin.NewDropboxLoginUC(dropboxOAuth, dropboxLinks, socialSignInUC, startSessionUC),
		refreshsession.NewRefreshSessionUC(sessions, users, issuer, config.Env.RefreshTokenTTL),
		logout.NewLogoutUC(sessions, issuer),
	)
	usersHandler := handlers.NewUsersHandler(
		getprofile.NewGetProfileUC(users, identities, dropboxLinks),
		updatefcmtoken.NewUpdateFCMTokenUC(users),
		linkdropbox.NewLinkDropboxUC(dropboxOAuth, identities, dropboxLinks),
		updatedropboxfolder.NewUpdateDropboxFolderUC(dropboxLinks),
		unlinkdropbox.NewUnlinkDropboxUC(dropboxLinks),
	)

	routes.SetupAuthRoutes(r, authHandler)
	routes.SetupUserRoutes(r, usersHandler, d.authMiddleware)

	api := &authAPI{
		getDropboxAccess: getdropboxaccess.NewGetDropboxAccessUC(dropboxLinks, dropboxOAuth),
		getFCMToken:      getfcmtoken.NewGetFCMTokenUC(users),
	}
	return closeLogged("users.db", db.Close), api
}
