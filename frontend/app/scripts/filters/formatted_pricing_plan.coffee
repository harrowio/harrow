angular.module('harrowApp').filter 'formattedPricingPlan', ($translate) ->
  (subject) ->
    out = []
    out.push $translate.instant('formattedPricingPlan.users', {users: subject.usersIncluded}, 'messageformat')
    if subject.uuid == "f975a385-3625-4883-b353-8f1febeb5b3e"
      out.push $translate.instant('formattedPricingPlan.unlimitedProjects', {}, 'messageformat')
    else
      out.push $translate.instant('formattedPricingPlan.projects', {projects: subject.projectsIncluded}, 'messageformat')
    out.push $translate.instant('formattedPricingPlan.additionalUsers', {price: subject.pricePerAdditionalUser})
    out.push $translate.instant('formattedPricingPlan.platform')
    out.join '<br>'
