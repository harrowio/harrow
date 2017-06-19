describe('Routing: verification', function () {
  var $scope, $state, $httpBackend
  beforeEach(angular.mock.inject(function ($rootScope, _$state_, _$httpBackend_, userResource) {
    $httpBackend = _$httpBackend_
    $scope = $rootScope
    $state = _$state_

    jasmine.authenticate()
  }))

  it('handles verification sent route', function () {
    var state = $state.get('errors/verification_email_sent')

    expect(state.parent).toEqual('layout_tight')
    expect(state.name).toEqual('errors/verification_email_sent')
    expect(state.url).toEqual('/a/errors/verification_email_sent')
    expect(state.views['main'].templateUrl).toEqual('views/errors/verification_email_sent.html')
    expect(state.views['main'].controller).toEqual('errorCtrl')
    expect(state.views['main'].controllerAs).toEqual('error')
  })
})
