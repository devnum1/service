package productdb

import (
	"fmt"
	"time"

	"github.com/ardanlabs/service/business/core/product"
	"github.com/ardanlabs/service/business/data/order"
)

// dbProduct represents an individual product.
type dbProduct struct {
	ID          string    `db:"product_id"`   // Unique identifier.
	Name        string    `db:"name"`         // Display name of the product.
	Cost        int       `db:"cost"`         // Price for one item in cents.
	Quantity    int       `db:"quantity"`     // Original number of items available.
	Sold        int       `db:"sold"`         // Aggregate field showing number of items sold.
	Revenue     int       `db:"revenue"`      // Aggregate field showing total cost of sold items.
	UserID      string    `db:"user_id"`      // ID of the user who created the product.
	DateCreated time.Time `db:"date_created"` // When the product was added.
	DateUpdated time.Time `db:"date_updated"` // When the product record was last modified.
}

// =============================================================================

func toDBProduct(prd product.Product) dbProduct {
	return dbProduct(prd)
}

func toCoreProduct(dbPrd dbProduct) product.Product {
	prd := product.Product(dbPrd)

	prd.DateCreated = time.Date(prd.DateCreated.Year(), prd.DateCreated.Month(), prd.DateCreated.Day(), prd.DateCreated.Hour(), prd.DateCreated.Minute(), prd.DateCreated.Second(), prd.DateCreated.Nanosecond(), time.UTC)
	prd.DateUpdated = time.Date(prd.DateUpdated.Year(), prd.DateUpdated.Month(), prd.DateUpdated.Day(), prd.DateUpdated.Hour(), prd.DateUpdated.Minute(), prd.DateUpdated.Second(), prd.DateUpdated.Nanosecond(), time.UTC)

	return prd
}

func toCoreProductSlice(dbProducts []dbProduct) []product.Product {
	prds := make([]product.Product, len(dbProducts))
	for i, dbPrd := range dbProducts {
		prds[i] = toCoreProduct(dbPrd)
	}
	return prds
}

// =============================================================================

// orderByfields is the map of fields that is used to translate between the
// application layer names and the database.
var orderByFields = map[string]string{
	product.OrderByID:       "product_id",
	product.OrderByName:     "name",
	product.OrderByCost:     "cost",
	product.OrderByQuantity: "quantity",
	product.OrderBySold:     "sold",
	product.OrderByRevenue:  "revenue",
	product.OrderByUserID:   "user_id",
}

// orderByClause validates the order by for correct fields and sql injection.
func orderByClause(orderBy order.By) (string, error) {
	if err := order.Validate(orderBy.Field, orderBy.Direction); err != nil {
		return "", err
	}

	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return by + " " + orderBy.Direction, nil
}
