package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ismelen/inkomi/back/auth-service/internal/config"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/api/handlers"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/api/routes"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/datasources"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/dropbox"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/google"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/repositories"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/security"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/dropboxlogin"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/getdropboxaccess"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/getfcmtoken"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/getprofile"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/googlelogin"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/linkdropbox"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/login"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/logout"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/refreshsession"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/register"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/socialsignin"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/startsession"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/unlinkdropbox"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/updatedropboxfolder"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/updatefcmtoken"
	"github.com/ismelen/inkomi/back/auth-service/migrations"
	"github.com/ismelen/inkomi/back/common/httpauth"
)

func main() {
	config.Load()
	ctx := context.Background()

	db, err := datasources.NewSQLiteDatasource(config.Env.DBPath, migrations.Users)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	cipher, err := security.NewAESTokenCipher(config.Env.TokenEncryptionKey)
	if err != nil {
		log.Fatal(err)
	}
	issuer, err := security.NewJWTTokenIssuer(config.Env.ServiceName, config.Env.JwtPrivateKey, config.Env.AccessTokenTTL)
	if err != nil {
		log.Fatal(err)
	}
	hasher := security.NewArgon2PasswordHasher()

	externalClient := &http.Client{Timeout: config.Env.ExternalTimeout}
	googleVerifier, err := google.NewIdTokenGoogleVerifier(ctx, externalClient, config.Env.GoogleClientIds)
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
	internalHandler := handlers.NewInternalHandler(
		getdropboxaccess.NewGetDropboxAccessUC(dropboxLinks, dropboxOAuth),
		getfcmtoken.NewGetFCMTokenUC(users),
	)

	authMiddleware, err := httpauth.NewAuthMiddleware(config.Env.JwtPublicKey)
	if err != nil {
		log.Fatal(err)
	}
	var servicePublicKeys map[string]string
	if err := json.Unmarshal([]byte(config.Env.ServicePublicKeys), &servicePublicKeys); err != nil {
		log.Fatalf("parse SERVICE_PUBLIC_KEYS: %v", err)
	}
	serviceMiddleware, err := httpauth.NewServiceAuthMiddleware(config.Env.ServiceName, servicePublicKeys)
	if err != nil {
		log.Fatal(err)
	}

	api := chi.NewRouter()
	routes.SetupAuthRoutes(api, authHandler)
	routes.SetupUserRoutes(api, usersHandler, authMiddleware)

	// The internal routes listen on their own port so the ingress never exposes them.
	internal := chi.NewRouter()
	routes.SetupInternalRoutes(internal, internalHandler, serviceMiddleware)

	errs := make(chan error, 2)
	go func() { errs <- newServer(config.Env.InternalPort, internal).ListenAndServe() }()
	go func() { errs <- newServer(config.Env.Port, api).ListenAndServe() }()

	log.Printf("Starting at port %s (internal %s)", config.Env.Port, config.Env.InternalPort)
	log.Fatal(<-errs)
}

func newServer(port string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
	}
}
