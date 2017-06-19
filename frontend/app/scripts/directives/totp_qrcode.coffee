QRCode = require 'qrcodejs2'

app = angular.module("harrowApp")

Controller = (
  $element
  $scope
) ->
  $scope.$watchGroup ["ctrl.email", "ctrl.secret"], ([email, secret]) =>
    if email?.length and secret?.length
      $element.empty()
      url = "otpauth://totp/Harrow:#{email}?secret=#{secret}&issuer=Harrow"
      new QRCode($element[0], url)

  @

TotpQRCode = () ->
  restrict: "E"
  scope:
    email: "@"
    secret: "@"
  bindToController: true
  controller: Controller
  controllerAs: "ctrl"

app.directive("totpQrCode", TotpQRCode)
