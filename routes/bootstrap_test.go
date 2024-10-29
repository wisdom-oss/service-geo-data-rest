package routes_test

import (
	"io"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi-validator"
)

var apiContract libopenapi.Document
var v validator.Validator

func TestMain(m *testing.M) {
	_ = godotenv.Load(".env", "../.env")

	apiContractFile, err := os.Open("../openapi.yaml")
	if err != nil {
		panic(err)
	}

	contents, err := io.ReadAll(apiContractFile)

	apiContract, err = libopenapi.NewDocumentWithTypeCheck(contents, false)
	if err != nil {
		panic(err)
	}
	var errs []error
	v, errs = validator.NewValidator(apiContract)
	if len(errs) > 0 {
		panic(errs)
	}
	os.Exit(m.Run())
}
