describe('Service: harrowStateful', function () {
  var $timeout, api
  beforeEach(angular.mock.inject(function (_$timeout_, Stateful) {
    $timeout = _$timeout_
    api = new Stateful()
  }))

  describe('.on', function () {
    it('calls event', function (done) {
      api.on('start', function () {
        expect(true).toBeTruthy()
      })
      api.on('start', function () {
        expect(true).toBeTruthy()
        done()
      })
      api.transitionTo('start')
    })
  })

  describe('.transitionTo', function () {
    it('transitions to timeout', function () {
      api.transitionTo('start')
      $timeout.flush(10000)
      expect(api.state).toEqual('softTimeout')
      $timeout.flush(60000)
      expect(api.state).toEqual('timeout')
    })

    it('transitions to timeout after defined timeout', function () {
      api.transitionTo('start', {timeout: 50})
      $timeout.flush(50)
      expect(api.state).toEqual('timeout')
    })

    it('retains state', function () {
      api.transitionTo('completed', {terminal: true})
      $timeout.flush()
      expect(api.state).toEqual('completed')
    })

    it('cancels timeout when set to idle', function () {
      api.transitionTo('start')
      $timeout.flush(10000)
      api.transitionTo('idle', {terminal: true})
      $timeout.flush(60000)
      expect(api.state).toEqual('idle')
    })
  })
})
