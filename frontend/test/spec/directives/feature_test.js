describe('Describe: feature', function () {
  var $compile, $scope, feature
  beforeEach(angular.mock.inject(function (_$compile_, $rootScope, _feature_) {
    $compile = _$compile_
    feature = _feature_
    $scope = $rootScope
  }))

  it('hides a code block by default', function () {
    var el = $compile('<div feature="hello.world">hidden</div>')($scope)
    $scope.$digest()
    expect(el.hasClass('ng-hide')).toBeTruthy()
  })

  it('shows null keys', function () {
    var el = $compile('<div feature="">hidden</div>')($scope)
    $scope.$digest()
    expect(el.hasClass('ng-hide')).toBeFalsy()
  })

  it('shows if feature is enabled', function () {
    feature.enabledFeatures.push({
      subject: {
        name: 'hello.world',
        enableAt: new Date(0)
      }
    })
    var el = $compile('<div feature="hello.world">hidden</div>')($scope)
    $scope.$digest()
    expect(el.hasClass('ng-hide')).toBeFalsy()
  })

  it('shows if feature is disabled', function () {
    var el = $compile('<div no-feature="hello.world">hidden</div>')($scope)
    $scope.$digest()
    expect(el.hasClass('ng-hide')).toBeFalsy()
  })

  it('hides if feature is enabled', function () {
    feature.enabledFeatures.push({
      subject: {
        name: 'hello.world',
        enabled: true
      }
    })
    var el = $compile('<div no-feature="hello.world">hidden</div>')($scope)
    $scope.$digest()
    expect(el.hasClass('ng-hide')).toBeTruthy()
  })
})
