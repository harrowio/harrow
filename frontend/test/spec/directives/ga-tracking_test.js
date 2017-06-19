describe('Directive gaTracking', function () {
  var $scope, $compile, $window
  beforeEach(angular.mock.inject(function ($rootScope, _$compile_, _$window_) {
    $compile = _$compile_
    $window = _$window_
    $scope = $rootScope
  }))

  it('binds a on-click event to "window.ga"', function () {
    $window.ga = function () {}
    spyOn($window, 'ga')
    var el = $compile('<a ga="\'send\', \'event\', \'user\', \'knowledgebase\'"></a>')($scope)
    el.triggerHandler('click')
    expect($window.ga).toHaveBeenCalledWith('send', 'event', 'user', 'knowledgebase')
  })
})
