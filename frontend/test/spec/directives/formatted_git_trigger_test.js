describe('Directive: formattedGitTrigger', function () {
  var el, $compile, $scope
  beforeEach(angular.mock.inject(function (_$compile_, $rootScope, $translate) {
    $scope = $rootScope
    $compile = _$compile_
    $translate
  }))
  it('expands `formatted-git-trigger` directive', function () {
    $scope.trigger = {
      subject: {
        matchRef: 'master',
        changeType: 'changed'
      }
    }
    $scope.script = {
      subject: {
        name: 'Unit Tests'
      }
    }
    $scope.environment = {
      subject: {
        name: 'Development'
      }
    }
    var input = `<div formatted-git-trigger="trigger" ng-init="{script: script, environment: environment}">`
    el = $compile(input)($scope)
    $scope.$digest()
    expect(el.text()).toEqual('When master is changed then trigger Unit Tests on Development')
    expect(el.html()).toEqual('When master is changed then trigger <strong>Unit Tests</strong> on <strong>Development</strong>')
  })
})
