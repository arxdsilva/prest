package controllers

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/prest/prest/adapters/postgres"
	"github.com/prest/prest/config"
	"github.com/prest/prest/testutils"
)

func init() {
	config.Load()
	postgres.Load()
}

func TestGetTables(t *testing.T) {
	var testCases = []struct {
		description string
		url         string
		method      string
		status      int
	}{
		{"Get tables without custom where clause", "/tables/prest-test", "GET", http.StatusOK},
		{"Get tables with custom where clause", "/tables/prest-test?c.relname=$eq.test", "GET", http.StatusOK},
		{"Get tables with custom order clause", "/tables/prest-test?_order=c.relname", "GET", http.StatusOK},
		{"Get tables with custom where clause and pagination", "/tables/prest-test?c.relname=$eq.test&_page=1&_page_size=20", "GET", http.StatusOK},
		{"Get tables with COUNT clause", "/tables/prest-test?_count=*", "GET", http.StatusOK},
		{"Get tables with distinct clause", "/tables/prest-test?_distinct=true", "GET", http.StatusOK},
		{"Get tables with custom where invalid clause", "/tables/prest-test?0c.relname=$eq.test", "GET", http.StatusBadRequest},
		{"Get tables with ORDER BY and invalid column", "/tables/prest-test?_order=0c.relname", "GET", http.StatusBadRequest},
		{"Get tables with noexistent column", "/tables/prest-test?c.rolooo=$eq.test", "GET", http.StatusBadRequest},
	}

	router := mux.NewRouter()
	router.HandleFunc("/tables/{database}", GetTables).Methods("GET")
	server := httptest.NewServer(router)
	defer server.Close()

	for _, tc := range testCases {
		t.Log(tc.description)
		testutils.DoRequest(t, server.URL+tc.url, nil, tc.method, tc.status, "GetTables")
	}
}

func TestGetTablesByDatabaseAndSchema(t *testing.T) {
	var testCases = []struct {
		description string
		url         string
		method      string
		status      int
	}{
		{"Get tables by database and schema without custom where clause", "/schemas/prest-test/public", "GET", http.StatusOK},
		{"Get tables by database and schema with custom where clause", "/schemas/prest-test/public?t.tablename=$eq.test", "GET", http.StatusOK},
		{"Get tables by database and schema with order clause", "/schemas/prest-test/public?t.tablename=$eq.test&_order=t.tablename", "GET", http.StatusOK},
		{"Get tables by database and schema with custom where clause and pagination", "/schemas/prest-test/public?t.tablename=$eq.test&_page=1&_page_size=20", "GET", http.StatusOK},
		{"Get tables by database and schema with distinct clause", "/schemas/prest-test/public?_distinct=true", "GET", http.StatusOK},
		// errors
		{"Get tables by database and schema with custom where invalid clause", "/schemas/prest-test/public?0t.tablename=$eq.test", "GET", http.StatusBadRequest},
		{"Get tables by databases and schema with custom where and pagination invalid", "/schemas/prest-test/public?t.tablename=$eq.test&_page=A&_page_size=20", "GET", http.StatusBadRequest},
		{"Get tables by databases and schema with ORDER BY and column invalid", "/schemas/prest-test/public?_order=0t.tablename", "GET", http.StatusBadRequest},
		{"Get tables by databases with noexistent column", "/schemas/prest-test/public?t.taababa=$eq.test", "GET", http.StatusBadRequest},
		{"Get tables by databases with not configured database", "/schemas/random/public?t.taababa=$eq.test", "GET", http.StatusBadRequest},
	}

	router := mux.NewRouter()
	router.HandleFunc("/schemas/{database}/{schema}", GetTablesByDatabaseAndSchema).Methods("GET")
	server := httptest.NewServer(router)
	defer server.Close()

	for _, tc := range testCases {
		t.Log(tc.description)
		testutils.DoRequest(t, server.URL+tc.url, nil, tc.method, tc.status, "GetTablesByDatabaseAndSchema")
	}
}

func TestSelectFromTables(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/operations/{database}/{schema}/{table}", SelectFromTables).Methods("GET")
	server := httptest.NewServer(router)
	defer server.Close()

	var testCases = []struct {
		description string
		url         string
		method      string
		status      int
		body        string
	}{
		{"execute select in a table with array", "/operations/prest-test/public/testarray", "GET", http.StatusOK, "[{\"id\":100,\"data\":[\"Gohan\",\"Goten\"]}]"},
		{"execute select in a table without custom where clause", "/operations/prest-test/public/test", "GET", http.StatusOK, ""},
		{"execute select in a table case sentive", "/operations/prest-test/public/Reply", "GET", http.StatusOK, "[{\"id\":1,\"name\":\"prest tester\"}, \n {\"id\":2,\"name\":\"prest-test-insert\"}, \n {\"id\":3,\"name\":\"3prest-test-batch-insert\"}, \n {\"id\":4,\"name\":\"3batch-prest-test-insert\"}, \n {\"id\":5,\"name\":\"copy\"}, \n {\"id\":6,\"name\":\"copy\"}]"},
		{"execute select in a table with count all fields *", "/operations/prest-test/public/test?_count=*", "GET", http.StatusOK, ""},
		{"execute select in a table with count function", "/operations/prest-test/public/test?_count=name", "GET", http.StatusOK, ""},
		{"execute select in a table with custom where clause", "/operations/prest-test/public/test?name=$eq.test", "GET", http.StatusOK, ""},
		{"execute select in a table with custom join clause", "/operations/prest-test/public/test?_join=inner:test8:test8.nameforjoin:$eq:test.name", "GET", http.StatusOK, ""},
		{"execute select in a table with order clause empty", "/operations/prest-test/public/test?_order=", "GET", http.StatusOK, ""},
		{"execute select in a table with custom where clause and pagination", "/operations/prest-test/public/test?name=$eq.test&_page=1&_page_size=20", "GET", http.StatusOK, ""},
		{"execute select in a table with select fields", "/operations/prest-test/public/test5?_select=celphone,name", "GET", http.StatusOK, ""},
		{"execute select in a table with select *", "/operations/prest-test/public/test5?_select=*", "GET", http.StatusOK, ""},
		{"execute select in a table with select * and distinct", "/operations/prest-test/public/test5?_select=*&_distinct=true", "GET", http.StatusOK, ""},

		{"execute select in a table with group by clause", "/operations/prest-test/public/test_group_by_table?_select=age,sum:salary&_groupby=age", "GET", http.StatusOK, ""},
		{"execute select in a table with group by and having clause", "/operations/prest-test/public/test_group_by_table?_select=age,sum:salary&_groupby=age->>having:sum:salary:$gt:3000", "GET", http.StatusOK, "[{\"age\":19,\"sum\":7997}]"},

		{"execute select in a view without custom where clause", "/operations/prest-test/public/view_test", "GET", http.StatusOK, ""},
		{"execute select in a view with count all fields *", "/operations/prest-test/public/view_test?_count=*", "GET", http.StatusOK, ""},
		{"execute select in a view with count function", "/operations/prest-test/public/view_test?_count=player", "GET", http.StatusOK, ""},
		{"execute select in a view with order function", "/operations/prest-test/public/view_test?_order=-player", "GET", http.StatusOK, ""},
		{"execute select in a view with custom where clause", "/operations/prest-test/public/view_test?player=$eq.gopher", "GET", http.StatusOK, ""},
		{"execute select in a view with custom join clause", "/operations/prest-test/public/view_test?_join=inner:test2:test2.name:eq:view_test.player", "GET", http.StatusOK, ""},
		{"execute select in a view with custom where clause and pagination", "/operations/prest-test/public/view_test?player=$eq.gopher&_page=1&_page_size=20", "GET", http.StatusOK, ""},
		{"execute select in a view with select fields", "/operations/prest-test/public/view_test?_select=player", "GET", http.StatusOK, ""},

		{"execute select in a table with invalid join clause", "/operations/prest-test/public/test?_join=inner:test2:test2.name", "GET", http.StatusBadRequest, ""},
		{"execute select in a table with invalid where clause", "/operations/prest-test/public/test?0name=$eq.test", "GET", http.StatusBadRequest, ""},
		{"execute select in a table with order clause and column invalid", "/operations/prest-test/public/test?_order=0name", "GET", http.StatusBadRequest, ""},
		{"execute select in a table with invalid pagination clause", "/operations/prest-test/public/test?name=$eq.test&_page=A", "GET", http.StatusBadRequest, ""},
		{"execute select in a table with invalid where clause", "/operations/prest-test/public/test?0name=$eq.test", "GET", http.StatusBadRequest, ""},
		{"execute select in a table with invalid count clause", "/operations/prest-test/public/test?_count=0name", "GET", http.StatusBadRequest, ""},
		{"execute select in a table with invalid order clause", "/operations/prest-test/public/test?_order=0name", "GET", http.StatusBadRequest, ""},
		{"execute select in a table with invalid fields using group by clause", "/operations/prest-test/public/test_group_by_table?_select=pa,sum:pum&_groupby=pa", "GET", http.StatusBadRequest, ""},
		{"execute select in a table with invalid fields using group by and having clause", "/operations/prest-test/public/test_group_by_table?_select=pa,sum:pum&_groupby=pa->>having:sum:pmu:$eq:150", "GET", http.StatusBadRequest, ""},

		{"execute select in a view with an other column", "/operations/prest-test/public/view_test?_select=celphone", "GET", http.StatusBadRequest, ""},
		{"execute select in a view with where and column invalid", "/operations/prest-test/public/view_test?0celphone=$eq.888888", "GET", http.StatusBadRequest, ""},
		{"execute select in a view with custom join clause invalid", "/operations/prest-test/public/view_test?_join=inner:test2.name:eq:view_test.player", "GET", http.StatusBadRequest, ""},
		{"execute select in a view with custom where clause and pagination invalid", "/operations/prest-test/public/view_test?player=$eq.gopher&_page=A&_page_size=20", "GET", http.StatusBadRequest, ""},
		{"execute select in a view with order by and column invalid", "/operations/prest-test/public/view_test?_order=0celphone", "GET", http.StatusBadRequest, ""},
		{"execute select in a view with count column invalid", "/operations/prest-test/public/view_test?_count=0celphone", "GET", http.StatusBadRequest, ""},

		{"execute select in a db that does not exist", "/operations/invalid/public/view_test?_count=0celphone", "GET", http.StatusBadRequest, ""},
	}

	for _, tc := range testCases {
		t.Log(tc.description)
		//config.PrestConf = &config.Prest{}
		//config.Load()

		if tc.body != "" {
			testutils.DoRequest(t, server.URL+tc.url, nil, tc.method, tc.status, "SelectFromTables", tc.body)
			continue
		}
		testutils.DoRequest(t, server.URL+tc.url, nil, tc.method, tc.status, "SelectFromTables")
	}
}

func TestInsertInTables(t *testing.T) {
	m := make(map[string]interface{})
	m["name"] = "prest-test"

	mJSON := make(map[string]interface{})
	mJSON["name"] = "prest-test"
	mJSON["data"] = `{"term": "name", "subterm": ["names", "of", "subterms"], "obj": {"emp": "prestd"}}`

	mARRAY := make(map[string]interface{})
	mARRAY["data"] = []string{"value 1", "value 2", "value 3"}

	router := mux.NewRouter()
	router.HandleFunc("/operations/{database}/{schema}/{table}", InsertInTables).Methods("POST")
	server := httptest.NewServer(router)
	defer server.Close()

	var testCases = []struct {
		description string
		url         string
		request     map[string]interface{}
		status      int
	}{
		{"execute insert in a table with array field", "/operations/prest-test/public/testarray", mARRAY, http.StatusCreated},
		{"execute insert in a table with jsonb field", "/operations/prest-test/public/testjson", mJSON, http.StatusCreated},
		{"execute insert in a table without custom where clause", "/operations/prest-test/public/test", m, http.StatusCreated},
		{"execute insert in a table with invalid database", "/operations/0prest-test/public/test", m, http.StatusBadRequest},
		{"execute insert in a table with invalid schema", "/operations/prest-test/0public/test", m, http.StatusNotFound},
		{"execute insert in a table with invalid table", "/operations/prest-test/public/0test", m, http.StatusNotFound},
		{"execute insert in a table with invalid body", "/operations/prest-test/public/test", nil, http.StatusBadRequest},

		{"execute insert in a database that does not exist", "/operations/invalid/public/0test", m, http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Log(tc.description)
		testutils.DoRequest(t, server.URL+tc.url, tc.request, "POST", tc.status, "InsertInTables")
	}
}

func TestBatchInsertInTables(t *testing.T) {

	m := make([]map[string]interface{}, 0)
	m = append(m, map[string]interface{}{"name": "bprest"}, map[string]interface{}{"name": "aprest"})

	mJSON := make([]map[string]interface{}, 0)
	mJSON = append(mJSON, map[string]interface{}{"name": "cprest", "data": `{"term": "name", "subterm": ["names", "of", "subterms"], "obj": {"emp": "prestd"}}`}, map[string]interface{}{"name": "dprest", "data": `{"term": "name", "subterms": ["names", "of", "subterms"], "obj": {"emp": "prestd"}}`})

	mARRAY := make([]map[string]interface{}, 0)
	mARRAY = append(mARRAY, map[string]interface{}{"data": []string{"1", "2"}}, map[string]interface{}{"data": []string{"1", "2", "3"}})

	router := mux.NewRouter()
	router.HandleFunc("/operations/batch/{database}/{schema}/{table}", BatchInsertInTables).Methods("POST")
	server := httptest.NewServer(router)
	defer server.Close()

	var testCases = []struct {
		description string
		url         string
		request     []map[string]interface{}
		status      int
		isCopy      bool
	}{
		{"execute insert in a table with array field", "/operations/batch/prest-test/public/testarray", mARRAY, http.StatusCreated, false},
		{"execute insert in a table with jsonb field", "/operations/batch/prest-test/public/testjson", mJSON, http.StatusCreated, false},
		{"execute insert in a table without custom where clause", "/operations/batch/prest-test/public/test", m, http.StatusCreated, false},
		{"execute insert in a table with invalid database", "/operations/batch/0prest-test/public/test", m, http.StatusBadRequest, false},
		{"execute insert in a table with invalid schema", "/operations/batch/prest-test/0public/test", m, http.StatusNotFound, false},
		{"execute insert in a table with invalid table", "/operations/batch/prest-test/public/0test", m, http.StatusNotFound, false},
		{"execute insert in a table with invalid body", "/operations/batch/prest-test/public/test", nil, http.StatusBadRequest, false},
		{"execute insert in a table with array field with copy", "/operations/batch/prest-test/public/testarray", mARRAY, http.StatusCreated, true},
		{"execute insert in a table with jsonb field with copy", "/operations/batch/prest-test/public/testjson", mJSON, http.StatusCreated, true},

		{"execute insert in a db that does not exist", "/operations/batch/invalid/public/test", nil, http.StatusBadRequest, false},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			byt, err := json.Marshal(tc.request)
			if err != nil {
				t.Error("error on json marshal", err)
			}
			req, err := http.NewRequest(http.MethodPost, server.URL+tc.url, bytes.NewReader(byt))
			if err != nil {
				t.Error("error on New Request", err)
			}
			if tc.isCopy {
				req.Header.Set("Prest-Batch-Method", "copy")
			}
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Error("error on Do Request", err)
			}
			if resp.StatusCode != tc.status {
				t.Errorf("expected %d, got: %d", tc.status, resp.StatusCode)
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Error("error on ioutil ReadAll", err)
			}
			if tc.isCopy && len(body) != 0 {
				t.Errorf("len body is %d", len(body))
			}
		})
	}
}

func TestDeleteFromTable(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/operations/{database}/{schema}/{table}", DeleteFromTable).Methods("DELETE")
	server := httptest.NewServer(router)
	defer server.Close()

	var testCases = []struct {
		description string
		url         string
		request     map[string]interface{}
		status      int
	}{
		{"execute delete in a table without custom where clause", "/operations/prest-test/public/test", nil, http.StatusOK},
		{"excute delete in a table with where clause", "/operations/prest-test/public/test?name=$eq.test", nil, http.StatusOK},
		{"execute delete in a table with invalid database", "/operatioons/0prest-test/public/test", nil, http.StatusBadRequest},
		{"execute delete in a table with invalid schema", "/operations/prest-test/0public/test", nil, http.StatusNotFound},
		{"execute delete in a table with invalid table", "/operations/prest-test/public/0test", nil, http.StatusNotFound},
		{"execute delete in a table with invalid where clause", "/operations/prest-test/public/test?0name=$eq.nuveo", nil, http.StatusBadRequest},
		// error
		{"execute delete in a invalid db", "/operations/invalid/public/0test", nil, http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Log(tc.description)
		testutils.DoRequest(t, server.URL+tc.url, tc.request, "DELETE", tc.status, "DeleteFromTable")
	}
}

func TestUpdateFromTable(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/operations/{database}/{schema}/{table}", UpdateTable).Methods("PUT", "PATCH")
	server := httptest.NewServer(router)
	defer server.Close()

	m := make(map[string]interface{})
	m["name"] = "prest"

	var testCases = []struct {
		description string
		url         string
		request     map[string]interface{}
		status      int
	}{
		{"execute update in a table without custom where clause", "/operations/prest-test/public/test", m, http.StatusOK},
		{"execute update in a table with where clause", "/operations/prest-test/public/test?name=$eq.test", m, http.StatusOK},
		{"execute update in a table with where clause and returning all fields", "/operations/prest-test/public/test?id=1&_returning=*", m, http.StatusOK},
		{"execute update in a table with where clause and returning name field", "/operations/prest-test/public/test?id=2&_returning=name", m, http.StatusOK},
		{"execute update in a table with invalid database", "/operations/0prest-test/public/test", m, http.StatusBadRequest},
		{"execute update in a table with invalid schema", "/operations/prest-test/0public/test", m, http.StatusNotFound},
		{"execute update in a table with invalid table", "/operations/prest-test/public/0test", m, http.StatusNotFound},
		{"execute update in a table with invalid where clause", "/operations/prest-test/public/test?0name=$eq.nuveo", m, http.StatusBadRequest},
		{"execute update in a table with invalid body", "/operations/prest-test/public/test?name=$eq.nuveo", nil, http.StatusBadRequest},
		// error
		{"execute update in a invalid db", "/operations/invalid/public/test", m, http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Log(tc.description)

		testutils.DoRequest(t, server.URL+tc.url, tc.request, "PUT", tc.status, "UpdateTable")
		testutils.DoRequest(t, server.URL+tc.url, tc.request, "PATCH", tc.status, "UpdateTable")
	}
}

func TestShowTable(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/show/{database}/{schema}/{table}", ShowTable).Methods("GET")
	server := httptest.NewServer(router)
	defer server.Close()

	var testCases = []struct {
		description string
		url         string
		method      string
		status      int
	}{
		{"execute select in a table test custom information table", "/show/prest-test/public/test", "GET", http.StatusOK},
		{"execute select in a table test2 custom information table", "/show/prest-test/public/test2", "GET", http.StatusOK},
		{"execute select in a invalid db", "/show/invalid/public/test2", "GET", http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Log(tc.description)
		testutils.DoRequest(t, server.URL+tc.url, nil, tc.method, tc.status, "ShowTable")
	}
}
