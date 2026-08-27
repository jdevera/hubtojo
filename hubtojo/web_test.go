package main

import (
	"bytes"
	"net"
	"strings"
	"testing"
)

func TestStatsPageRenders(t *testing.T) {
	var output bytes.Buffer
	store := NewStatsStore("test", 3600)

	err := statsPage.ExecuteTemplate(&output, "stats.html", store.Snapshot())
	if err != nil {
		t.Fatalf("render stats page: %v", err)
	}
	if !strings.Contains(output.String(), "HubToJo") {
		t.Fatal("rendered stats page does not contain the application name")
	}
}

func TestStartWebServerReturnsBindError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve test address: %v", err)
	}
	defer listener.Close()

	server, err := StartWebServer(listener.Addr().String(), NewStatsStore("test", 3600))
	if err == nil {
		server.Close()
		t.Fatal("expected an error when the web address is already in use")
	}
}
