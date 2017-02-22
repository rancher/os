// Cepko implements easy-to-use communication with CloudSigma's VMs through a
// virtual serial port without bothering with formatting the messages properly
// nor parsing the output with the specific and sometimes confusing shell tools
// for that purpose.
//
// Having the server definition accessible by the VM can be useful in various
// ways. For example it is possible to easily determine from within the VM,
// which network interfaces are connected to public and which to private
// network. Another use is to pass some data to initial VM setup scripts, like
// setting the hostname to the VM name or passing ssh public keys through
// server meta.
//
// Example usage:
//
//   package main
//
//   import (
//           "fmt"
//
//           "github.com/cloudsigma/cepgo"
//   )
//
//   func main() {
//           c := cepgo.NewCepgo()
//           result, err := c.Meta()
//           if err != nil {
//                   panic(err)
//           }
//           fmt.Printf("%#v", result)
//   }
//
// Output:
//
//   map[string]string{
//   	"optimize_for":"custom",
//   	"ssh_public_key":"ssh-rsa AAA...",
//   	"description":"[...]",
//   }
//
// For more information take a look at the Server Context section API Docs:
// http://cloudsigma-docs.readthedocs.org/en/latest/server_context.html
package cepgo

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"

	"github.com/tarm/goserial"
)

const (
	requestPattern = "<\n%s\n>"
	EOT            = '\x04' // End Of Transmission
)

var (
	SerialPort string = "/dev/ttyS1"
	Baud       int    = 115200
)

// Sets the serial port. If the operating system is windows CloudSigma's server
// context is at COM2 port, otherwise (linux, freebsd, darwin) the port is
// being left to the default /dev/ttyS1.
func init() {
	if runtime.GOOS == "windows" {
		SerialPort = "COM2"
	}
}

// The default fetcher makes the connection to the serial port,
// writes given query and reads until the EOT symbol.
func fetchViaSerialPort(key string) ([]byte, error) {
	config := &serial.Config{Name: SerialPort, Baud: Baud}
	connection, err := serial.OpenPort(config)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(requestPattern, key)
	if _, err := connection.Write([]byte(query)); err != nil {
		return nil, err
	}

	reader := bufio.NewReader(connection)
	answer, err := reader.ReadBytes(EOT)
	if err != nil {
		return nil, err
	}

	return answer[0 : len(answer)-1], nil
}

// Queries to the serial port can be executed only from instance of this type.
// The result from each of them can be either interface{}, map[string]string or
// a single in case of single value is returned. There is also a public metod
// who directly calls the fetcher and returns raw []byte from the serial port.
type Cepgo struct {
	fetcher func(string) ([]byte, error)
}

// Creates a Cepgo instance with the default serial port fetcher.
func NewCepgo() *Cepgo {
	cepgo := new(Cepgo)
	cepgo.fetcher = fetchViaSerialPort
	return cepgo
}

// Creates a Cepgo instance with custom fetcher.
func NewCepgoFetcher(fetcher func(string) ([]byte, error)) *Cepgo {
	cepgo := new(Cepgo)
	cepgo.fetcher = fetcher
	return cepgo
}

// Fetches raw []byte from the serial port using directly the fetcher member.
func (c *Cepgo) FetchRaw(key string) ([]byte, error) {
	return c.fetcher(key)
}

// Fetches a single key and tries to unmarshal the result to json and returns
// it. If the unmarshalling fails it's safe to assume the result it's just a
// string and returns it.
func (c *Cepgo) Key(key string) (interface{}, error) {
	var result interface{}

	fetched, err := c.FetchRaw(key)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(fetched, &result)
	if err != nil {
		return string(fetched), nil
	}
	return result, nil
}

// Fetches all the server context. Equivalent of c.Key("")
func (c *Cepgo) All() (interface{}, error) {
	return c.Key("")
}

// Fetches only the object meta field and makes sure to return a proper
// map[string]string
func (c *Cepgo) Meta() (map[string]string, error) {
	rawMeta, err := c.Key("/meta/")
	if err != nil {
		return nil, err
	}

	return typeAssertToMapOfStrings(rawMeta)
}

// Fetches only the global context and makes sure to return a proper
// map[string]string
func (c *Cepgo) GlobalContext() (map[string]string, error) {
	rawContext, err := c.Key("/global_context/")
	if err != nil {
		return nil, err
	}

	return typeAssertToMapOfStrings(rawContext)
}

// Just a little helper function that uses type assertions in order to convert
// a interface{} to map[string]string if this is possible.
func typeAssertToMapOfStrings(raw interface{}) (map[string]string, error) {
	result := make(map[string]string)

	dictionary, ok := raw.(map[string]interface{})
	if !ok {
		return nil, errors.New("Received bytes are formatted badly")
	}

	for key, rawValue := range dictionary {
		if value, ok := rawValue.(string); ok {
			result[key] = value
		} else {
			return nil, errors.New("Server context metadata is formatted badly")
		}
	}
	return result, nil
}
