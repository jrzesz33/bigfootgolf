module bigfoot/golf/mcp

go 1.23.4

require (
	bigfoot/golf/common v0.0.0
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/gorilla/mux v1.8.1
	github.com/mark3labs/mcp-go v0.39.1
)

require (
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/invopop/jsonschema v0.13.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/neo4j/neo4j-go-driver/v5 v5.28.3 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace bigfoot/golf/common => ../pkg
