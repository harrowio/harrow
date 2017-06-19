app = angular.module("harrowApp")

RepositoryImportCtrl = (
  @organization
  @project
  @repositories
  @flash
  @$translate
  oauth
  @repositoryResource
) ->
  @ghScopes = undefined
  oauth.getGithubUser().then (user) =>
    oauth.getGithubUserRepositories(user.data.login).then (repos) =>
      @ghScopes ||= {}
      @ghScopes[user.data.login] = repos.data

    oauth.getGithubOrganizations().then (orgs) =>
      angular.forEach orgs.data, (org) =>
        oauth.getGithubOrgRepositories(org.login).then (repos) =>
          @ghScopes ||= {}
          @ghScopes[org.login] = repos.data
  @

RepositoryImportCtrl::import = (ghRepo) ->
  @repositoryResource.save
    subject:
      name: ghRepo.full_name
      url: if ghRepo.private then ghRepo.ssh_url else ghRepo.clone_url
      projectUuid: @project.subject.uuid
      githubImported: true
      githubLogin: ghRepo.owner.login
      githubRepo: ghRepo.name
  .then (repository) =>
    @repositories.push repository
    @flash.success = @$translate.instant("oauth.github.importGithubRepository.success")
    return
  .catch (repository) =>
    @flash.error = @$translate.instant("oauth.github.importGithubRepository.fail", repository)
    return


RepositoryImportCtrl::alreadyImported = (ghRepo) ->
  imported = false
  angular.forEach @repositories, (repo) ->
    imported ||= repo.subject.url == ghRepo.ssh_url || repo.subject.url == ghRepo.clone_url
  imported

RepositoryImportCtrl::importedToProjectUuid = (ghRepo) ->
  projectUuid = false
  angular.forEach @repositories, (repo) ->
    if repo.subject.url == ghRepo.ssh_url || repo.subject.url == ghRepo.clone_url
      projectUuid = repo.subject.projectUuid
  projectUuid

app.controller("repositoryImportCtrl", RepositoryImportCtrl)
