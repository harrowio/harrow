window.braintree =
  api:
    Client: (obj, cb) ->
      return {
        tokenizeCard: (card, cb) ->
          return cb(null, '28a6ed48-4c87-4621-8d52-0f99ea248152')
      }

nestedKeys = (results, validations, expected) ->
  Object.keys(validations).forEach (key) ->
    if typeof validations[key] == 'object'
      nestedKeys(results, validations[key], expected[key])
    else if typeof validations[key] == 'function'
      if validations[key] == String
        result = (typeof expected == 'string' || expected instanceof String)
      throw new Error("Expected #{validations} to equal #{expected} for #{key}") unless result
      results.push result
    else
      result = validations[key] == expected[key]
      throw new Error("Expected #{validations[key]} to equal #{expected[key]} for #{key}") unless result
      results.push result
    return
  results

jasmine.validateHttpParams = (validations) ->
  (data) ->
    results = []
    json = JSON.parse(data)
    nestedKeys(results, validations, json)
    response = results.every((result) ->
      result == true
    )
    if response == false
      throw new Error('Expected ' + data + ' to contain ' + JSON.stringify(validations))
    response

jasmine.getJSONFixture = (file) ->
  return require("../fixtures/#{file}")

ICMock = (
) ->
  @

ICMock::boot = () ->

ICMock::newMessage = (msg) ->

ICMock::onTransition = () ->

ICMock::user = () ->

appMock = angular.module('harrowAppMock', ['ngMockE2E', 'harrowApp'])

appMock.constant("endpoint", "test.local/api")
appMock.service('ic', ICMock)
appMock.run ($httpBackend) ->
  $httpBackend.when('GET', /views\/.*\.html/).respond(200, '<div></div>')
  $httpBackend.when('GET', /icons\.svg/).respond(200, '')
  $httpBackend.when('PUT', /\/api\/sessions/).respond(200, jasmine.getJSONFixture('PUT_api_sessions.json'))
  $httpBackend.when('GET', /\/api\/api-features/).respond(200, jasmine.getJSONFixture('GET_api_app_features.json'))

beforeEach angular.mock.module("harrowAppMock"), () ->

beforeEach angular.mock.inject ($urlRouter) ->
  spyOn($urlRouter, "sync").and.stub()

afterEach angular.mock.inject (authentication) ->
  authentication.clear()

afterEach angular.mock.inject ($httpBackend, $rootScope) ->
  unless $rootScope.$$phase
    $httpBackend.verifyNoOutstandingExpectation()
    $httpBackend.verifyNoOutstandingRequest()
    $httpBackend.resetExpectations()

jasmine.authenticate = () ->
  angular.mock.inject ($httpBackend) ->
    localStorage.setItem('Harrow-Session-Uuid', 'abc123')
    $httpBackend.expect('GET', /\/api\/sessions\/[0-9a-f-]+/)
      .respond(200, jasmine.getJSONFixture('PUT_api_sessions.json'))
    $httpBackend.expect('GET', /\/api\/users\/[0-9a-f-]+/)
      .respond(200, jasmine.getJSONFixture('GET_api_user.json'))

    $httpBackend.flush(2)
