//nolint:dupl
package v2Routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"microservice/internal/config"
	"microservice/middlewares"
	v2Routes "microservice/routes/v2"
)

func Test_LayerContents(t *testing.T) {
	router := gin.New()
	router.Use(config.Middlewares()...)
	router.GET("/v2/content/:layerID/", middlewares.ResolveV2Layer, v2Routes.AttributedLayerContents)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v2/content/1e694f36-cf68-426a-b6a3-7660163b03e6/", nil)
	router.ServeHTTP(w, req)

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

func Test_LayerContents_InvalidLayerID(t *testing.T) {
	router := gin.New()
	router.Use(config.Middlewares()...)
	router.GET("/v2/content/:layerID/", middlewares.ResolveLayer, v2Routes.AttributedLayerContents)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v2/content/invalid-layer-id/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
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

func Test_LayerContents_MissingLayerID(t *testing.T) {
	router := gin.New()
	router.Use(config.Middlewares()...)
	router.GET("/v2/content/:layerID/", middlewares.ResolveLayer, v2Routes.AttributedLayerContents)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v2/content//", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
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
