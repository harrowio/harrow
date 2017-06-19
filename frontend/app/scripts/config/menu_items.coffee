menuItems = {}

menuItems.notifier = {}
menuItems.notifier.slackNotifier = { name: 'Slack', notifierType: 'slackNotifier', icon: 'icon-slack', sref: 'notifiers.slackNotifier' }
menuItems.notifier.emailNotifier = { name: 'Email', notifierType: 'emailNotifier', icon: 'icon-mail', sref: 'notifiers.emailNotifier' }
menuItems.notifier.taskNotifier = { name: 'Task', notifierType: 'taskNotifier', icon: 'icon-tasks', sref: 'notifiers.taskNotifier' }


menuItems.notifiers = [
  menuItems.notifier.slackNotifier,
  menuItems.notifier.taskNotifier
  # menuItems.notifier.emailNotifier,
]
# menuItems.notifiers.push { name: 'Slack', notifierType: 'slackNotifier', icon: 'icon-slack', sref: 'notifiers.slackNotifier' }
# menuItems.notifiers.push { name: 'Email', notifierType: 'emailNotifier', icon: 'icon-mail', sref: 'notifiers.emailNotifier' }
# menuItems.notifiers.push { name: 'Webhooks', notifierType: 'webhooksNotifier', icon: 'icon-webhooks', sref: 'notifiers.webhooksNotifier' }

menuItems.trigger = {}
menuItems.trigger.gitTrigger = { title: 'Git Trigger', icon: 'icon-git', sref: 'triggers.git', editSref: 'triggers.git.edit' }
menuItems.trigger.schedule = { title: 'Schedule', icon: 'icon-clock', sref: 'triggers.schedule' }
menuItems.trigger.webhook = { title: 'Webhook', icon: 'icon-webhooks', sref: 'triggers.webhook' }
menuItems.trigger.task = { title: 'Task Trigger', icon: 'icon-tasks', sref: 'triggers.task' }

menuItems.triggers = [
  menuItems.trigger.gitTrigger
  menuItems.trigger.schedule
  menuItems.trigger.webhook
  menuItems.trigger.task
]

menuItems.taskEdit = [
  { title: 'Triggers', subtitle: 'Automaticly run this task', icon: 'icon-triggers', sref: 'task.edit.triggers' }
  { title: 'Notification Rules', subtitle: 'Toggle on or off notifications', icon: 'icon-output', sref: 'task.edit.notification-rules' }
  { icon: 'icon-error',       title: "Delete Task",        subtitle: "Danger Zone",   sref: 'task.edit.archive' }
]

menuItems.projectEdit = [
  {icon: 'icon-info', title: "Details",  subtitle: "Project name, etc",           sref: "projects/edit.details"},
  {icon: 'icon-repositories', title: "Repositories",  subtitle: "Manage Git repositories",           sref: "repositories"},
  {icon: 'icon-triggers',     title: "Triggers",      subtitle: "Webhooks & changes in your code",   sref: "triggers"},
  {icon: 'icon-environments', title: "Environments",  subtitle: "Configuration, keys & secrets",     sref: "environments"},
  {icon: 'icon-tasks',        title: "Scripts",       subtitle: "Scripts and other task definitions", sref: "scripts"},
  {icon: 'icon-output',       title: "Notifiers",     subtitle: "Notify people, services or tasks",  sref: "notifiers"},
  {icon: 'icon-user',         title: "People",        subtitle: "Manage team members",               sref: "projects/edit.people"},
  {icon: 'icon-error',       title: "Delete Project",        subtitle: "Danger Zone",               sref: "projects/edit.archive"},

]

menuItems.organizationEdit = [
  {icon: 'icon-info', title: 'Details', subtitle: 'Organization name, etc', sref: 'organization.edit.details'}
  # {icon: 'icon-user', title: 'People', subtitle: "Manage team members", sref: 'organization.edit.people'}
  {icon: 'icon-star', title: 'Billing', subtitle: "Organization Billing", sref: 'billing'}
  {icon: 'icon-error', title: 'Delete Organization', subtitle: "Danger Zone", sref: 'organization.edit.archive'}
]

menuItems.wizard = [
  {icon: 'icon-complete',     title: '1. Setup your account',   subtitle: 'Create Organization & Project',                  sref: 'wizard.create'}
  {icon: 'icon-repositories', title: '2. Connect your code',    subtitle: 'Add a repository to your project',               sref: 'wizard.project.connect'}
  {icon: 'icon-tasks',        title: '3. Choose your defaults', subtitle: 'Add a environments and scripts to your project', sref: 'wizard.project.stencils'}
]

menuItems.settings = [
  {icon: null, title: "Personal Details",    subtitle: "Name, email, password,  etc",          sref: "settings"},
  {icon: null, title: "Project Memberships", subtitle: "Manage your access to other projects", sref: "settings.project-memberships"},
  {icon: null, title: "Hints & Tips",        subtitle: "Reset systemwide prompts",             sref: "settings.prompts"},
  {icon: null, title: "Multi-Factor Auth",   subtitle: "Enhance your account security",        sref: "settings.mfa"},
  {icon: null, title: "OAuth",               subtitle: "Connect with Github",                  sref: "settings.oauth", feature: 'oauth.github.authentication'},
]

angular.module('harrowApp').constant 'menuItems', menuItems
