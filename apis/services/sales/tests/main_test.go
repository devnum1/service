package tests

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"

	authbuild "github.com/ardanlabs/service/apis/services/auth/http/build/all"
	authmux "github.com/ardanlabs/service/apis/services/auth/http/mux"
	salesbuild "github.com/ardanlabs/service/apis/services/sales/build/all"
	salesmux "github.com/ardanlabs/service/apis/services/sales/mux"
	"github.com/ardanlabs/service/app/api/apptest"
	"github.com/ardanlabs/service/app/api/authsrv"
	"github.com/ardanlabs/service/business/data/dbtest"
	"github.com/ardanlabs/service/foundation/docker"
)

var c *docker.Container

func TestMain(m *testing.M) {
	code, err := run(m)
	if err != nil {
		fmt.Println(err)
	}

	os.Exit(code)
}

func run(m *testing.M) (int, error) {
	var err error

	c, err = dbtest.StartDB()
	if err != nil {
		return 1, err
	}
	defer dbtest.StopDB(c)

	return m.Run(), nil
}

func startTest(t *testing.T, testName string) (*dbtest.Test, *apptest.AppTest) {
	dbTest := dbtest.NewTest(t, c, testName)

	// -------------------------------------------------------------------------

	authMux := authmux.WebAPI(authmux.Config{
		Shutdown: make(chan os.Signal, 1),
		Log:      dbTest.Log,
		Auth:     dbTest.Auth,
		DB:       dbTest.DB,
		BusCrud: authmux.BusCrud{
			Delegate: dbTest.Core.BusCrud.Delegate,
			User:     dbTest.Core.BusCrud.User,
		},
	}, authbuild.Routes())

	logFunc := func(ctx context.Context, msg string) {
		t.Logf("authapi: message: %s", msg)
	}

	server := httptest.NewServer(authMux)
	authSrv := authsrv.New(server.URL, logFunc)

	// -------------------------------------------------------------------------

	appTest := apptest.New(salesmux.WebAPI(salesmux.Config{
		Shutdown: make(chan os.Signal, 1),
		Log:      dbTest.Log,
		AuthSrv:  authSrv,
		DB:       dbTest.DB,
		BusCrud: salesmux.BusCrud{
			Delegate: dbTest.Core.BusCrud.Delegate,
			Home:     dbTest.Core.BusCrud.Home,
			Product:  dbTest.Core.BusCrud.Product,
			User:     dbTest.Core.BusCrud.User,
		},
		BusView: salesmux.BusView{
			Product: dbTest.Core.BusView.Product,
		},
	}, salesbuild.Routes()))

	return dbTest, appTest
}
