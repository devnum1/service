// Package productgrp maintains the group of handlers for product access.
package productgrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/web/auth"
	v1Web "github.com/ardanlabs/service/business/web/v1"
	"github.com/ardanlabs/service/foundation/web"
)

// Handlers manages the set of product endpoints.
type Handlers struct {
	Product *product.Core
	Auth    *auth.Auth
}

// Create adds a new product to the system.
func (h Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var np product.NewProduct
	if err := web.Decode(r, &np); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	prod, err := h.Product.Create(ctx, np, web.GetTime(ctx))
	if err != nil {
		return fmt.Errorf("creating new product, np[%+v]: %w", np, err)
	}

	return web.Respond(ctx, w, prod, http.StatusCreated)
}

// Update updates a product in the system.
func (h Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	claims := auth.GetClaims(ctx)

	var upd product.UpdateProduct
	if err := web.Decode(r, &upd); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	id := web.Param(r, "id")

	prd, err := h.Product.QueryByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, product.ErrInvalidID):
			return v1Web.NewRequestError(err, http.StatusBadRequest)
		case errors.Is(err, product.ErrNotFound):
			return v1Web.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("querying product[%s]: %w", id, err)
		}
	}

	if claims.Subject != prd.UserID && h.Auth.Authorize(ctx, claims, auth.RuleAdminOnly) != nil {
		return auth.NewAuthError("auth failed")
	}

	if err := h.Product.Update(ctx, id, upd, web.GetTime(ctx)); err != nil {
		switch {
		case errors.Is(err, product.ErrInvalidID):
			return v1Web.NewRequestError(err, http.StatusBadRequest)
		case errors.Is(err, product.ErrNotFound):
			return v1Web.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("ID[%s] Product[%+v]: %w", id, &upd, err)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// Delete removes a product from the system.
func (h Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	claims := auth.GetClaims(ctx)
	id := web.Param(r, "id")

	prd, err := h.Product.QueryByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, product.ErrInvalidID):
			return v1Web.NewRequestError(err, http.StatusBadRequest)
		case errors.Is(err, product.ErrNotFound):

			// Don't send StatusNotFound here since the call to Delete
			// below won't if this product is not found. We only know
			// this because we are doing the Query for the UserID.
			return v1Web.NewRequestError(err, http.StatusNoContent)
		default:
			return fmt.Errorf("querying product[%s]: %w", id, err)
		}
	}

	if claims.Subject != prd.UserID && h.Auth.Authorize(ctx, claims, auth.RuleAdminOnly) != nil {
		return auth.NewAuthError("auth failed")
	}

	if err := h.Product.Delete(ctx, id); err != nil {
		switch {
		case errors.Is(err, product.ErrInvalidID):
			return v1Web.NewRequestError(err, http.StatusBadRequest)
		default:
			return fmt.Errorf("ID[%s]: %w", id, err)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// Query returns a list of products with paging.
func (h Handlers) Query(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	page := web.Param(r, "page")
	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		return v1Web.NewRequestError(fmt.Errorf("invalid page format, page[%s]", page), http.StatusBadRequest)
	}
	rows := web.Param(r, "rows")
	rowsPerPage, err := strconv.Atoi(rows)
	if err != nil {
		return v1Web.NewRequestError(fmt.Errorf("invalid rows format, rows[%s]", rows), http.StatusBadRequest)
	}

	orderBy, err := product.Order.FromQueryString(r.URL.Query().Get("orderby"))
	if err != nil {
		return v1Web.NewRequestError(err, http.StatusBadRequest)
	}

	products, err := h.Product.Query(ctx, orderBy, pageNumber, rowsPerPage)
	if err != nil {
		return fmt.Errorf("unable to query for products: %w", err)
	}

	return web.Respond(ctx, w, products, http.StatusOK)
}

// QueryByID returns a product by its ID.
func (h Handlers) QueryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id := web.Param(r, "id")
	prod, err := h.Product.QueryByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, product.ErrInvalidID):
			return v1Web.NewRequestError(err, http.StatusBadRequest)
		case errors.Is(err, product.ErrNotFound):
			return v1Web.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("ID[%s]: %w", id, err)
		}
	}

	return web.Respond(ctx, w, prod, http.StatusOK)
}
