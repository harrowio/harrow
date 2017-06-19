describe "Directive: harrowForm", () ->

  beforeAll ->
    angular.module('harrowApp').config ($translateProvider) ->
      $translateProvider.translations 'en_test',
        forms:
          testForm:
            submit: 'Send'
            label:
              name: "Name"
      $translateProvider.preferredLanguage('en_test')


  beforeEach angular.mock.inject (@$controller, @$compile, @$rootScope, @$timeout, @$translate) ->

  describe 'Intergration with `harrow-input`', ->
    it 'includes the inputs', () ->
      scope = @$rootScope.$new()
      el = @$compile('<harrow-form translation-root="forms.testForm"><harrow-input><input name="name" ng-model="testForm.name" required="true" /></harrow-input></harrow-form>')(scope)
      scope.$digest()
      expect(el.find('input').length).toEqual(1)
      expect(el.find('label').text()).toEqual('Name')
      expect(el.find('.btn--primary').text()).toEqual('Send')

  it 'has translationRoot on scope', () ->
    el = @$compile('<harrow-form translation-root="forms.testForm"></harrow-form>')(@$rootScope)
    @$rootScope.$digest()

    expect(el.scope().harrowForm).toBeDefined()
    expect(el.scope().harrowForm.translationRoot).toBeDefined()
    expect(el.scope().harrowForm.translationRoot).toEqual('forms.testForm')


  describe "when no-controls is absent", ->
    it "should not remove form actions", () ->
      source = '<harrow-form><a class="btn">Button</a></harrow-form>'
      el = @$compile(source)(@$rootScope)
      @$rootScope.$digest()
      expect(el.find('.btn').length).toEqual(2)
      expect(el.find(".form-actions").length).toEqual(1)
      expect(el.find('.form-actions .btn').length).toEqual(2)

  describe "when no-controls is present", ->
    it "should not add form actions", () ->
      source = '<harrow-form no-controls><p>hello world</p><a class="btn">Button</a></harrow-form>'
      el = @$compile(source)(@$rootScope)
      @$rootScope.$digest()
      @$rootScope.$digest()
      expect(el.find('.btn').length).toEqual(1)
      expect(el.find(".form-actions").length).toEqual(0)
      expect(el.find('.form-actions .btn').length).toEqual(0)
