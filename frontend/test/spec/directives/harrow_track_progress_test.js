describe('Directive: trackProgress', function () {
  var el, $scope, $compile
  beforeEach(angular.mock.inject(function (_$compile_, $rootScope) {
    $scope = $rootScope
    $compile = _$compile_
  }))

  it('renders basic track', function () {
    el = $compile('<div track-progress="0"></div>')($scope)
    $scope.$digest()
    expect(el.find('.trackProgress__station').length).toEqual(2)
    expect(el.find('.trackProgress__station:first').hasClass('active')).toBeTruthy()
    expect(el.find('.trackProgress__station:last-child').hasClass('active')).toBeFalsy()
  })

  it('renders partial filled track', function () {
    el = $compile('<div track-progress="1"></div>')($scope)
    $scope.$digest()
    expect(el.find('.trackProgress__station').length).toEqual(2)
    expect(el.find('.trackProgress__station:first').hasClass('active')).toBeTruthy()
    expect(el.find('.trackProgress__station:first').hasClass('completed')).toBeTruthy()
    expect(el.find('.trackProgress__station:last-child').hasClass('active')).toBeTruthy()
    expect(el.find('.trackProgress__station:last-child').hasClass('completed')).toBeFalsy()
  })

  it('renders full track', function () {
    el = $compile('<div track-progress="2"></div>')($scope)
    $scope.$digest()
    expect(el.find('.trackProgress__station').length).toEqual(2)
    expect(el.find('.trackProgress__station:first').hasClass('active')).toBeTruthy()
    expect(el.find('.trackProgress__station:first').hasClass('completed')).toBeTruthy()
    expect(el.find('.trackProgress__station:last-child').hasClass('active')).toBeTruthy()
    expect(el.find('.trackProgress__station:last-child').hasClass('completed')).toBeTruthy()

    el = $compile('<div track-progress="2" track-progress-stations="4"></div>')($scope)
    $scope.$digest()
    expect(el.find('.trackProgress__station').length).toEqual(4)
    expect(el.find('.trackProgress__station:first').hasClass('active')).toBeTruthy()
    expect(el.find('.trackProgress__station:first').hasClass('completed')).toBeTruthy()
    expect(el.find('.trackProgress__station:last-child').hasClass('active')).toBeTruthy()
    expect(el.find('.trackProgress__station:last-child').hasClass('completed')).toBeTruthy()
  })
})
