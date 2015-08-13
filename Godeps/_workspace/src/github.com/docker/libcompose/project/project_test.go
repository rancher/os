package project

import (
	"fmt"
	"testing"
)

func TestEventEquality(t *testing.T) {
	if fmt.Sprintf("%s", SERVICE_START) != "Started" ||
		fmt.Sprintf("%v", SERVICE_START) != "Started" {
		t.Fatalf("SERVICE_START String() doesn't work: %s %v", SERVICE_START, SERVICE_START)
	}

	if fmt.Sprintf("%s", SERVICE_START) != fmt.Sprintf("%s", SERVICE_UP) {
		t.Fatal("Event messages do not match")
	}

	if SERVICE_START == SERVICE_UP {
		t.Fatal("Events match")
	}
}
