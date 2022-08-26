// This file contains all globally used variables and their default values
package vars

import (
	"database/sql"

	"microservice/structs"
)

// TODO: Change the service name and remove the comment
const ServiceName = "template-service"

var (
	/*
		ApiGatewayHost

		The host on which the api gateway resides. The api gateway used in the project is the Kong API Gateway
	*/
	ApiGatewayHost string
	/*
		ApiGatewayAdminPort

		The port on which the admin api of the gateway listens for incoming requests.
	*/
	ApiGatewayAdminPort string
	/*
		ApiGatewayServicePath

		The path under which the service is reachable. The path needs to start with a leading slash "/"
	*/
	ApiGatewayServicePath string
	/*
		HttpListenPort

		The port on which the internally used http server listens for new connections.

		Default value: 8000
	*/
	HttpListenPort string
	/*
		ExecuteHealthcheck

		An indicator which sets the behaviour of the main function of the microservice.
		If it is set to true the service will try to access itself after executing all init() steps.

		Default value: false
	*/
	ExecuteHealthcheck bool
	/*
		ScopeConfigFilePath

		The path under which the scope configuration for this service is stored.

		Default value: /microservice/res/scope.json
	*/
	ScopeConfigFilePath string
	/*
		Scope

		The parsed scope configuration file which is used by the Authorizer
	*/
	Scope *structs.ScopeInformation
	/*
		PostgresHost

		The host on which the consumer database is running on
	*/
	PostgresHost string
	/*
		PostgresUser

		The user used to access the postgres database
	*/
	PostgresUser string
	/*
		PostgresPassword

		The password of the user used to access the postgres database
	*/
	PostgresPassword string
	/*
		PostgresPort

		The port on which the postgres database listens on for new connections. Default value: 5432
	*/
	PostgresPort string
	/*
		PostgresConnection

		The postgres connection which has been made during the initialization.
		This connection is shared throughout the microservice
	*/
	PostgresConnection *sql.DB
)
