;(function (window) {
  var sidebar = document.querySelector('.sidebar')
  var sidebarAccount = document.querySelector('.sidebar__header__account')
  var el = document.querySelector('.tab--sidebarToggle')
  if (el) {
    el.addEventListener('click', function (e) {
      e.preventDefault()
      sidebar.classList.toggle('sidebar--open')
      return false
    }, false)

    function closeSidebar (event) {
      if (event.target !== el && event.target.parentElement !== el) {
        sidebar.classList.remove('sidebar--open')
      }
    }
    document.querySelector('.layout__header').addEventListener('click', closeSidebar)
    document.querySelector('.layout__content').addEventListener('click', closeSidebar)
    document.querySelector('.sidebar').addEventListener('click', closeSidebar)
  }
  if (sidebarAccount) {
    sidebarAccount.addEventListener('click', function (e) {
      sidebar.querySelector('.sidebar__content').classList.toggle('sidebar__content--second')
    })
  }
})(window)
