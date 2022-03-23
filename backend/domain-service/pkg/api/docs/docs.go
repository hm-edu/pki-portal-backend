// Package docs GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag
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
            "url": "https://github.com/hm-edu/pki-service/blob/main/LICENSE"
        },
        "license": {
            "name": "Apache License",
            "url": "https://github.com/hm-edu/pki-service/blob/main/LICENSE"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/healthz": {
            "get": {
                "description": "Used by Kubernetes Liveness Probe",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Kubernetes"
                ],
                "summary": "Liveness check",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api.healthResponse"
                        }
                    },
                    "503": {
                        "description": "Service Unavailable",
                        "schema": {
                            "$ref": "#/definitions/api.healthResponse"
                        }
                    }
                }
            }
        },
        "/readyz": {
            "get": {
                "description": "Used by Kubernetes Readiness Probe",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Kubernetes"
                ],
                "summary": "Readiness check",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/api.healthResponse"
                        }
                    },
                    "503": {
                        "description": "Service Unavailable",
                        "schema": {
                            "$ref": "#/definitions/api.healthResponse"
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
                    "401": {
                        "description": "Forbidden",
                        "schema": {
                            "$ref": "#/definitions/models.Error"
                        }
                    },
                    "403": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/models.Error"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "api.healthResponse": {
            "type": "object",
            "properties": {
                "status": {
                    "type": "string"
                }
            }
        },
        "models.Error": {
            "type": "object",
            "properties": {
                "details": {
                    "type": "string"
                },
                "message": {
                    "type": "string"
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
	Title:            "Domain Service",
	Description:      "Go microservice for Domain management.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
