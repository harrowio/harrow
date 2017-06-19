###*
* @ngdoc config
* @name serverErrorHTTPInterceptor
* @description intercepts all requests with the server and broadcasts a event on the app root if there is a issue with the server.
# @example
<example>
  <file name="index.html">
    <script>
      angular.module('harrowApp', []).
        controller('ExampleCtrl', function ($scope) {
          $scope.$on('http.serverError', function(event, status) {
            alert("Server Error: ", status)
          })

          $scope.checkServer = function () {
            $http({
              method: "POST",
              url: 'http://test.host/api/projects/'
            })
          }
        })
    </script>
    <a ng-click="checkServer()">Create</a>
  </file>
</example>
###
ServerErrorHTTPInterceptor = (
  @$rootScope
  @$q
) ->
  @broadcastServerError = (response) ->
    if /^5\d{2,}$/.test(response.status)
      @$rootScope.$broadcast 'http.serverError', response.status, response.data
    if response.status == 403
      @$rootScope.$broadcast 'http.forbidden', response.status, response.data
    if response.status == 402 || response.data?.reason == 'limits_exceeded'
      @$rootScope.$broadcast 'http.paymentRequired', response.status, response.data
  {
    responseError: (response) =>
      @broadcastServerError(response)
      @$q.reject(response)
  }

app = angular.module('harrowApp')

app.factory 'serverErrorHTTPInterceptor', ServerErrorHTTPInterceptor
app.config ($httpProvider) ->
  $httpProvider.interceptors.push 'serverErrorHTTPInterceptor'
