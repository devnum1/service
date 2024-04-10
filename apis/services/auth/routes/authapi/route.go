package authapi

import (
	"net/http"

	midhttp "github.com/ardanlabs/service/app/api/mid/http"
	"github.com/ardanlabs/service/app/core/crud/userapp"
	"github.com/ardanlabs/service/business/api/auth"
	"github.com/ardanlabs/service/business/core/crud/userbus"
	"github.com/ardanlabs/service/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	UserBus *userbus.Core
	Auth    *auth.Auth
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	authen := midhttp.Authenticate(cfg.UserBus, cfg.Auth)

	api := newAPI(userapp.NewCore(cfg.UserBus, cfg.Auth), cfg.Auth)
	app.Handle(http.MethodGet, version, "/users/token/{kid}", api.token, authen)
	app.Handle(http.MethodPost, version, "/users/authorize", api.authorize)
}
