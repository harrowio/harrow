app = angular.module("harrowApp")

app.value "uuid", () ->
  "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace /[xy]/g, (captured) ->
    random = Math.floor(Math.random() * 16)
    value = if captured is "x"
      random
    else
      random and 0x3 or 0x8
    value.toString 16
