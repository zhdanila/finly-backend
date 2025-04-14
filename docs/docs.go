// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/auth/login": {
            "post": {
                "description": "Authenticates a user with the provided credentials",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User"
                ],
                "summary": "Login a user",
                "operationId": "login-user",
                "parameters": [
                    {
                        "description": "User Credentials",
                        "name": "credentials",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/finly-backend_internal_service_auth.LoginRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/finly-backend_internal_service_auth.LoginResponse"
                        }
                    }
                }
            }
        },
        "/auth/logout": {
            "post": {
                "description": "Logs out a user by invalidating their authentication token",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User"
                ],
                "summary": "Logout a user",
                "operationId": "logout-user",
                "parameters": [
                    {
                        "description": "Authentication Token",
                        "name": "token",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/finly-backend_internal_service_auth.LogoutRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/finly-backend_internal_service_auth.LogoutResponse"
                        }
                    }
                }
            }
        },
        "/auth/me": {
            "post": {
                "description": "Retrieves information about the currently authenticated user",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User"
                ],
                "summary": "Get user information",
                "operationId": "get-user-info",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Authentication Token",
                        "name": "token",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/finly-backend_internal_service_auth.MeResponse"
                        }
                    }
                }
            }
        },
        "/auth/refresh": {
            "post": {
                "description": "Refreshes a user's authentication token by generating a new one",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User"
                ],
                "summary": "Refresh a user token",
                "operationId": "refresh-token",
                "parameters": [
                    {
                        "description": "Authentication Token",
                        "name": "token",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/finly-backend_internal_service_auth.RefreshTokenRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/finly-backend_internal_service_auth.RefreshTokenResponse"
                        }
                    }
                }
            }
        },
        "/auth/register": {
            "post": {
                "description": "Registers a new user with the provided details",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "User"
                ],
                "summary": "RegisterUser a new user",
                "operationId": "register-user",
                "parameters": [
                    {
                        "description": "User Details",
                        "name": "user",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/finly-backend_internal_service_auth.RegisterRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/finly-backend_internal_service_auth.RegisterResponse"
                        }
                    }
                }
            }
        },
        "/budget": {
            "post": {
                "description": "Creates a new budget for the user with the provided details",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Budget"
                ],
                "summary": "Create a new budget",
                "operationId": "create-budget",
                "parameters": [
                    {
                        "description": "Budget Details",
                        "name": "budget",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/finly-backend_internal_service_budget.CreateBudgetRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/finly-backend_internal_service_budget.CreateBudgetResponse"
                        }
                    }
                }
            }
        },
        "/budget/{budget_id}": {
            "get": {
                "description": "Retrieves a budget by its ID for the specified user",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Budget"
                ],
                "summary": "Get budget by ID",
                "operationId": "get-budget-by-id",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Budget ID",
                        "name": "budget_id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "User ID",
                        "name": "user_id",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/finly-backend_internal_service_budget.GetBudgetByIDResponse"
                        }
                    }
                }
            }
        },
        "/budget/{budget_id}/history": {
            "get": {
                "description": "Retrieves the history of a budget for the specified user",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Budget"
                ],
                "summary": "Get budget history",
                "operationId": "get-budget-history",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Budget ID",
                        "name": "budget_id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "User ID",
                        "name": "user_id",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/finly-backend_internal_service_budget.GetBudgetHistoryResponse"
                        }
                    }
                }
            }
        },
        "/category": {
            "get": {
                "description": "Retrieves all categories for the user",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Category"
                ],
                "summary": "List all categories",
                "operationId": "list-categories",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/finly-backend_internal_service_category.GetCategoryByIDResponse"
                            }
                        }
                    }
                }
            },
            "post": {
                "description": "Creates a new category for the user with the provided details",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Category"
                ],
                "summary": "Create a new category",
                "operationId": "create-category",
                "parameters": [
                    {
                        "description": "Category Details",
                        "name": "category",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/finly-backend_internal_service_category.CreateCategoryRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/finly-backend_internal_service_category.CreateCategoryResponse"
                        }
                    }
                }
            }
        },
        "/category/{id}": {
            "get": {
                "description": "Retrieves the category details for the given ID",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Category"
                ],
                "summary": "Get category by ID",
                "operationId": "get-category-by-id",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Category ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/finly-backend_internal_service_category.GetCategoryByIDResponse"
                        }
                    }
                }
            },
            "delete": {
                "description": "Deletes the category with the given ID",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Category"
                ],
                "summary": "Delete a category",
                "operationId": "delete-category",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Category ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/finly-backend_internal_service_category.DeleteCategoryResponse"
                        }
                    }
                }
            }
        },
        "/transaction": {
            "post": {
                "description": "Creates a new transaction for the user with the provided details",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Transaction"
                ],
                "summary": "Create a new transaction",
                "operationId": "create-transaction",
                "parameters": [
                    {
                        "description": "Transaction Details",
                        "name": "transaction",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/finly-backend_internal_service_transaction.CreateTransactionRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/finly-backend_internal_service_transaction.CreateTransactionResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "finly-backend_internal_domain_enums_e_transaction_type.Enum": {
            "type": "string",
            "enum": [
                "deposit",
                "withdrawal",
                "initial"
            ],
            "x-enum-varnames": [
                "Deposit",
                "Withdrawal",
                "Initial"
            ]
        },
        "finly-backend_internal_service_auth.LoginRequest": {
            "type": "object",
            "required": [
                "email",
                "password"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "password": {
                    "type": "string",
                    "maxLength": 100,
                    "minLength": 5
                }
            }
        },
        "finly-backend_internal_service_auth.LoginResponse": {
            "type": "object",
            "properties": {
                "token": {
                    "type": "string"
                }
            }
        },
        "finly-backend_internal_service_auth.LogoutRequest": {
            "type": "object",
            "properties": {
                "authToken": {
                    "type": "string"
                }
            }
        },
        "finly-backend_internal_service_auth.LogoutResponse": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string"
                }
            }
        },
        "finly-backend_internal_service_auth.MeResponse": {
            "type": "object",
            "required": [
                "email",
                "first_name",
                "last_name"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "first_name": {
                    "type": "string",
                    "maxLength": 100,
                    "minLength": 1
                },
                "last_name": {
                    "type": "string",
                    "maxLength": 100,
                    "minLength": 1
                }
            }
        },
        "finly-backend_internal_service_auth.RefreshTokenRequest": {
            "type": "object",
            "properties": {
                "authToken": {
                    "type": "string"
                }
            }
        },
        "finly-backend_internal_service_auth.RefreshTokenResponse": {
            "type": "object",
            "properties": {
                "token": {
                    "type": "string"
                }
            }
        },
        "finly-backend_internal_service_auth.RegisterRequest": {
            "type": "object",
            "required": [
                "email",
                "first_name",
                "last_name",
                "password"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "first_name": {
                    "type": "string",
                    "maxLength": 100,
                    "minLength": 1
                },
                "last_name": {
                    "type": "string",
                    "maxLength": 100,
                    "minLength": 1
                },
                "password": {
                    "type": "string",
                    "maxLength": 100,
                    "minLength": 8
                }
            }
        },
        "finly-backend_internal_service_auth.RegisterResponse": {
            "type": "object",
            "properties": {
                "token": {
                    "type": "string"
                }
            }
        },
        "finly-backend_internal_service_budget.BudgetHistory": {
            "type": "object",
            "properties": {
                "balance": {
                    "type": "number"
                },
                "budget_id": {
                    "type": "string"
                },
                "created_at": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                }
            }
        },
        "finly-backend_internal_service_budget.CreateBudgetRequest": {
            "type": "object",
            "required": [
                "amount",
                "currency",
                "userID"
            ],
            "properties": {
                "amount": {
                    "type": "number"
                },
                "currency": {
                    "type": "string"
                },
                "userID": {
                    "type": "string"
                }
            }
        },
        "finly-backend_internal_service_budget.CreateBudgetResponse": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string"
                }
            }
        },
        "finly-backend_internal_service_budget.GetBudgetByIDResponse": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "currency": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                },
                "user_id": {
                    "type": "string"
                }
            }
        },
        "finly-backend_internal_service_budget.GetBudgetHistoryResponse": {
            "type": "object",
            "properties": {
                "budget_history": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/finly-backend_internal_service_budget.BudgetHistory"
                    }
                }
            }
        },
        "finly-backend_internal_service_category.CreateCategoryRequest": {
            "type": "object",
            "required": [
                "description",
                "name",
                "userID"
            ],
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "is_user_category": {
                    "type": "boolean"
                },
                "name": {
                    "type": "string"
                },
                "userID": {
                    "type": "string"
                },
                "user_id": {
                    "type": "string"
                }
            }
        },
        "finly-backend_internal_service_category.CreateCategoryResponse": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string"
                }
            }
        },
        "finly-backend_internal_service_category.DeleteCategoryResponse": {
            "type": "object"
        },
        "finly-backend_internal_service_category.GetCategoryByIDResponse": {
            "type": "object",
            "required": [
                "description",
                "name"
            ],
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "is_user_category": {
                    "type": "boolean"
                },
                "name": {
                    "type": "string"
                },
                "user_id": {
                    "type": "string"
                }
            }
        },
        "finly-backend_internal_service_transaction.CreateTransactionRequest": {
            "type": "object",
            "required": [
                "amount",
                "budget_id",
                "category_id",
                "type",
                "userID"
            ],
            "properties": {
                "amount": {
                    "type": "number"
                },
                "budget_id": {
                    "type": "string"
                },
                "category_id": {
                    "type": "string"
                },
                "note": {
                    "type": "string"
                },
                "type": {
                    "enum": [
                        "deposit",
                        "withdrawal"
                    ],
                    "allOf": [
                        {
                            "$ref": "#/definitions/finly-backend_internal_domain_enums_e_transaction_type.Enum"
                        }
                    ]
                },
                "userID": {
                    "type": "string"
                }
            }
        },
        "finly-backend_internal_service_transaction.CreateTransactionResponse": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
