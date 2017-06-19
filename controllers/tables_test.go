package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"time"

	"io/ioutil"

	"strings"

	"github.com/gorilla/mux"
)

func TestGetTables(t *testing.T) {
	var testCases = []struct {
		description string
		url         string
		method      string
		status      int
	}{
		{"Get tables without custom where clause", "/tables", "GET", http.StatusOK},
		{"Get tables with custom where clause", "/tables?c.relname=$eq.test", "GET", http.StatusOK},
		{"Get tables with custom order clause", "/tables?_order=c.relname", "GET", http.StatusOK},
		{"Get tables with custom where clause and pagination", "/tables?c.relname=$eq.test&_page=1&_page_size=20", "GET", http.StatusOK},
		{"Get tables with COUNT clause", "/tables?_count=*", "GET", http.StatusOK},
		{"Get tables with custom where invalid clause", "/tables?0c.relname=$eq.test", "GET", http.StatusBadRequest},
		{"Get tables with ORDER BY and invalid column", "/tables?_order=0c.relname", "GET", http.StatusBadRequest},
		{"Get tables with noexistent column", "/tables?c.rolooo=$eq.test", "GET", http.StatusBadRequest},
	}

	router := mux.NewRouter()
	router.HandleFunc("/tables", GetTables).Methods("GET")
	server := httptest.NewServer(router)
	defer server.Close()

	for _, tc := range testCases {
		t.Log(tc.description)
		doRequest(t, server.URL+tc.url, nil, tc.method, tc.status, "GetTables")
	}
}

func TestGetTablesByDatabaseAndSchema(t *testing.T) {
	var testCases = []struct {
		description string
		url         string
		method      string
		status      int
	}{
		{"Get tables by database and schema without custom where clause", "/prest/public", "GET", http.StatusOK},
		{"Get tables by database and schema with custom where clause", "/prest/public?t.tablename=$eq.test", "GET", http.StatusOK},
		{"Get tables by database and schema with order clause", "/prest/public?t.tablename=$eq.test&_order=t.tablename", "GET", http.StatusOK},
		{"Get tables by database and schema with custom where clause and pagination", "/prest/public?t.tablename=$eq.test&_page=1&_page_size=20", "GET", http.StatusOK},
		// errors
		{"Get tables by database and schema with custom where invalid clause", "/prest/public?0t.tablename=$eq.test", "GET", http.StatusBadRequest},
		{"Get tables by databases and schema with custom where and pagination invalid", "/prest/public?t.tablename=$eq.test&_page=A&_page_size=20", "GET", http.StatusBadRequest},
		{"Get tables by databases and schema with ORDER BY and column invalid", "/prest/public?_order=0t.tablename", "GET", http.StatusBadRequest},
		{"Get tables by databases with noexistent column", "/prest/public?t.taababa=$eq.test", "GET", http.StatusBadRequest},
	}

	router := mux.NewRouter()
	router.HandleFunc("/{database}/{schema}", GetTablesByDatabaseAndSchema).Methods("GET")
	server := httptest.NewServer(router)
	defer server.Close()

	for _, tc := range testCases {
		t.Log(tc.description)
		doRequest(t, server.URL+tc.url, nil, tc.method, tc.status, "GetTablesByDatabaseAndSchema")
	}
}

func TestSelectFromTables(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/{database}/{schema}/{table}", SelectFromTables).Methods("GET")
	server := httptest.NewServer(router)
	defer server.Close()

	var testCases = []struct {
		description string
		url         string
		method      string
		status      int
		body        string
	}{
		{"execute select in a table with array", "/prest/public/testarray", "GET", http.StatusOK, "[{\"id\":100,\"data\":[\"Gohan\",\"Goten\"]}]"},
		{"execute select in a table without custom where clause", "/prest/public/test", "GET", http.StatusOK, ""},
		{"execute select in a table with count all fields *", "/prest/public/test?_count=*", "GET", http.StatusOK, ""},
		{"execute select in a table with count function", "/prest/public/test?_count=name", "GET", http.StatusOK, ""},
		{"execute select in a table with custom where clause", "/prest/public/test?name=$eq.nuveo", "GET", http.StatusOK, ""},
		{"execute select in a table with custom join clause", "/prest/public/test?_join=inner:test8:test8.nameforjoin:$eq:test.name", "GET", http.StatusOK, ""},
		{"execute select in a table with order clause", "/prest/public/test?_order=name", "GET", http.StatusOK, ""},
		{"execute select in a table with order clause empty", "/prest/public/test?_order=", "GET", http.StatusOK, ""},
		{"execute select in a table with custom where clause and pagination", "/prest/public/test?name=$eq.nuveo&_page=1&_page_size=20", "GET", http.StatusOK, ""},
		{"execute select in a table with select fields", "/prest/public/test5?_select=celphone,name", "GET", http.StatusOK, ""},
		{"execute select in a table with select *", "/prest/public/test5?_select=*", "GET", http.StatusOK, ""},
		{"execute select in a view without custom where clause", "/prest/public/view_test", "GET", http.StatusOK, ""},
		{"execute select in a view with count all fields *", "/prest/public/view_test?_count=*", "GET", http.StatusOK, ""},
		{"execute select in a view with count function", "/prest/public/view_test?_count=player", "GET", http.StatusOK, ""},
		{"execute select in a view with order function", "/prest/public/view_test?_order=-player", "GET", http.StatusOK, ""},
		{"execute select in a view with custom where clause", "/prest/public/view_test?player=$eq.gopher", "GET", http.StatusOK, ""},
		{"execute select in a view with custom join clause", "/prest/public/view_test?_join=inner:test2:test2.name:eq:view_test.player", "GET", http.StatusOK, ""},
		{"execute select in a view with custom where clause and pagination", "/prest/public/view_test?player=$eq.gopher&_page=1&_page_size=20", "GET", http.StatusOK, ""},
		{"execute select in a view with select fields", "/prest/public/view_test?_select=player", "GET", http.StatusOK, ""},

		// errors
		{"execute select in a table with invalid join clause", "/prest/public/test?_join=inner:test2:test2.name", "GET", http.StatusBadRequest, ""},
		{"execute select in a table with invalid where clause", "/prest/public/test?0name=$eq.nuveo", "GET", http.StatusBadRequest, ""},
		{"execute select in a table with order clause and column invalid", "/prest/public/test?_order=0name", "GET", http.StatusBadRequest, ""},
		{"execute select in a table with invalid pagination clause", "/prest/public/test?name=$eq.nuveo&_page=A", "GET", http.StatusBadRequest, ""},
		{"execute select in a table with invalid where clause", "/prest/public/test?0name=$eq.nuveo", "GET", http.StatusBadRequest, ""},
		{"execute select in a table with invalid count clause", "/prest/public/test?_count=0name", "GET", http.StatusBadRequest, ""},
		{"execute select in a table with invalid order clause", "/prest/public/test?_order=0name", "GET", http.StatusBadRequest, ""},
		{"execute select in a view with an other column", "/prest/public/view_test?_select=celphone", "GET", http.StatusBadRequest, ""},
		{"execute select in a view with where and column invalid", "/prest/public/view_test?0celphone=$eq.888888", "GET", http.StatusBadRequest, ""},
		{"execute select in a view with custom join clause invalid", "/prest/public/view_test?_join=inner:test2.name:eq:view_test.player", "GET", http.StatusBadRequest, ""},
		{"execute select in a view with custom where clause and pagination invalid", "/prest/public/view_test?player=$eq.gopher&_page=A&_page_size=20", "GET", http.StatusBadRequest, ""},
		{"execute select in a view with order by and column invalid", "/prest/public/view_test?_order=0celphone", "GET", http.StatusBadRequest, ""},
		{"execute select in a view with count column invalid", "/prest/public/view_test?_count=0celphone", "GET", http.StatusBadRequest, ""},
	}

	for _, tc := range testCases {
		t.Log(tc.description)
		if tc.body != "" {
			doRequest(t, server.URL+tc.url, nil, tc.method, tc.status, "SelectFromTables", tc.body)
			continue
		}
		doRequest(t, server.URL+tc.url, nil, tc.method, tc.status, "SelectFromTables")
	}
}

func TestInsertInTables(t *testing.T) {
	m := make(map[string]interface{})
	m["name"] = "prest"

	mJSON := make(map[string]interface{})
	mJSON["name"] = "prest"
	mJSON["data"] = `{"term": "name", "subterm": ["names", "of", "subterms"], "obj": {"emp": "nuveo"}}`

	mARRAY := make(map[string]interface{})
	mARRAY["data"] = []string{"value 1", "value 2", "value 3"}

	router := mux.NewRouter()
	router.HandleFunc("/{database}/{schema}/{table}", InsertInTables).Methods("POST")
	server := httptest.NewServer(router)
	defer server.Close()

	var testCases = []struct {
		description string
		url         string
		request     map[string]interface{}
		status      int
	}{
		{"execute insert in a table with array field", "/prest/public/testarray", mARRAY, http.StatusOK},
		{"execute insert in a table with jsonb field", "/prest/public/testjson", mJSON, http.StatusOK},
		{"execute insert in a table without custom where clause", "/prest/public/test", m, http.StatusOK},
		{"execute insert in a table with invalid database", "/0prest/public/test", m, http.StatusBadRequest},
		{"execute insert in a table with invalid schema", "/prest/0public/test", m, http.StatusBadRequest},
		{"execute insert in a table with invalid table", "/prest/public/0test", m, http.StatusBadRequest},
		{"execute insert in a table with invalid body", "/prest/public/test", nil, http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Log(tc.description)
		doRequest(t, server.URL+tc.url, tc.request, "POST", tc.status, "InsertInTables")
	}
}

func TestDeleteFromTable(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/{database}/{schema}/{table}", DeleteFromTable).Methods("DELETE")
	server := httptest.NewServer(router)
	defer server.Close()

	var testCases = []struct {
		description string
		url         string
		request     map[string]interface{}
		status      int
	}{
		{"execute delete in a table without custom where clause", "/prest/public/test", nil, http.StatusOK},
		{"excute delete in a table with where clause", "/prest/public/test?name=$eq.nuveo", nil, http.StatusOK},
		{"execute delete in a table with invalid database", "/0prest/public/test", nil, http.StatusBadRequest},
		{"execute delete in a table with invalid schema", "/prest/0public/test", nil, http.StatusBadRequest},
		{"execute delete in a table with invalid table", "/prest/public/0test", nil, http.StatusBadRequest},
		{"execute delete in a table with invalid where clause", "/prest/public/test?0name=$eq.nuveo", nil, http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Log(tc.description)
		doRequest(t, server.URL+tc.url, tc.request, "DELETE", tc.status, "DeleteFromTable")
	}
}

func TestUpdateFromTable(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/{database}/{schema}/{table}", UpdateTable).Methods("PUT", "PATCH")
	server := httptest.NewServer(router)
	defer server.Close()

	m := make(map[string]interface{}, 0)
	m["name"] = "prest"

	var testCases = []struct {
		description string
		url         string
		request     map[string]interface{}
		status      int
	}{
		{"execute update in a table without custom where clause", "/prest/public/test", m, http.StatusOK},
		{"excute update in a table with where clause", "/prest/public/test?name=$eq.nuveo", m, http.StatusOK},
		{"execute update in a table with invalid database", "/0prest/public/test", m, http.StatusBadRequest},
		{"execute update in a table with invalid schema", "/prest/0public/test", m, http.StatusBadRequest},
		{"execute update in a table with invalid table", "/prest/public/0test", m, http.StatusBadRequest},
		{"execute update in a table with invalid where clause", "/prest/public/test?0name=$eq.nuveo", m, http.StatusBadRequest},
		{"execute update in a table with invalid body", "/prest/public/test?name=$eq.nuveo", nil, http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Log(tc.description)
		doRequest(t, server.URL+tc.url, tc.request, "PUT", tc.status, "UpdateTable")
		doRequest(t, server.URL+tc.url, tc.request, "PATCH", tc.status, "UpdateTable")
	}
}

func TestRequestTimeout(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/{database}/{schema}/{table}", SelectFromTables).Methods("GET")
	server := httptest.NewServer(router)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Microsecond)
	defer cancel()
	req, err := http.NewRequest("GET", "/prest/public/test5", nil)
	if err != nil {
		t.Errorf("expected no errors, but has %v", err)
	}
	req = req.WithContext(ctx)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("expected no errors, but has %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status code 400, but got %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("expected no errors, but has %v", err)
	}
	if !strings.Contains(string(body), "context") {
		t.Error("do not contain a context error message")
	}
}
