describe('Describe: clipboard', function () {
  var $compile, $scope, feature
  beforeEach(angular.mock.inject(function (_$compile_, $rootScope, _feature_) {
    $compile = _$compile_
    feature = _feature_
    $scope = $rootScope
  }))

  it('expands clipboard element', function () {
    $scope.thing = 'hello world'
    var el = $compile('<div clipboard="thing">Click to copy</div>')($scope)
    $scope.$digest()
    expect(el.text()).toEqual('Copy to Clipboard')
  /* sadly, can't trigger a click on the flash element
  el.triggerHandler('click')
  expect(el.text()).toEqual('Copied to Clipboard')
  */
  })
})
