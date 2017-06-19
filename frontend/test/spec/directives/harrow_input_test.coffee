describe "Directive: harrowInput", () ->

  beforeEach angular.mock.inject (@$compile, @$rootScope, @$timeout) ->

  # these are examples for tags that must receive a .form-control
  positives = [
    "<select></select>"
    "<textarea></textarea>"
    "<input>"
    "<input type=\"text\">"
    """
<textarea ui-ace="{mode: 'sh', theme: 'solarized_light'}" />
    """
  ]
  # these are examples for tags that must not receive a .form-control
  negatives = [
    "<input type=\"radio\">"
    "<input type=\"checkbox\">"
  ]

  for p in positives
    it "should add .form-control to #{p} elements", ->
      source = "<div><harrow-input>#{p}</harrow-input></div>"
      el = @$compile(source)(@$rootScope)
      @$rootScope.$digest()
      # because manual transclusion with $evalAsync
      @$timeout () ->
        expect(el.find(".form-control").length).toEqual(1)

  for n in negatives
    it "should not add .form-control to #{n} elements", ->
      source = "<harrow-input>#{n}</harrow-input>"
      el = @$compile(source)(@$rootScope)
      @$rootScope.$digest()
      @$timeout () ->
        expect(el.find(".form-control").length).toEqual(0)
