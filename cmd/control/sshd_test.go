package control

import (
	"bufio"
	"strings"
	"testing"
)

func TestCountEstablished(t *testing.T) {
	tcp := `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 00000000:0016 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 1 1 0000000000000000 100 0 0 10 0
   1: 0F02000A:0016 0202000A:D2F4 01 00000000:00000000 02:000A7E1C 00000000     0        0 2 4 0000000000000000 20 4 30 10 -1
   2: 0F02000A:8E2C 0202000A:0016 01 00000000:00000000 02:000A7E1C 00000000     0        0 3 4 0000000000000000 20 4 30 10 -1
   3: 0F02000A:08AE 0202000A:D2F5 01 00000000:00000000 02:000A7E1C 00000000     0        0 4 4 0000000000000000 20 4 30 10 -1
`
	// Only the inbound connection to port 22 counts: 0A is LISTEN and the
	// third line is an outgoing connection to port 22 of another host.
	if n, err := countEstablished(bufio.NewScanner(strings.NewReader(tcp)), 22); err != nil || n != 1 {
		t.Fatalf("expected 1, got %d %v", n, err)
	}
	if n, _ := countEstablished(bufio.NewScanner(strings.NewReader(tcp)), 2222); n != 1 {
		t.Fatalf("expected 1 on port 2222, got %d", n)
	}
}
