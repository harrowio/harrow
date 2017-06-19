describe('Routing: triggers', function () {
  var authentication, $scope, $state
  beforeEach(angular.mock.inject(function ($rootScope, _$state_, _authentication_) {
    authentication = _authentication_
    $scope = $rootScope
    $state = _$state_
    spyOn(authentication, 'hasValidSession').and.callFake(function () {
      return true
    })
    spyOn(authentication, 'hasNoSession').and.callFake(function () {
      return false
    })
  }))

  it('passes through to triggers for Project', function () {
    var state = $state.get('triggers', {projectUuid: 'abc123'})

    expect(state.parent).toEqual('projects/edit')
    expect(state.name).toEqual('triggers')
    expect(state.url).toEqual('/triggers?{scriptUuid}')
    expect(state.data.requiresAuth).toBeTruthy()
  })

  it('passes through to triggers for Task', function () {
    var state = $state.get('triggers', {projectUuid: 'abc123', taskUuid: 'abc123'})

    expect(state.parent).toEqual('projects/edit')
    expect(state.name).toEqual('triggers')
    expect(state.url).toEqual('/triggers?{scriptUuid}')
    expect(state.data.requiresAuth).toBeTruthy()
  })

  it('passes though to new trigger for Project', function () {
    var state = $state.get('triggers.gitTrigger', {projectUuid: 'abc123'})

    expect(state.name).toEqual('triggers.gitTrigger')
    expect(state.url).toEqual('/git')
    var project = {
      subject: {
        uuid: 'abc123'
      }
    }
    expect(state.resolve.trigger(project, null).subject.projectUuid).toEqual('abc123')
  })
})
