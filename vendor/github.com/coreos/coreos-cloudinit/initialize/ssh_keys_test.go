// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package initialize

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCloudConfigUsersUrlMarshal(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gh_res := `
[
  {
    "key": "ssh-dss AAAAB3NzaC1kc3MAAACBAIHAu822ggSkIHrJYvhmBceOSVjuflfQm8RbMMDNVe9relQfuPbN+nxGGTCKzPLebeOcX+Wwi77TPXWwK3BZMglfXxhABlFPsuMb63Tqp94pBYsJdx/iFj9iGo6pKoM1k8ubOcqsUnq+BR9895zRbE7MjdwkGo67+QhCEwvkwAnNAAAAFQCuddVqXLCubzqnWmeHLQE+2GFfHwAAAIBnlXW5h15ndVuwi0htF4oodVSB1KwnTWcuBK+aE1zRs76yvRb0Ws+oifumThDwB/Tec6FQuAfRKfy6piChZqsu5KvL98I+2t5yyi1td+kMvdTnVL2lW44etDKseOcozmknCOmh4Dqvhl/2MwrDAhlPaN08EEq9h3w3mXtNLWH64QAAAIBAzDOKr17llngaKIdDXh+LtXKh87+zfjlTA36/9r2uF2kYE5uApDtu9sPCkt7+YBQt7R8prADPckwAiXwVdk0xijIOpLDBmoydQJJRQ+zTMxvpQmUr/1kUOv0zb+lB657CgvN0vVTmP2swPeMvgntt3C4vw7Ab+O+MS9peOAJbbQ=="
  },
  {
    "key": "ssh-dss AAAAB3NzaC1kc3MAAACBANxpzIbTzKTeBRaOIdUxwwGwvDasTfU/PonhbNIuhYjc+xFGvBRTumox2F+luVAKKs4WdvA4nJXaY1OFi6DZftk5Bp4E2JaSzp8ulAzHsMexDdv6LGHGEJj/qdHAL1vHk2K89PpwRFSRZI8XRBLjvkr4ZgBKLG5ZILXPJEPP2j3lAAAAFQCtxoTnV8wy0c4grcGrQ+1sCsD7WQAAAIAqZsW2GviMe1RQrbZT0xAZmI64XRPrnLsoLxycHWlS7r6uUln2c6Ae2MB/YF0d4Kd1XZii9GHj7rrypqEo7MW8uSabhu70nmu1J8m2O3Dsr+4oJLeat9vwPsJV92IKO0jQwjKnAOHOiB9JKGeCw+NfXfogbti9/q38Q6XcS+SI5wAAAIEA1803Y2h+tOOpZXAsNIwl9mRfExWzLQ3L7knwJdznQu/6SW1H/1oyoYLebuk187Qj2UFI5qQ6AZNc49DvohWx0Cg6ABcyubNyoaCjZKWIdxVnItHWNbLe//+tyTu0I2eQwJOORsEPK5gMpf599C7wXQ//DzZOWbTWiHEX52gCTmk="
  },
  {
    "id": 5224438,
    "key": "ssh-dss AAAAB3NzaC1kc3MAAACBAPKRWdKhzGZuLAJL6M1eM51hWViMqNBC2C6lm2OqGRYLuIf1GJ391widUuSf4wQqnkR22Q9PCmAZ19XCf11wBRMnuw9I/Z3Bt5bXfc+dzFBCmHYGJ6wNSv++H9jxyMb+usmsenWOFZGNO2jN0wrJ4ay8Yt0bwtRU+VCXpuRLszMzAAAAFQDZUIuPjcfK5HLgnwZ/J3lvtvlUjQAAAIEApIkAwLuCQV5j3U6DmI/Y6oELqSUR2purFm8jo8jePFfe1t+ghikgD254/JXlhDCVgY0NLXcak+coJfGCTT23quJ7I5xdpTn/OZO2Q6Woum/bijFC/UWwQbLz0R2nU3DoHv5v6XHQZxuIG4Fsxa91S+vWjZFtI7RuYlBCZA//ANMAAACBAJO0FojzkX6IeaWLqrgu9GTkFwGFazZ+LPH5JOWPoPn1hQKuR32Uf6qNcBZcIjY7SF0P7HF5rLQd6zKZzHqqQQ92MV555NEwjsnJglYU8CaaZsfYooaGPgA1YN7RhTSAuDmUW5Hyfj5BH4NTtrzrvJxIhDoQLf31Fasjw00r4R0O"
  }
]
`
		fmt.Fprintln(w, gh_res)
	}))
	defer ts.Close()

	keys, err := fetchUserKeys(ts.URL)
	if err != nil {
		t.Fatalf("Encountered unexpected error: %v", err)
	}
	expected := "ssh-dss AAAAB3NzaC1kc3MAAACBAIHAu822ggSkIHrJYvhmBceOSVjuflfQm8RbMMDNVe9relQfuPbN+nxGGTCKzPLebeOcX+Wwi77TPXWwK3BZMglfXxhABlFPsuMb63Tqp94pBYsJdx/iFj9iGo6pKoM1k8ubOcqsUnq+BR9895zRbE7MjdwkGo67+QhCEwvkwAnNAAAAFQCuddVqXLCubzqnWmeHLQE+2GFfHwAAAIBnlXW5h15ndVuwi0htF4oodVSB1KwnTWcuBK+aE1zRs76yvRb0Ws+oifumThDwB/Tec6FQuAfRKfy6piChZqsu5KvL98I+2t5yyi1td+kMvdTnVL2lW44etDKseOcozmknCOmh4Dqvhl/2MwrDAhlPaN08EEq9h3w3mXtNLWH64QAAAIBAzDOKr17llngaKIdDXh+LtXKh87+zfjlTA36/9r2uF2kYE5uApDtu9sPCkt7+YBQt7R8prADPckwAiXwVdk0xijIOpLDBmoydQJJRQ+zTMxvpQmUr/1kUOv0zb+lB657CgvN0vVTmP2swPeMvgntt3C4vw7Ab+O+MS9peOAJbbQ=="
	if keys[0] != expected {
		t.Fatalf("expected %s, got %s", expected, keys[0])
	}
	expected = "ssh-dss AAAAB3NzaC1kc3MAAACBAPKRWdKhzGZuLAJL6M1eM51hWViMqNBC2C6lm2OqGRYLuIf1GJ391widUuSf4wQqnkR22Q9PCmAZ19XCf11wBRMnuw9I/Z3Bt5bXfc+dzFBCmHYGJ6wNSv++H9jxyMb+usmsenWOFZGNO2jN0wrJ4ay8Yt0bwtRU+VCXpuRLszMzAAAAFQDZUIuPjcfK5HLgnwZ/J3lvtvlUjQAAAIEApIkAwLuCQV5j3U6DmI/Y6oELqSUR2purFm8jo8jePFfe1t+ghikgD254/JXlhDCVgY0NLXcak+coJfGCTT23quJ7I5xdpTn/OZO2Q6Woum/bijFC/UWwQbLz0R2nU3DoHv5v6XHQZxuIG4Fsxa91S+vWjZFtI7RuYlBCZA//ANMAAACBAJO0FojzkX6IeaWLqrgu9GTkFwGFazZ+LPH5JOWPoPn1hQKuR32Uf6qNcBZcIjY7SF0P7HF5rLQd6zKZzHqqQQ92MV555NEwjsnJglYU8CaaZsfYooaGPgA1YN7RhTSAuDmUW5Hyfj5BH4NTtrzrvJxIhDoQLf31Fasjw00r4R0O"
	if keys[2] != expected {
		t.Fatalf("expected %s, got %s", expected, keys[2])
	}
}
