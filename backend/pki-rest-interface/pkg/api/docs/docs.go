// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "Source Code",
            "url": "https://github.com/hm-edu/portal-backend"
        },
        "license": {
            "name": "Apache License",
            "url": "https://github.com/hm-edu/portal-backend/blob/main/LICENSE"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/smime/": {
            "get": {
                "security": [
                    {
                        "API": []
                    }
                ],
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "SMIME"
                ],
                "summary": "SMIME List Endpoint",
                "parameters": [
                    {
                        "type": "string",
                        "name": "email",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "certificates",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/portal_apis.ListSmimeResponse_CertificateDetails"
                            }
                        }
                    },
                    "default": {
                        "description": "Error processing the request",
                        "schema": {
                            "$ref": "#/definitions/echo.HTTPError"
                        }
                    }
                }
            }
        },
        "/smime/csr": {
            "post": {
                "security": [
                    {
                        "API": []
                    }
                ],
                "description": "This endpoint handles a provided CSR. The validity of the CSR is checked and passed to the sectigo server in combination with the basic user information extracted from the JWT.\nThe server uses his own configuration, so the profile and the lifetime of the certificate can not be modified.\nAfterwards the new certificate is returned as X509 certificate.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "SMIME"
                ],
                "summary": "SMIME CSR Endpoint",
                "parameters": [
                    {
                        "description": "The CSR",
                        "name": "csr",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.CsrRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "certificate",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "default": {
                        "description": "Error processing the request",
                        "schema": {
                            "$ref": "#/definitions/echo.HTTPError"
                        }
                    }
                }
            }
        },
        "/smime/revoke": {
            "post": {
                "security": [
                    {
                        "API": []
                    }
                ],
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "SMIME"
                ],
                "summary": "SMIME Revoke Endpoint",
                "parameters": [
                    {
                        "description": "The serial of the certificate to revoke and the reason",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.RevokeRequest"
                        }
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    },
                    "default": {
                        "description": "Error processing the request",
                        "schema": {
                            "$ref": "#/definitions/echo.HTTPError"
                        }
                    }
                }
            }
        },
        "/ssl/": {
            "get": {
                "security": [
                    {
                        "API": []
                    }
                ],
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "SSL"
                ],
                "summary": "SSL List Endpoint",
                "responses": {
                    "200": {
                        "description": "Certificates",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/portal_apis.SslCertificateDetails"
                            }
                        }
                    },
                    "default": {
                        "description": "Error processing the request",
                        "schema": {
                            "$ref": "#/definitions/echo.HTTPError"
                        }
                    }
                }
            }
        },
        "/ssl/active": {
            "get": {
                "security": [
                    {
                        "API": []
                    }
                ],
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "SSL"
                ],
                "summary": "SSL List active certificates Endpoint",
                "parameters": [
                    {
                        "type": "string",
                        "description": "domain search by domain",
                        "name": "domain",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Certificates",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/portal_apis.SslCertificateDetails"
                            }
                        }
                    },
                    "default": {
                        "description": "Error processing the request",
                        "schema": {
                            "$ref": "#/definitions/echo.HTTPError"
                        }
                    }
                }
            }
        },
        "/ssl/csr": {
            "post": {
                "security": [
                    {
                        "API": []
                    }
                ],
                "description": "This endpoint handles a provided CSR. The validity of the CSR is checked and passed to the sectigo server.",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "SSL"
                ],
                "summary": "SSL CSR Endpoint",
                "parameters": [
                    {
                        "description": "The CSR",
                        "name": "csr",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.CsrRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "certificate",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "default": {
                        "description": "Error processing the request",
                        "schema": {
                            "$ref": "#/definitions/echo.HTTPError"
                        }
                    }
                }
            }
        },
        "/ssl/revoke": {
            "post": {
                "security": [
                    {
                        "API": []
                    }
                ],
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "SSL"
                ],
                "summary": "SSL Revoke Endpoint",
                "parameters": [
                    {
                        "description": "The serial of the certificate to revoke and the reason",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.RevokeRequest"
                        }
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    },
                    "default": {
                        "description": "Error processing the request",
                        "schema": {
                            "$ref": "#/definitions/echo.HTTPError"
                        }
                    }
                }
            }
        },
        "/whoami": {
            "get": {
                "security": [
                    {
                        "API": []
                    }
                ],
                "description": "Returns the username of the logged in user",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User"
                ],
                "summary": "whoami Endpoint",
                "responses": {
                    "200": {
                        "description": "Username",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/echo.HTTPError"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "echo.HTTPError": {
            "type": "object",
            "properties": {
                "message": {}
            }
        },
        "model.CsrRequest": {
            "type": "object",
            "required": [
                "csr"
            ],
            "properties": {
                "csr": {
                    "type": "string"
                }
            }
        },
        "model.RevokeRequest": {
            "type": "object",
            "required": [
                "reason",
                "serial"
            ],
            "properties": {
                "reason": {
                    "type": "string"
                },
                "serial": {
                    "type": "string"
                }
            }
        },
        "portal_apis.ListSmimeResponse_CertificateDetails": {
            "type": "object",
            "properties": {
                "expires": {
                    "$ref": "#/definitions/timestamppb.Timestamp"
                },
                "id": {
                    "type": "integer"
                },
                "serial": {
                    "type": "string"
                },
                "status": {
                    "type": "string"
                }
            }
        },
        "portal_apis.SslCertificateDetails": {
            "type": "object",
            "properties": {
                "ca": {
                    "type": "string"
                },
                "common_name": {
                    "type": "string"
                },
                "created": {
                    "$ref": "#/definitions/timestamppb.Timestamp"
                },
                "db_id": {
                    "type": "integer"
                },
                "expires": {
                    "$ref": "#/definitions/timestamppb.Timestamp"
                },
                "id": {
                    "type": "integer"
                },
                "issued_by": {
                    "type": "string"
                },
                "not_before": {
                    "$ref": "#/definitions/timestamppb.Timestamp"
                },
                "serial": {
                    "type": "string"
                },
                "source": {
                    "type": "string"
                },
                "status": {
                    "type": "string"
                },
                "subject_alternative_names": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "timestamppb.Timestamp": {
            "type": "object",
            "properties": {
                "nanos": {
                    "description": "Non-negative fractions of a second at nanosecond resolution. Negative\nsecond values with fractions must still have non-negative nanos values\nthat count forward in time. Must be from 0 to 999,999,999\ninclusive.",
                    "type": "integer"
                },
                "seconds": {
                    "description": "Represents seconds of UTC time since Unix epoch\n1970-01-01T00:00:00Z. Must be from 0001-01-01T00:00:00Z to\n9999-12-31T23:59:59Z inclusive.",
                    "type": "integer"
                }
            }
        }
    },
    "securityDefinitions": {
        "API": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "2.0",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "PKI Service",
	Description:      "Go microservice for PKI management. Provides an wrapper above the sectigo API.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
