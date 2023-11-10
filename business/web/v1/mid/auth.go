package mid

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ardanlabs/service/business/core/home"
	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/core/user"
	"github.com/ardanlabs/service/business/web/v1/auth"
	"github.com/ardanlabs/service/business/web/v1/response"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/google/uuid"
)

// Set of error variables for handling user group errors.
var (
	ErrInvalidID = errors.New("ID is not in its proper form")
)

// Authenticate validates a JWT from the `Authorization` header.
func Authenticate(a *auth.Auth) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			claims, err := a.Authenticate(ctx, r.Header.Get("authorization"))
			if err != nil {
				return auth.NewAuthError("authenticate: failed: %s", err)
			}

			if claims.Subject == "" {
				return auth.NewAuthError("authorize: you are not authorized for that action, no claims")
			}

			subjectID, err := uuid.Parse(claims.Subject)
			if err != nil {
				return response.NewError(ErrInvalidID, http.StatusBadRequest)
			}

			ctx = auth.SetUserID(ctx, subjectID)
			ctx = auth.SetClaims(ctx, claims)

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

// Authorize validates that an authenticated user has at least one role from a
// specified list. This method constructs the actual function that is used.
func Authorize(a *auth.Auth, rule string) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			claims := auth.GetClaims(ctx)
			if err := a.Authorize(ctx, claims, uuid.UUID{}, rule); err != nil {
				return auth.NewAuthError("authorize: you are not authorized for that action, claims[%v] rule[%v]: %s", claims.Roles, rule, err)
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

// AuthorizeUser validates that an authenticated user has at least one role from a
// specified list. This method constructs the actual function that is used.
func AuthorizeUser(a *auth.Auth, rule string, usrCore *user.Core) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			var userID uuid.UUID

			if id := web.Param(r, "user_id"); id != "" {
				var err error
				userID, err = uuid.Parse(id)
				if err != nil {
					return response.NewError(ErrInvalidID, http.StatusBadRequest)
				}

				usr, err := usrCore.QueryByID(ctx, userID)
				if err != nil {
					switch {
					case errors.Is(err, user.ErrNotFound):
						return response.NewError(err, http.StatusNoContent)
					default:
						return fmt.Errorf("querybyid: userID[%s]: %w", userID, err)
					}
				}

				ctx = SetUser(ctx, usr)
			}

			claims := auth.GetClaims(ctx)
			if err := a.Authorize(ctx, claims, userID, rule); err != nil {
				return auth.NewAuthError("authorize: you are not authorized for that action, claims[%v] rule[%v]: %s", claims.Roles, rule, err)
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

func AuthorizeProduct(a *auth.Auth, rule string, prdCore *product.Core) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			var userID uuid.UUID

			if id := web.Param(r, "product_id"); id != "" {
				var err error
				productID, err := uuid.Parse(id)
				if err != nil {
					return response.NewError(ErrInvalidID, http.StatusBadRequest)
				}

				prd, err := prdCore.QueryByID(ctx, productID)
				if err != nil {
					switch {
					case errors.Is(err, product.ErrNotFound):
						return response.NewError(err, http.StatusNoContent)
					default:
						return fmt.Errorf("querybyid: productID[%s]: %w", productID, err)
					}
				}

				userID = prd.UserID
				ctx = SetProduct(ctx, prd)
			}

			claims := auth.GetClaims(ctx)
			if err := a.Authorize(ctx, claims, userID, rule); err != nil {
				return auth.NewAuthError("authorize: you are not authorized for that action, claims[%v] rule[%v]: %s", claims.Roles, rule, err)
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

func AuthorizeHome(a *auth.Auth, rule string, hmeCore *home.Core) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			var userID uuid.UUID

			if id := web.Param(r, "home_id"); id != "" {
				var err error
				homeID, err := uuid.Parse(id)
				if err != nil {
					return response.NewError(ErrInvalidID, http.StatusBadRequest)
				}

				hme, err := hmeCore.QueryByID(ctx, homeID)
				if err != nil {
					switch {
					case errors.Is(err, home.ErrNotFound):
						return response.NewError(err, http.StatusNoContent)
					default:
						return fmt.Errorf("querybyid: homeID[%s]: %w", homeID, err)
					}
				}

				userID = hme.UserID
				ctx = SetHome(ctx, hme)
			}

			claims := auth.GetClaims(ctx)
			if err := a.Authorize(ctx, claims, userID, rule); err != nil {
				return auth.NewAuthError("authorize: you are not authorized for that action, claims[%v] rule[%v]: %s", claims.Roles, rule, err)
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
