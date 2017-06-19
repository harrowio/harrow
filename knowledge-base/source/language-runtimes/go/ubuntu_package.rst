Installing Go (golang) in Harrow using an Debian/Ubuntu package
===============================================================

Installing Go (golang) from an Ubuntu package is one of the quickest ways to
get Go installed in your Harrow container however you are at the mercy of the
Ubuntu package maintainers.

The current Go version in the package repository is ``1.5`` is available at
``http://packages.ubuntu.com/precise/devel/golang-go``:

.. code-block:: bash

  hfold install-go
  sudo apt-get update
  sudo apt-get install -y golang
  hfold --end

With Go installed, it's easy to use Go, as you would normally, example taken
from `this Go Playground example`_:

.. code-block:: bash

  #!/bin/bash -e

  hfold install-go
  sudo apt-get update
  sudo apt-get install -y golang
  hfold --end

  cat > example.go <<-EOF
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
  go run example.go

This can be used to simply and easily give you a sane Go environment. Don't
forget that Go's core distribution may not always include things like ``go
doc`` and ``goimports``, etc.

.. _downloads page: https://golang.org/dl/
.. _this Go Playground example: http://play.golang.org/p/usdLCoVEZR
