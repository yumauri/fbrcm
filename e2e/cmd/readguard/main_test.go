package main

import (
	"strings"
	"testing"
)

func TestValidateAllowsReadsToConfiguredHost(t *testing.T) {
	method := "GET"
	destination := "firebase.example:443"
	path := "/v1/projects/demo/remoteConfig"
	value := payload{}
	value.Request.Method = &method
	value.Request.Destination = &destination
	value.Request.Path = &path
	if err := validate(value, []allowedRequest{{Method: "GET", Host: "firebase.example", Path: path}}); err != nil {
		t.Fatal(err)
	}
}

func TestValidateBlocksUnexpectedMethod(t *testing.T) {
	method := "PUT"
	destination := "firebase.example"
	path := "/v1/projects/demo/remoteConfig"
	value := payload{}
	value.Request.Method = &method
	value.Request.Destination = &destination
	value.Request.Path = &path
	if err := validate(value, []allowedRequest{{Method: "GET", Host: "firebase.example", Path: path}}); err == nil || !strings.Contains(err.Error(), "unexpected request") {
		t.Fatalf("validate() error = %v", err)
	}
}

func TestValidateBlocksOtherHost(t *testing.T) {
	method := "GET"
	destination := "other.example"
	path := "/v1/projects/demo/remoteConfig"
	value := payload{}
	value.Request.Method = &method
	value.Request.Destination = &destination
	value.Request.Path = &path
	if err := validate(value, []allowedRequest{{Method: "GET", Host: "firebase.example", Path: path}}); err == nil || !strings.Contains(err.Error(), "unexpected request") {
		t.Fatalf("validate() error = %v", err)
	}
}

func TestValidateAllowsExplicitMethodAcrossHosts(t *testing.T) {
	method := "PUT"
	destination := "second.example"
	path := "/v1/projects/demo/remoteConfig"
	value := payload{}
	value.Request.Method = &method
	value.Request.Destination = &destination
	value.Request.Path = &path
	allowed := []allowedRequest{
		{Method: "GET", Host: "first.example", Path: path},
		{Method: "PUT", Host: "second.example", Path: path},
	}
	if err := validate(value, allowed); err != nil {
		t.Fatal(err)
	}
}

func TestValidateRequiresExactDeclaredQuery(t *testing.T) {
	method := "PUT"
	destination := "firebase.example:443"
	path := "/v1/projects/demo/remoteConfig"
	value := payload{}
	value.Request.Method = &method
	value.Request.Destination = &destination
	value.Request.Path = &path
	allowed := []allowedRequest{{Method: method, Host: "firebase.example", Path: path, Query: "validateOnly=true"}}
	if err := validate(value, allowed); err == nil {
		t.Fatal("validate accepted a PUT without validateOnly=true")
	}
	query := "validateOnly=true"
	value.Request.Query = &query
	if err := validate(value, allowed); err != nil {
		t.Fatal(err)
	}
}

func TestDecodeAllowedRequests(t *testing.T) {
	allowed, err := decodeAllowedRequests(`[{"method":"get","host":" firebase.example ","path":" /v1/demo "}]`)
	if err != nil {
		t.Fatal(err)
	}
	if len(allowed) != 1 || allowed[0].Method != "GET" || allowed[0].Host != "firebase.example" || allowed[0].Path != "/v1/demo" {
		t.Fatalf("allowed requests = %#v", allowed)
	}
}
