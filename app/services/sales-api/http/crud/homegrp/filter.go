package homegrp

import (
	"net/http"

	"github.com/ardanlabs/service/app/services/sales-api/apis/crud/homeapi"
	"github.com/ardanlabs/service/business/api/page"
)

func parseQueryParams(r *http.Request) (homeapi.QueryParams, error) {
	const (
		orderBy                  = "orderBy"
		filterPage               = "page"
		filterRow                = "row"
		filterByHomeID           = "home_id"
		filterByUserID           = "user_id"
		filterByType             = "type"
		filterByStartCreatedDate = "start_date_created"
		filterByEndCreatedDate   = "end_date_created"
	)

	values := r.URL.Query()

	var filter homeapi.QueryParams

	pg, err := page.Parse(r)
	if err != nil {
		return homeapi.QueryParams{}, err
	}
	filter.Page = pg.Number
	filter.Rows = pg.RowsPerPage

	if orderBy := values.Get(orderBy); orderBy != "" {
		filter.OrderBy = orderBy
	}

	if homeID := values.Get(filterByHomeID); homeID != "" {
		filter.ID = homeID
	}

	if userID := values.Get(filterByUserID); userID != "" {
		filter.UserID = userID
	}

	if typ := values.Get(filterByType); typ != "" {
		filter.Type = typ
	}

	if startDate := values.Get(filterByStartCreatedDate); startDate != "" {
		filter.StartCreatedDate = startDate
	}

	if endDate := values.Get(filterByStartCreatedDate); endDate != "" {
		filter.EndCreatedDate = endDate
	}

	return filter, nil
}
