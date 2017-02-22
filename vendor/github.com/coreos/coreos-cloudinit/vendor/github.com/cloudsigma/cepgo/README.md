cepgo
=====

Cepko implements easy-to-use communication with CloudSigma's VMs through a
virtual serial port without bothering with formatting the messages properly nor
parsing the output with the specific and sometimes confusing shell tools for
that purpose.

Having the server definition accessible by the VM can be useful in various
ways. For example it is possible to easily determine from within the VM, which
network interfaces are connected to public and which to private network.
Another use is to pass some data to initial VM setup scripts, like setting the
hostname to the VM name or passing ssh public keys through server meta.

Example usage:

    package main

    import (
            "fmt"

            "github.com/cloudsigma/cepgo"
    )

    func main() {
            c := cepgo.NewCepgo()
            result, err := c.Meta()
            if err != nil {
                    panic(err)
            }
            fmt.Printf("%#v", result)
    }

Output:

    map[string]interface {}{
        "optimize_for":"custom",
        "ssh_public_key":"ssh-rsa AAA...",
        "description":"[...]",
    }

For more information take a look at the Server Context section of CloudSigma
API Docs: http://cloudsigma-docs.readthedocs.org/en/latest/server_context.html
