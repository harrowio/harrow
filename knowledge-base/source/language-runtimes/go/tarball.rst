Using Go (golang) in Harrow from a tarball
==========================================

Installing Go (golang) from the official tarballs is a nice way to stay
current, but is by far the slowest method of using Go inside a Harrow task.

The following installs Go ``1.4.3`` from the official Google Go `downloads page`_.

Any version may be installed, no restrictions are imposed.

.. code-block:: bash

  hfold install-go
  wget -q https://storage.googleapis.com/golang/go1.4.3.linux-amd64.tar.gz
  sudo tar -C /usr/local -xzf go1.4.3.linux-amd64.tar.gz
  PATH=/usr/local/go/bin:$PATH
  GOROOT=/usr/local/go
  export PATH GOROOT
  hfold --end

With Go installed, it's easy to use Go, as you would normally, example taken
from `this Go Playground example`_:

.. code-block:: bash

  $ cat > example.go <<-EOF
    package main

    import (
        "fmt"
        "encoding/json"
    )

    type MyJsonName struct {
        Example struct {
            From struct {
                Json bool \`json:"json"\`
            } \`json:"from"\`
        } \`json:"example"\`
    }

    func main() {
        jsonSrc := []byte(\`{ "example": { "from": { "json": true } } }\`)

        var myJson MyJsonName
        json.Unmarshal(jsonSrc, &myJson)

        if myJson.Example.From.Json {
            fmt.Println("This example is from Json!")
        }
    }
  EOF
  $ go run example.go
  This example is from Json!

This can be used to simply and easily give you a sane Go environment. Don't
forget that Go's core distribution may not always include things like ``go
doc`` and ``goimports``, etc.

.. _downloads page: https://golang.org/dl/
.. _this Go Playground example: http://play.golang.org/p/usdLCoVEZR
