app = angular.module("harrowApp")

app.config ($provide) ->
  $provide.decorator '$uiViewScroll', ($delegate) ->
    (uiViewElement) ->
      window.scrollTo(0, 0);
