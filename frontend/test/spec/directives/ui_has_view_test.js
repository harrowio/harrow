describe('Directive: uiHasView', function () {
  var el, $state, $compile, $scope

  beforeAll(function () {
    angular.module('harrowApp').config(function ($stateProvider) {
      $stateProvider.state('test-dashboard', {
        parent: 'layout',
        views: {
          main: {
            templateUrl: 'test-dashboard.html'
          }
        }
      })
      $stateProvider.state('test-dashboard-with-header', {
        parent: 'layout',
        data: {
          showViews: ['header']
        },
        views: {
          main: {
            templateUrl: 'test-dashboard.html'
          },
          'header@layout': {
            templateUrl: 'test-header.html'
          }
        }
      })
    })
  })

  beforeEach(function () {
    angular.mock.inject(function (_$rootScope_, _$compile_, _$state_, $q, $httpBackend, $templateCache, authentication) {
      spyOn(authentication, 'hasValidSession').and.callFake(function () {
        return true
      })
      $state = _$state_
      $compile = _$compile_
      $scope = _$rootScope_
      $templateCache.put('test-dashboard.html', 'DASHBOARD!!!')
      $templateCache.put('test-header.html', 'HEADER!!!')
    })
  })

  it('shows element when view is defined', function () {
    $state.go('test-dashboard-with-header')
    el = $compile('<div><ui-view></ui-view></div>')($scope)
    $scope.$digest()
    expect(el.find('[ui-view="header"]').length).toEqual(1, 'has header')
    expect(el.find('[ui-view="header"]')[0].style.display).not.toEqual('none')
  })

  it('doesnt show the element header element', function () {
    $state.go('test-dashboard')
    el = $compile('<div><ui-view></ui-view></div>')($scope)
    $scope.$digest()
    expect(el.find('[ui-view="header"]').length).toEqual(1)
    expect(el.find('[ui-view="header"]')[0].style.display).toEqual('none')
  })
})
