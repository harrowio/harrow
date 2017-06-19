describe('Directive: harrowFieldInput', function () {
  var el, $compile, $scope
  beforeAll(function () {
    angular.module('harrowApp').config(function ($translateProvider) {
      $translateProvider.translations('en_test', {
        forms: {
          testForm: {
            label: {
              name: 'Your Name',
              email: 'Your Email'
            }
          }
        },
        errors: {
          field: {
            required: 'Required',
            invalid_totp_token: 'Must be a number with six digits',
            too_short: 'Too short',
            email: 'Must be an email address',
            minlength: 'Too short',
            maxlength: 'Too long',
            number: 'Must be a number',
            unique_violation: 'Already in use',
            match: 'Does not match'
          }
        }
      })
      $translateProvider.preferredLanguage('en_test')
    })
  })
  beforeEach(angular.mock.inject(function (_$compile_, $rootScope, $translate) {
    $scope = $rootScope
    $compile = _$compile_
    $translate
  }))
  it('expands `harrow-field-input` directive', function () {
    $scope.user = {
      name: ''
    }
    var input = `<div>
      <harrow-form translation-root="forms.testForm" no-controls>
        <div class="field__group">
          <harrow-field-input>
            <input name="name" ng-model="user.name" ng-required="true" required="true">
          </harrow-field-input>
        </div>
      </harrow-form>
    </div>`
    el = $compile(input)($scope)
    $scope.$digest()
    expect(el.find('.field__input').length).toEqual(1)
    expect(el.find('.field__input > label').length).toEqual(1)
    expect(el.find('.field__input > label').text()).toEqual('Your Name')
    expect(el.find('.field__input > input').length).toEqual(1)
    expect(el.find('.field__input > span').attr('data-error-messages')).toEqual('Required')
  })

  it("doesn't leak scope label", function () {
    $scope.user = {
      name: '',
      email: ''
    }
    var input = `<div>
      <harrow-form translation-root="forms.testForm" no-controls>
        <div class="field__group">
          <harrow-field-input>
            <input name="name" ng-model="user.name" ng-required="true" required="true">
          </harrow-field-input>
          <harrow-field-input>
            <input name="email" ng-model="user.email" ng-required="true" required="true">
          </harrow-field-input>
        </div>
      </harrow-form>
    </div>`
    el = $compile(input)($scope)
    $scope.$digest()

    expect(el.find('.field__input:nth-child(1)').length).toEqual(1)
    expect(el.find('.field__input:nth-child(1) > label').length).toEqual(1)
    expect(el.find('.field__input:nth-child(1) > label').text()).toEqual('Your Name')
    expect(el.find('.field__input:nth-child(1) > input').length).toEqual(1)
    expect(el.find('.field__input:nth-child(1) > span').attr('data-error-messages')).toEqual('Required')

    expect(el.find('.field__input:nth-child(2)').length).toEqual(1)
    expect(el.find('.field__input:nth-child(2) > label').length).toEqual(1)
    expect(el.find('.field__input:nth-child(2) > label').text()).toEqual('Your Email')
    expect(el.find('.field__input:nth-child(2) > input').length).toEqual(1)
    expect(el.find('.field__input:nth-child(2) > span').attr('data-error-messages')).toEqual('Required')
  })
})
