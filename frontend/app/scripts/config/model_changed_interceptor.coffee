###*
* @ngdoc config
* @name modelChangedHTTPInterceptor
* @description intercepts all POST, PUT, PATCH, DELETE requests and emits
# #{tableName}Changed event to from the $rootScope
# @example
<example>
  <file name="index.html">
    <script>
      angular.module('harrowApp', []).
        controller('ExampleCtrl', function ($scope) {
          $scope.$on('projectsChanged', function(event, uuid, response) {
            $scope.projects[uuid] = response.data
          })

          $scope.createProject = function () {
            $http({
              method: "POST",
              url: 'http://test.host/api/projects/'
            })
          }
        })
    </script>
    <a ng-click="createProject()">Create</a>

    {{projects}}
  </file>
</example>
###
ModelChangedHTTPInterceptor = (
  @$rootScope
  $filter
) ->
  @broadcastChangeEvent = (response) ->
    matcher = response.config.url.match(/\/api\/([\w-]+)(?:\/([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}))?$/)
    return unless matcher
    tableName = $filter('camelCase')(matcher[1])
    tableName = $filter('singularize')(tableName)
    if /^(POST|PUT|PATCH|DELETE)$/.test(response.config.method)
      @$rootScope.$broadcast "#{tableName}Changed", matcher[2], response
  {
    response: (response) =>
      @broadcastChangeEvent(response)
      response
  }

angular.module('harrowApp').factory 'modelChangedHTTPInterceptor', ModelChangedHTTPInterceptor
angular.module('harrowApp').config ($httpProvider) ->
  $httpProvider.interceptors.push 'modelChangedHTTPInterceptor'
