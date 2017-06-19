describe('Filter: toTrusted', function () {
  var api, $compile, $scope
  beforeEach(angular.mock.inject(function ($filter, _$compile_, $rootScope) {
    $compile = _$compile_
    $scope = $rootScope.$new()
    api = $filter('toTrusted')
  }))
  it('does not raise error about $sce', function () {
    $scope.html = '<b>strong</b>'
    var actual = $compile('<div ng-bind-html="html | toTrusted"></div>')($scope)
    expect(function () {$scope.$digest()}).not.toThrowError(/^\[\$sce:unsafe\] Attempting to use an unsafe value in a safe context\./)
    expect(actual[0].innerHTML).toEqual('<b>strong</b>')
  })

  it('raises error about $sce', function () {
    $scope.html = '<b>strong</b>'
    var actual = $compile('<div ng-bind-html="html"></div>')($scope)
    expect(function () {$scope.$digest()}).toThrowError(/^\[\$sce:unsafe\] Attempting to use an unsafe value in a safe context\./)
    expect(actual[0].innerHTML).toEqual('')
  })
})
