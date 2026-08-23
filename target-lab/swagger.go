package main

import "net/http"

const swaggerJSON = `{
  "openapi": "3.0.3",
  "info": {"title": "GoAPIScanner Target Lab", "version": "1.0.0"},
  "servers": [{"url": "/"}],
  "paths": {
    "/api/users": {
      "get": {
        "operationId": "listUsers",
        "parameters": [
          {"name": "id", "in": "query", "schema": {"type": "string"}}
        ]
      }
    },
    "/api/users/blind": {
      "get": {
        "operationId": "blindUsers",
        "parameters": [
          {"name": "id", "in": "query", "schema": {"type": "string"}}
        ]
      }
    },
    "/api/search": {
      "get": {
        "operationId": "search",
        "parameters": [
          {"name": "q", "in": "query", "schema": {"type": "string"}}
        ]
      }
    },
    "/api/admin/secret": {
      "get": {
        "operationId": "adminSecret",
        "parameters": [
          {"name": "Authorization", "in": "header", "schema": {"type": "string"}}
        ]
      }
    },
    "/api/files": {
      "get": {
        "operationId": "readFile",
        "parameters": [
          {"name": "path", "in": "query", "schema": {"type": "string"}}
        ]
      }
    },
    "/api/ping": {
      "get": {
        "operationId": "pingHost",
        "parameters": [
          {"name": "host", "in": "query", "schema": {"type": "string"}}
        ]
      }
    }
  }
}`

func serveSwagger(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(swaggerJSON))
}
