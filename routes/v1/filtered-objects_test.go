// nolint
package routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"microservice/internal/config"
	"microservice/middlewares"
	"microservice/routes/v1"
)

func Test_FilteredObjects(t *testing.T) {
	router := gin.New()
	router.Use(config.Middlewares()...)

	router.GET("/content/:layerID/filtered", middlewares.ResolveLayer, routes.FilteredLayerContents)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/content/1e694f36-cf68-426a-b6a3-7660163b03e6/filtered?relation=contains&other_layer=e517edaa-8d7b-4f10-9cfc-56a7c56109f0&key=03101", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	if t.Failed() {
		t.Log(w.Body.String())
	}

	valid, validationErrors := v.ValidateHttpRequestResponse(req, w.Result())
	if !valid {
		t.Fail()
		for _, e := range validationErrors {
			t.Logf("Type: %s, Failure: %s\n", e.ValidationType, e.Message)
			if e.SchemaValidationErrors != nil {
				t.Logf("Schema Error: %s, Line: %d, Col: %d\n",
					e.SchemaValidationErrors[0].Reason,
					e.SchemaValidationErrors[0].Line,
					e.SchemaValidationErrors[0].Column)
			}
		}
	}
}

func Test_FilteredObjects_NoObjects(t *testing.T) {
	router := gin.New()
	router.Use(config.Middlewares()...)

	router.GET("/content/:layerID/filtered", middlewares.ResolveLayer, routes.FilteredLayerContents)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/content/e1fde37d-69aa-43fc-8338-588dc09f7ff2/filtered?relation=within&other_layer=1e694f36-cf68-426a-b6a3-7660163b03e6&key=02102", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	if t.Failed() {
		t.Log(w.Body.String())
	}

	valid, validationErrors := v.ValidateHttpRequestResponse(req, w.Result())
	if !valid {
		t.Fail()
		for _, e := range validationErrors {
			t.Logf("Type: %s, Failure: %s\n", e.ValidationType, e.Message)
			if e.SchemaValidationErrors != nil {
				t.Logf("Schema Error: %s, Line: %d, Col: %d\n",
					e.SchemaValidationErrors[0].Reason,
					e.SchemaValidationErrors[0].Line,
					e.SchemaValidationErrors[0].Column)
			}
		}
	}
}

func Test_FilteredObjects_InvalidBaseLayer(t *testing.T) {
	router := gin.New()
	router.Use(config.Middlewares()...)

	router.GET("/content/:layerID/filtered", middlewares.ResolveLayer, routes.FilteredLayerContents)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/content/invalid-baselayer/filtered?relation=within&other_layer=e517edaa-8d7b-4f10-9cfc-56a7c56109f0&key=03101", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	valid, validationErrors := v.ValidateHttpRequestResponse(req, w.Result())
	if !valid {
		t.Fail()
		for _, e := range validationErrors {
			t.Logf("Type: %s, Failure: %s\n", e.ValidationType, e.Message)
			if e.SchemaValidationErrors != nil {
				t.Logf("Schema Error: %s, Line: %d, Col: %d\n",
					e.SchemaValidationErrors[0].Reason,
					e.SchemaValidationErrors[0].Line,
					e.SchemaValidationErrors[0].Column)
			}
		}
	}
}

func Test_FilteredObjects_InvalidTopLayer(t *testing.T) {
	router := gin.New()
	router.Use(config.Middlewares()...)

	router.GET("/content/:layerID/filtered", middlewares.ResolveLayer, routes.FilteredLayerContents)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/content/1e694f36-cf68-426a-b6a3-7660163b03e6/filtered?relation=within&other_layer=X517edaa-8d7b-4f10-9cfc-56a7c56109f0&key=03101", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	valid, validationErrors := v.ValidateHttpRequestResponse(req, w.Result())
	if !valid {
		t.Fail()
		for _, e := range validationErrors {
			t.Logf("Type: %s, Failure: %s\n", e.ValidationType, e.Message)
			if e.SchemaValidationErrors != nil {
				t.Logf("Schema Error: %s, Line: %d, Col: %d\n",
					e.SchemaValidationErrors[0].Reason,
					e.SchemaValidationErrors[0].Line,
					e.SchemaValidationErrors[0].Column)
			}
		}
	}
}

func Test_FilteredObjects_MissingRelation(t *testing.T) {
	router := gin.New()
	router.Use(config.Middlewares()...)

	router.GET("/content/:layerID/filtered", middlewares.ResolveLayer, routes.FilteredLayerContents)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/content/1e694f36-cf68-426a-b6a3-7660163b03e6/filtered?other_layer=e517edaa-8d7b-4f10-9cfc-56a7c56109f0&key=03101", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	if t.Failed() {
		t.Log(w.Body.String())
	}

	valid, validationErrors := v.ValidateHttpResponse(req, w.Result())
	if !valid {
		t.Fail()
		for _, e := range validationErrors {
			t.Logf("Type: %s, Failure: %s\n", e.ValidationType, e.Message)
			if e.SchemaValidationErrors != nil {
				t.Logf("Schema Error: %s, Line: %d, Col: %d\n",
					e.SchemaValidationErrors[0].Reason,
					e.SchemaValidationErrors[0].Line,
					e.SchemaValidationErrors[0].Column)
			}
		}
	}
}

func Test_FilteredObjects_MissingOtherLayer(t *testing.T) {
	router := gin.New()
	router.Use(config.Middlewares()...)

	router.GET("/content/:layerID/filtered", middlewares.ResolveLayer, routes.FilteredLayerContents)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/content/1e694f36-cf68-426a-b6a3-7660163b03e6/filtered?relation=within&key=03101", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	if t.Failed() {
		t.Log(w.Body.String())
	}

	valid, validationErrors := v.ValidateHttpResponse(req, w.Result())
	if !valid {
		t.Fail()
		for _, e := range validationErrors {
			t.Logf("Type: %s, Failure: %s\n", e.ValidationType, e.Message)
			if e.SchemaValidationErrors != nil {
				t.Logf("Schema Error: %s, Line: %d, Col: %d\n",
					e.SchemaValidationErrors[0].Reason,
					e.SchemaValidationErrors[0].Line,
					e.SchemaValidationErrors[0].Column)
			}
		}
	}
}

func Test_FilteredObjects_MissingKeys(t *testing.T) {
	router := gin.New()
	router.Use(config.Middlewares()...)

	router.GET("/content/:layerID/filtered", middlewares.ResolveLayer, routes.FilteredLayerContents)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/content/1e694f36-cf68-426a-b6a3-7660163b03e6/filtered?other_layer=e517edaa-8d7b-4f10-9cfc-56a7c56109f0", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	if t.Failed() {
		t.Log(w.Body.String())
	}

	valid, validationErrors := v.ValidateHttpResponse(req, w.Result())
	if !valid {
		t.Fail()
		for _, e := range validationErrors {
			t.Logf("Type: %s, Failure: %s\n", e.ValidationType, e.Message)
			if e.SchemaValidationErrors != nil {
				t.Logf("Schema Error: %s, Line: %d, Col: %d\n",
					e.SchemaValidationErrors[0].Reason,
					e.SchemaValidationErrors[0].Line,
					e.SchemaValidationErrors[0].Column)
			}
		}
	}
}
