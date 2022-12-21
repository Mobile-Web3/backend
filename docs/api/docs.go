// Package api GENERATED BY SWAG; DO NOT EDIT
// This file was generated by swaggo/swag
package api

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
        "/balance/check": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "balance"
                ],
                "summary": "Получить инфу о балансе",
                "parameters": [
                    {
                        "description": "body",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/api.checkRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/api.Response"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "result": {
                                            "$ref": "#/definitions/chain.CheckResponse"
                                        }
                                    }
                                }
                            ]
                        }
                    }
                }
            }
        },
        "/chains/all": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "chain"
                ],
                "summary": "Получение данных о сетях",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/api.Response"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "result": {
                                            "type": "array",
                                            "items": {
                                                "$ref": "#/definitions/chain.ShortResponse"
                                            }
                                        }
                                    }
                                }
                            ]
                        }
                    }
                }
            }
        },
        "/transaction/send": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "transaction"
                ],
                "summary": "Отправить транзакцию",
                "parameters": [
                    {
                        "description": "body",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/chain.SendTxInput"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/api.Response"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "result": {
                                            "$ref": "#/definitions/chain.SendTxResponse"
                                        }
                                    }
                                }
                            ]
                        }
                    }
                }
            }
        },
        "/transaction/simulate": {
            "post": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "transaction"
                ],
                "summary": "Симуляция транзакции для расчета параметров",
                "parameters": [
                    {
                        "description": "body",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/chain.SimulateTxInput"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/api.Response"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "result": {
                                            "$ref": "#/definitions/chain.SimulateTxResponse"
                                        }
                                    }
                                }
                            ]
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "api.Response": {
            "type": "object",
            "properties": {
                "error": {
                    "type": "string"
                },
                "isSuccess": {
                    "type": "boolean"
                },
                "result": {}
            }
        },
        "api.checkRequest": {
            "type": "object",
            "properties": {
                "walletAddress": {
                    "type": "string"
                }
            }
        },
        "chain.CheckResponse": {
            "type": "object",
            "properties": {
                "availableAmount": {
                    "type": "string"
                },
                "stakedAmount": {
                    "type": "string"
                },
                "totalAmount": {
                    "type": "string"
                }
            }
        },
        "chain.SendTxInput": {
            "type": "object",
            "properties": {
                "amount": {
                    "type": "string"
                },
                "from": {
                    "type": "string"
                },
                "gasAdjusted": {
                    "type": "string"
                },
                "gasPrice": {
                    "type": "string"
                },
                "memo": {
                    "type": "string"
                },
                "mnemonic": {
                    "type": "string"
                },
                "to": {
                    "type": "string"
                }
            }
        },
        "chain.SendTxResponse": {
            "type": "object",
            "properties": {
                "data": {
                    "type": "string"
                },
                "gasUsed": {
                    "type": "integer"
                },
                "gasWanted": {
                    "type": "integer"
                },
                "height": {
                    "type": "integer"
                },
                "rawLog": {
                    "type": "string"
                },
                "txHash": {
                    "type": "string"
                }
            }
        },
        "chain.ShortResponse": {
            "type": "object",
            "properties": {
                "base": {
                    "type": "string"
                },
                "bech32Prefix": {
                    "type": "string"
                },
                "chainId": {
                    "type": "string"
                },
                "chainName": {
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "display": {
                    "type": "string"
                },
                "keyAlgos": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "logoPngUrl": {
                    "type": "string"
                },
                "logoSvgUrl": {
                    "type": "string"
                },
                "prettyName": {
                    "type": "string"
                },
                "slip44": {
                    "type": "integer"
                },
                "symbol": {
                    "type": "string"
                }
            }
        },
        "chain.SimulateTxInput": {
            "type": "object",
            "properties": {
                "amount": {
                    "type": "string"
                },
                "from": {
                    "type": "string"
                },
                "memo": {
                    "type": "string"
                },
                "mnemonic": {
                    "type": "string"
                },
                "to": {
                    "type": "string"
                }
            }
        },
        "chain.SimulateTxResponse": {
            "type": "object",
            "properties": {
                "averageGasPrice": {
                    "type": "string"
                },
                "gasAdjusted": {
                    "type": "string"
                },
                "highGasPrice": {
                    "type": "string"
                },
                "lowGasPrice": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "",
	BasePath:         "/api",
	Schemes:          []string{},
	Title:            "Swagger UI",
	Description:      "API",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
