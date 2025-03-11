package routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"microservice/internal/config"
	"microservice/routes"
)

func Test_IdentifyObject(t *testing.T) {
	router := gin.New()
	router.Use(config.Middlewares()...)
	router.GET("/identify", routes.IdentifyObject)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/identify?key=03", nil)
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

func Test_IdentifyObject_MissingKeys(t *testing.T) {
	router := gin.New()
	router.Use(config.Middlewares()...)
	router.GET("/identify", routes.IdentifyObject)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/identify", nil)
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
