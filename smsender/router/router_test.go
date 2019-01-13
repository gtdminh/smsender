package router

import (
	"fmt"
	"testing"

	"github.com/minchao/smsender/smsender/model"
	"github.com/minchao/smsender/smsender/providers/dummy"
	"github.com/minchao/smsender/smsender/store"
	dummystore "github.com/minchao/smsender/smsender/store/dummy"
)

type testRouteStore struct {
	*dummystore.RouteStore
}

func (rs *testRouteStore) SaveAll(routes []*model.Route) store.Channel {
	storeChannel := make(store.Channel, 1)

	go func() {
		storeChannel <- store.Result{}
		close(storeChannel)
	}()

	return storeChannel
}

func createRouter() *Router {
	dummyProvider1 := dummy.New("dummy1")
	dummyProvider2 := dummy.New("dummy2")
	router := Router{store: &dummystore.Store{DummyRoute: &testRouteStore{}}}

	_ = router.Add(model.NewRoute("default", `^\+.*`, dummyProvider1, true))
	_ = router.Add(model.NewRoute("japan", `^\+81`, dummyProvider2, true))
	_ = router.Add(model.NewRoute("taiwan", `^\+886`, dummyProvider2, true))
	_ = router.Add(model.NewRoute("telco", `^\+886987`, dummyProvider2, true))
	_ = router.Add(model.NewRoute("user", `^\+886987654321`, dummyProvider2, true))

	return &router
}

func compareOrder(routes []*model.Route, expected []string) error {
	got := []string{}
	isNotMatch := false
	for i, route := range routes {
		got = append(got, route.Name)
		if route.Name != expected[i] {
			isNotMatch = true
		}
	}
	if isNotMatch {
		return fmt.Errorf("routes expecting %v, but got %v", expected, got)
	}
	return nil
}

func TestRouter_GetAll(t *testing.T) {
	router := createRouter()

	if err := compareOrder(router.GetAll(), []string{"user", "telco", "taiwan", "japan", "default"}); err != nil {
		t.Fatal(err)
	}
}

func TestRouter_Get(t *testing.T) {
	router := createRouter()

	route := router.Get("japan")
	if route == nil || route.Name != "japan" {
		t.Fatal("got wrong route")
	}
	route = router.Get("usa")
	if route != nil {
		t.Fatal("route should be nil")
	}
}

func TestRouter_Set(t *testing.T) {
	router := createRouter()
	provider := dummy.New("dummy")

	route := model.NewRoute("user", `^\+886999999999`, provider, true).SetFrom("sender")

	if err := router.Set(route.Name, route.Pattern, route.GetProvider(), route.From, true); err == nil {
		newRoute := router.Get("user")
		if newRoute == nil {
			t.Fatal("route is not equal")
		}
		if newRoute.Name != route.Name {
			t.Fatal("route.Name is not equal")
		}
		if newRoute.Pattern != route.Pattern {
			t.Fatal("route.Pattern is not equal")
		}
		if newRoute.GetProvider() == nil || newRoute.GetProvider().Name() != route.GetProvider().Name() {
			t.Fatal("route.Provider is not equal")
		}
		if newRoute.From != route.From {
			t.Fatal("route.From is not equal")
		}
	}
	if ok := router.Get("user").Match("+886987654321"); ok {
		t.Fatal("route should not matched")
	}
	if ok := router.Get("user").Match("+886999999999"); !ok {
		t.Fatal("route should be matched")
	}

	if err := router.Set("france", "", provider, "", true); err == nil {
		t.Fatal("set route should be failed")
	}
}

func TestRouter_Remove(t *testing.T) {
	router := createRouter()

	_ = router.Remove("telco")
	_ = router.Remove("japan")
	if len(router.routes) != 3 {
		t.Fatal("remove route failed")
	}
	if err := compareOrder(router.routes, []string{"user", "taiwan", "default"}); err != nil {
		t.Fatal(err)
	}
}

func TestRouter_Reorder(t *testing.T) {
	newRouter := func() *Router {
		router := Router{store: &dummystore.Store{DummyRoute: &testRouteStore{}}}
		provider := dummy.New("dummy")
		for _, r := range []string{"D", "C", "B", "A"} {
			_ = router.Add(model.NewRoute(r, "", provider, true))
		}
		return &router
	}

	router := newRouter()

	if err := router.Reorder(-1, 0, 0); err == nil {
		t.Fatal("got incorrect error: nil")
	}
	if err := router.Reorder(4, 0, 0); err == nil {
		t.Fatal("got incorrect error: nil")
	}
	if err := router.Reorder(1, 0, 0); err == nil {
		t.Fatal("got incorrect error: nil")
	}
	if err := router.Reorder(0, 0, 0); err == nil {
		t.Fatal("got incorrect error: nil")
	}
	if err := router.Reorder(1, 4, 0); err == nil {
		t.Fatal("got incorrect error: nil")
	}
	if err := router.Reorder(0, 1, -1); err == nil {
		t.Fatal("got incorrect error: nil")
	}
	if err := router.Reorder(0, 1, 5); err == nil {
		t.Fatal("got incorrect error: nil")
	}

	checkReorderRoutes(t, newRouter(), 1, 2, 3, []string{"A", "B", "C", "D"})
	checkReorderRoutes(t, newRouter(), 2, 2, 1, []string{"A", "C", "D", "B"})
	checkReorderRoutes(t, newRouter(), 0, 2, 4, []string{"C", "D", "A", "B"})
}

func checkReorderRoutes(t *testing.T, router *Router, rangeStart, rangeLength, insertBefore int, expected []string) {
	if err := router.Reorder(rangeStart, rangeLength, insertBefore); err != nil {
		t.Fatalf("reorder routes error: %v", err)
	}
	if err := compareOrder(router.routes, expected); err != nil {
		t.Fatal(err)
	}
}

type routeTest struct {
	phone       string
	shouldMatch bool
	route       string
	provider    string
}

func TestRouter_Match(t *testing.T) {
	router := createRouter()

	tests := []routeTest{
		{
			phone:       "+886987654321",
			shouldMatch: true,
			route:       "user",
			provider:    "dummy2",
		},
		{
			phone:       "+886987654322",
			shouldMatch: true,
			route:       "telco",
			provider:    "dummy2",
		},
		{
			phone:       "+886900000001",
			shouldMatch: true,
			route:       "taiwan",
			provider:    "dummy2",
		},
		{
			phone:       "+819000000001",
			shouldMatch: true,
			route:       "japan",
			provider:    "dummy2",
		},
		{
			phone:       "+10000000001",
			shouldMatch: true,
			route:       "default",
			provider:    "dummy1",
		},
		{
			phone:       "woo",
			shouldMatch: false,
			route:       "",
			provider:    "",
		},
	}

	for i, test := range tests {
		match, ok := router.Match(test.phone)
		if test.shouldMatch {
			if !ok {
				t.Fatalf("test '%d' should match", i)
			}
			if test.route != match.Name {
				t.Fatalf("test '%d' route.Name is not equal", i)
			}
			if test.provider != match.GetProvider().Name() {
				t.Fatalf("test '%d' route.Provider is not equal", i)
			}
		} else {
			if ok {
				t.Fatalf("test '%d' should not match", i)
			}
		}
	}
}

func TestRouter_Match2(t *testing.T) {
	router := createRouter()
	router.Get("telco").IsActive = false

	if match, ok := router.Match("+886987"); ok {
		if match.Name != "taiwan" {
			t.Fatal("test route.Name should be 'taiwan'")
		}
	} else {
		t.Fatal("test should match")
	}
}
