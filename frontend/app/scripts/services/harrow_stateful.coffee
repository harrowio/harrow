###*
# @ngdog Service
# @name harrowStateful
# @description Handles the transitioning and callbacks for stateful events.
###

###*
# @callback listenerCallback
# @param {Object} options
###
angular.module('harrowApp').factory 'Stateful', ($timeout, $log) ->
  Stateful = () ->
    @state = 'initial'
    @timeout = null
    @softTimeout = null
    @listeners = {}
    @isTerminal = true
    @

  ###*
  # @name on
  # @param {string} state - State name to listen to
  # @param {listenerCallback} listener - The callback that is called when a state is transitioned to.
  ###
  Stateful::on = (state, listener) ->
    @listeners[state] = [] unless angular.isObject(@listeners[state])
    @listeners[state].push listener

  ###*
  # @name transitionTo
  # @description will transition to given state.
  #   Then after 10 seconds a `softTimeout` is issued.
  #   And after 50 more seconds a `timeout` is issued
  # @param {string} state - State name to transition to
  # @param {Object} options - Configuration options for the transition
  # @param {number} options.timeout - The maximum duration before changing state to timeout
  # @param {number} options.terminal - Acts as a end state, and disables timeout.
  ###
  Stateful::transitionTo = (state, options = {}) ->
    return if state == @state
    $log.debug("Stateful::transitionTo", state, options)
    if state == 'error' || state == 'complete' || state == 'success'
      options.terminal = true
    options.terminal = false unless options.terminal == true
    options.timeout = 60000 unless angular.isNumber(options.timeout)
    options.softTimeout = 10000 unless angular.isNumber(options.softTimeout)

    @isTerminal = options.terminal

    @state = state
    if angular.isArray(@listeners[state])
      @listeners[state].forEach (listener) ->
        listener(options)

    if state != 'softTimeout'
      if @softTimeout
        $timeout.cancel(@softTimeout)

      if @timeout and state != 'timeout'
        $timeout.cancel(@timeout)

    if options.terminal == false and state != 'timeout' and state != 'softTimeout'
      $log.debug('Stateful::transitionTo Setting Timeout')
      @timeout = $timeout =>
        @transitionTo 'timeout', terminal: true
        return
      , options.timeout
      @timeout.catch =>
        $log.debug('Stateful::transitionTo HardTimeout Canceled')
        return
      @timeout.then =>
        @timeout = undefined
        return

      @softTimeout = $timeout =>
        @transitionTo 'softTimeout'
        return
      , options.softTimeout
      @softTimeout.catch =>
        $log.debug('Stateful::transitionTo SoftTimeout Canceled')
        return
      @softTimeout.then =>
        @softTimeout = undefined
        return

  Stateful
