Using Go (golang) in Harrow via the official Docker container
=============================================================

Using the `official Go Docker containers`_ is one of the fastest and most
reliable ways to use Go on Harrow.

The containers are pulled directly each time, which can be time consuming (a
few seconds), but will eventually be cached in a Harrow-local Docker repository
mirror.

This example is taken from `the Go Playground`_, the last line is critical,
mounting the directory (``-v``) and changing the working directory in the
container (``-w``) before running a command (``go run example.go``):

.. code-block:: bash

   #!/bin/bash -e

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

   sudo docker run -v "$PWD":/usr/src/example -w /usr/src/example golang:latest go run example.go

.. _official Go Docker containers: https://hub.docker.com/_/golang/
.. _the Go Playground: http://play.golang.org/p/usdLCoVEZR
