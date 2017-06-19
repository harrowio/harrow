Modal = (
  @$rootScope
  @$compile
  @$document
  @$q
) ->
  @
Modal::show = (config) ->
  deferred = @$q.defer()
  throw new Error('Must define `modal` in config') unless config.modal
  if config.templateUrl
    html = """<div modal="#{config.modal}" modal-mode="#{config.mode}" modal-full-screen>
      <div ng-include="'#{config.templateUrl}'"></div>
    </div>"""
  else if config.templateFn
    html = config.templateFn(config)
  else
    html = """<div modal="#{config.modal}" modal-mode="#{config.mode}" modal-full-screen>
      <h2>#{config.title}</h2>
      <p>#{config.content}</p>
      <a href="#{config.href}" class="btn btn--border">#{config.name}</a>
    </div>"""
  compiled = @$compile(html)(@$rootScope)
  @$document.find('body').append(compiled)
  @$rootScope.$on 'modal:dismissed', deferred.resolve
  return deferred.promise

angular.module('harrowApp').service 'modal',Modal
