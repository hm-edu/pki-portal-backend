package api

import (
	"encoding/json"
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/swaggo/swag"
)

// ToOpenApi3 converts an existing swag.Spec to an OpenAPI3 version.
// Since swaggo does not support OpenAPI3 and OpenAPI2 does not support Bearer Authentication we
// simply patch the OpenAPI2 doc and wire that doc to the swagger UI.
func ToOpenApi3(spec *swag.Spec) (*openapi3.T, error) {
	var doc openapi2.T
	err := json.Unmarshal([]byte(spec.ReadDoc()), &doc)
	if err != nil {
		return nil, err
	}
	v3, _ := openapi2conv.ToV3(&doc)
	v3.Components.SecuritySchemes = make(openapi3.SecuritySchemes)
	v3.Components.SecuritySchemes["API"] = &openapi3.SecuritySchemeRef{Value: openapi3.NewJWTSecurityScheme()}
	return v3, nil
}
