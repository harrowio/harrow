Controller = (
  @ga
  @project
  @invitationResource
  @flash
  @$q
  @$state
  @$interpolate
  @$translate
) ->
  @recipients = [
    {email: ""}
  ]
  @invitation =
    subject:
      projectUuid: @project.subject.uuid
      message: @$interpolate('''Hey,

I've been setting up some of our DevOps tasks on Harrow.io, I'd like for you to accept this invitation and join me.

I've setup {{invitePeople.tasks.length}} task(s) for {{wizardUsers.organization.subject.name}} {{wizardUsers.project.name}}. The tasks are connected to our {{wizardUsers.repositories[0].subject.url}} repository.

See you there.''')({wizardUsers: @})
  @

Controller::addAnotherRow = ->
  @recipients.push {email: ''}

Controller::save = () ->
  @ga 'send', 'event', 'wizard', 'users', 'formSubmitted'
  promises = []
  for recipient in @recipients
    @invitation.subject.recipientName  = recipient.name
    @invitation.subject.email          = recipient.email
    @invitation.subject.membershipType = 'member'
    promises.push @invitationResource.save(@invitation)

  @$q.all(promises).then (responses) =>
    @ga 'send', 'event', 'wizard', 'users', 'formSuccess'
    @flash.success = @$translate.instant('forms.wizard.users.flashes.success', {count: responses.length}, 'messageformat')
    @$state.go('projects/show', {projectUuid: @project.subject.uuid})
    return
  .catch (reason) =>
    @ga 'send', 'event', 'wizard', 'users', 'formError'
    @flash.error = @$translate.instant('forms.wizard.users.flashes.fail')
    @$q.reject(reason)

angular.module('harrowApp').controller 'wizardUsersCtrl', Controller
