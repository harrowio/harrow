app = angular.module("harrowApp")

app.constant "en_GB",
  misc:
    verbs:
      edit: "Edit"
    properties:
      private: "Private"
      public: "Public"
    duration:
      withHours: "{{hours}}hours {{minutes}}min {{seconds}}sec"
      withMinutes: "{{minutes}}m{{seconds}}s"
      withSeconds: "{{seconds}}s"
      empty: "immediately"
    defaultPlanUuid: "b99a21cc-b108-466e-aa4d-bde10ebbe1f3"
  moreCommits: '{count, plural, one{and 1 more commit} other{and # more commits}}'
  intercom:
    blocked: "It appears that you have an ad blocker enabled. We respect your decision, but one of our support tools is also blocked by it."
  formattedPricingPlan:
    planName: "{planUuid, select, b99a21cc-b108-466e-aa4d-bde10ebbe1f3{Free} c9f45ff0-e22f-4170-9abc-5a09d400dcb2{Basic} 1b67ab5d-9fff-46a3-89ac-113fbd360790{Small Team} f975a385-3625-4883-b353-8f1febeb5b3e{Big Team} other{{name}}}"
    extras: "{projects, plural, one{1 Project} other{# Projects}} to use, and {users, plural, one{1 User} other{# Users}} to invite. Reason: {reason}."
    price: "{{price | currency:'$':0 }}"
    users: "{users, plural, one{1 User} other{# Users}}"
    projects: "{projects, plural, one{1 Project} other{# Projects}}"
    unlimitedProjects: "Unlimited Projects"
    additionalUsers: "Additional users: {{price | currency:'$':0}}"
    performance: "{tasks, plural, one{Basic Performance} two{Good Performance} other{Great Performance }}"
    platform: "Cloud Platform"
  wsErrorIndicator:
    disconnected: "Disconnected"
    reconnect: "Reconnect"
  errors:
    serverError: "Oops, it looks like a Server Error has occured. Do not worry, our team is on the case, if you'd like you can open support and chat with us."
    field:
      required: "Required"
      invalid_totp_token: "Must be a number with six digits"
      too_short: "Too short"
      email: "Must be an email address"
      minlength: "Too short"
      maxlength: "Too long"
      number: "Must be a number"
      unique_violation: "Already in use"
      match: "Does not match"
    http403:
      header: "Permission denied"
      description: "You do not have permission to access this page."
      hint: "Try contacting the manager of the organization or project to get access."
    http403_session_invalidated:
      header: "Session invalidated"
      description: "Another login for your account invalidated this session"
      hint: "Upgrade your subscription to invite collaborators to your project."
      solution: "Log in again"
    http404:
      header: "Not Found"
      description: "The page you were trying to load could not be found."
      hint: "If you are certain this page exists, it might be that you do not have permission to access it. If it exists at all."
    http418:
      header: "I'm a teapot"
      description: "The server is acting like a tea pot. This means we have an exceptional internal condition."
      hint: "Sorry, there's nothing you can do except to try again, and contact us if it doesn't improve. Rest assured our monitoring alerts are scrambling Thunderbird 2 now."
    http502:
      header: "Bad Gateway"
      description: "We can't reach the backend right now, please try again shortly."
      hint: "Please try again shortly, this error means Harrow has an internal error which we are working to fix."
    http500:
      header: "Internal error"
      description: "Something went wrong on our side."
      hint:  "Sorry, there's nothing you can do except to try again, and contact us if it doesn't improve. Rest assured our monitoring alerts are scrambling Thunderbird 2 now."
    blocked:
      header: "Please verify your email address!"
      description: """
      We sent an email with an account confirmation link when you created your account. Please check your spam folder, or contact us at team@harrow.io if you need any assistance.
      """
      resendVerificationEmail: "Click here if you haven't received the email."
      hint: """
      Every new account is allowed a 12 hour grace period before we require the email address associated with an account to be verified.
      """
      flashes:
        failedToResendVerificationEmail: """
        Failed to resend verification email.
        """
    verification_email_sent:
      header: "Verfication email sent!"
      description: """
      We have sent an email with a link for verifying your account to
      this email address:
      """
    github_existing_unlinked_user:
      header: "GitHub login failed"
      description: """
      We couldn't match the login information returned by GitHub to any in our database. However, it looks like you already have a Harrow account.
      """
      hint: """
      To link your account to GitHub, please log in to Harrow with username and password and add an OAuth connection in the user settings.
      If you don't have a password, you can use the password reset form to set one.
      """
    back_to_safety: "Back to safety"
  membershipTypes:
    guest: "Guest"
    member: "Member"
    manager: "Manager"
    owner: "Owner"
  navigation:
    nav:
      organizations: "Organisations"
      projects: "Projects"
      settings: "Your Settings"
    user:
      logIn: "Log In"
      logOut: "Log Out"
      signUp: "Sign Up"
  repositoryList:
    import: "Import from GitHub"
    connectGithub: "You can easily import repositories if you connect Harrow with GitHub using OAuth, "
    oauthSettings: "connect now."
  repositoryImport:
    title: "Import Repositories from GitHub"
    text: "Select a GitHub repository for import into Harrow"
    noneFound: "We could not find any repositories in the connected GitHub account"
    oauthSettings: "Please check your OAuth settings"
    alreadyImported: "Repository imported"
    done: "Back to your project"
    import: "Import"
  dashboard:
    pane:
      0:
        body: "To automatically import a list of projects for you to choose from"
        title: "Connect with a Git provider"
      1:
        C2A: "It\"s dangerous to go alone"
        body: "Organizations are the places we put your colleagues, repositories, environments, servers, tasks and logs, you\"ll be needing one of these."
        title: "Create an organization"
      2:
        C2A: "Party on..."
        body: "Coding is better with friends, or colleagues, at least. Invite them to grant them access to your projects."
        title: "Invite your collaborators"
    title: "Your Dashboard"
    projects: "{count, plural, one{1 active project} other{# active projects}}"
    organizations: "Your Organizations"
    scripts: "{count, plural, one{1 active script} other{# active scripts}}"
    createOrganization: "Create New Organization"
    title: "Your Dashboard"
    recentHumanActivity: "Recent Activity"
  scripts:
    otherEnvironments: "{count, plural, one{1 other environment} other{# other environments}}"
  footer:
    copyright: "© {{YEAR}} Harrow Inc."
  orgSetup:
    defineScripts:
      submit: "Use This Script"
      placeholder:
        name: "Hello World"
      label:
        body: "Script Body"
        description: "Description"
        name: "Script Name"
      flashes:
        success: "Script {{name}} saved"
        fail: "Unable to create Script {{name}}"
  forms:
    login:
      submit: "Log in"
      label:
        email: "Email"
        password: "Password"
      placeholder:
        email: "Your Email"
        password: "Your Password"
      flashes:
        loginFailed: "Email address or password was not valid"
      forgotPassword: "Forgot password?"
      github: "Log in with GitHub*"
      noAccount:
        callToActionButtonTitle: 'Create an account'
    setUpAccount:
      submit: "Set up account"
      placeholder:
        organizationName: "Acme Inc."
        projectName: "Widgets"
      label:
        projectName: "Project"
        organizationName: "Organization"
      flashes:
        success: "Account {{organization.subject.name}}/{{project.subject.name}} set up"
        fail: "Unable to set up account {{organization.subject.name}}/{{project.subject.name}}"
    addPrivateRepository:
      submit: "Connect"
      placeholder:
        protocol: "https"
        host: "github.com"
        port: "443"
        path: "/capistrano/capistrano.git"
        username: "john-doe"
        password: "••••••••••"
      label:
        protocol: "Protocol"
        host: "Host"
        port: "Port"
        path: "Path"
        username: "Username"
        password: "Password"
    addRepository:
      submit: "Add repository"
      label:
        repositoryUrl: "Repository URL"
      placeholder:
        repositoryUrl: "https://github.com/capistrano/capistrano.git"
      hint:
        repositoryUrl: "If you're connecting to a private repository we'll ask you for your credentials next."
    sessionConfirmation:
      submit: "Confirm"
      header: "TOTP code required"
      explanation: "TOTP based multi-factor authentication is activated for your account.  Please enter your current TOTP code to authenticate and complete the login process."
      label:
        totp: "TOTP Code"
      placeholder:
        totp: "123456"
      flashes:
        confirmationFailed: "TOTP code verification failed"
    forgotPassword:
      submit: "Reset password..."
      label:
        email: "Email"
      placeholder:
        email: "Your Email"
      flashes:
        success: "Password reset link has been sent"
        fail: "Unable to send password reset link"
    verifyEmail:
      flashes:
        success: "Email address verified"
        fail: "Unable to verify email address"
    resetPassword:
      submit: "Reset password..."
      errors:
        passwordConfirmation:
          match: "Passwords do not match"
      label:
        email: "Email"
        password: "Your new password (at least 10 characters)"
        passwordConfirmation: "Please confirm your password"
      placeholder:
        email: "Your Email"
        password: "••••••••••"
        passwordConfirmation: "••••••••••"
      flashes:
        success: "Password successfully reset"
        fail: "Unable to reset password"
    enableTotp:
      submit: "Confirm"
      secret: "Your TOTP secret token: "
      header: "Enable TOTP multi-factor authentication"
      explanation: "Scan this QR code or enter the provided secret token manually into your TOTP client.  Enter the code provided by your TOTP client to enable TOTP authentication for this account.  All future logins will require such a code."
      label:
        totp: "TOTP Code"
        totpToken: "TOTP Code"
      placeholder:
        totpToken: "123456"
      flashes:
        confirmationFailed: "You must enter one valid TOTP code to enable TOTP."
    disableTotp:
      submit: "Confirm"
      header: "Disabling TOTP"
      warning: "You are about to disable TOTP multi-factor authentication. Please enter your current TOTP code to authorize this action."
      label:
        totp: "TOTP Code"
      placeholder:
        totp: "123456"
      flashes:
        confirmationFailed: "TOTP code verification failed"
    invitationForm:
      title: "Invite To Project"
      recipientName: "Recipient Name"
      message: "Message"
      email: "Email"
      flashes:
        success: "Invited {{recipientName}}"
        fail: "Failed to invite {{recipientName}}"
      label:
        email: "Email"
        message: "Message"
        recipientName: "Recipient Name"
      placeholder:
        email: "john.doe@example.com"
        recipientName: "John Doe"
        message: "Welcome aboard!"
      submit: "Send invitation"
    environmentForm:
      submit: "Save Environment"
      placeholder:
        name: "e.g \"Development\""
      label:
        name: "Environment Name"
      flashes:
        success: "Environment {{name}} saved"
        fail: "Unable to save Environment {{name}}"
    variableForm:
      submit: "Save Variable"
      placeholder:
        name: "e.g MY_ENV"
        value: "e.g development"
      label:
        name: "Name"
        value: "Value"
    envForm:
      submit: "Add new Secret"
      placeholder:
        name: "e.g AWS_ACCESS_KEY"
        value: "AKIAAAAAABBBB"
      label:
        name: "e.g Name"
        value: "Value"
      flashes:
        success: "Added Secret {{name}}"
        fail: "Unable to add Secret {{name}}"
    sshForm:
      submit: "Generate SSH key"
      placeholder:
        name: "e.g staging cluster key"
      label:
        name: "Generate SSH Key"
      flashes:
        success: "Enqueued key {{name}} for generation"
        fail: "Unable to generate key {{name}}"
    taskForm:
      submit: "Save Task"
      label:
        name: "Task Name"
        description: "Task Description"
        environmentUuid: "Environment"
        scriptUuid: "Script"
      placeholder:
        description: "Tell your colleagues what this task is about (optional)"
      flashes:
        success: "Task saved"
        fail: "Unable to save task"
        delete:
          success: "Task has been deleted"
          fail: "Unable to delete Task"
    script:
      placeholder:
        name: "Hello World"
        projectUuid: "Project"
      label:
        body: "Script Body"
        name: "Script Name"
        projectUuid: "Project"
        environmentUuid: "Environment"
      flashes:
        create:
          success: "Script {{name}} saved"
          fail: "Unable to save Script {{name}}"
        delete:
          success: "Script {{name}} has been deleted"
          fail: "Unable to delete Script {{name}}"
    organizationForm:
      submit: "Save Organisation"
      errors:
        name:
          required: "An organisation name is required"
      label:
        name: "Name"
        public: "Privacy"
      placeholder:
        name: "Your Organization Name (e.g Twitter. Inc)"
      flashes:
        success: "Organization {{name}} saved"
        fail: "Unable to save Organization {{name}}"
        delete:
          success: 'Organisation {{name}} deleted'
          fail: 'Unable to delete Organization {{name}}'
    billing:
      submit: "Subscribe"
      prompt:
        trialWarning: 'You have {count, plural, one{1 day} other{# days}} remaining of your trial, please upgrade now to conitinue enjoying Harrow.'
        trialEnded: 'Your trial has expired, please upgrade now to continue using harrow.'
        upgrade: "Are you sure you want to upgrade to {{pricePerMonth}}?"
        downgrade: "Are you sure you want to downgrade to {{pricePerMonth}}?"
      label:
        number: "Credit Card Number"
        cvv: "Verification Code"
        expirationMonth: "Month (MM)"
        expirationYear: "Year (YYYY)"
        cardholderName: "Cardholder Name"
      button:
        planChangeDirection: '{isUpgrade, select, true{Upgrade} other{Downgrade}}'
      flashes:
        success: "Billing plan changed successfully"
        processing: "Processing..."
    userSettings:
      submit: "Save Personal Data"
      passwordHint: "Please enter your current password to make any changes to your personal details."
      noProjectMemberships: "You don't have any project memberships."
      noticeAboutOrganizations: "You can’t leave <em>Organizations</em> that you are a member of.  This page only lists projects you have been invited to."
      noticeAboutPrompts: "The prompts and tips displayed throughout the site can be restored by clicking the button below."
      errors:
        passwordConfirmation:
          match: "Passwords do not match"
      label:
        name: "Name"
        email: "Email Address"
        password: "Password"
        newPassword: "New Password"
        passwordConfirmation: "New Password (again)"
      placeholder:
        name: "Your Full Name"
        email: "Your Email Address"
        password: "Your Current Password"
        newPassword: "Your New Password"
        passwordConfirmation: "Confirm your New Password"
      flashes:
        success: "Personal data saved"
        failure: "Failed to save personal data"
      withoutPassword: "You are currently logged in via OAuth. If you want to give yourself a password, please use the "
      withoutPasswordLink: "password reset form."
    project:
      errors:
        name:
          required: "A Project name is required"
        organizationUuid:
          required: "Please select an organisation"
      label:
        name: "Name"
        public: "Make project public"
      placeholder:
        name: "e.g. Meep Meep App"
      flashes:
        create:
          success: "Project {{name}} saved"
          fail: "Unable to save Project {{name}}"
        delete:
          success: "Project {{name}} deleted successfully"
          fail: "Unable to delete Project {{name}}"
        deletion:
          success: "Project {{name}} deleted successfully"
          failure: "Unable to delete Project {{name}}"

    scheduleForm:
      submit: "Save Schedule"
      label:
        type: "Type of Schedule"
        timespec: "Timespec"
        cronspec: "Cronspec"
        description: "Description"
      placeholder:
        timespec: "now + 5 minutes"
        cronspec: "0/5 * * * *"
      errors:
        timespec:
          invalid: "Invalid timespec, should look like \"now + 5 minutes\""
        cronspec:
          invalid: "Invalid cron expression, should look like \"0/5 * * * *\""
      flashes:
        success: "Task scheduled"
        fail: "Unable to create Schedule"
    repositoryForm:
      submit: "Save Repository"
      label:
        name: "Name"
        projectUuid: "Project"
        protocol: "Protocol"
        host: "Host"
        url: "URL"
        port: "Port"
        path: "Path"
        username: "Username"
        password: "Password or OAuth token"
      placeholder:
        name: "e.g. Backend, Frontend, Blog, Chef Recipes, etc"
        url: "e.g. git@github.com:rails/rails.git"
        host: "github.com"
        path: "capistrano/documentation.git"
      flashes:
        saved: "Repository {{name}} saved"
        deleted: "Repository {{name}} deleted"
        failure: "Unable to save Repository {{name}}"
    repository:
      flashes:
        save:
          success: 'Repository {{name}} saved'
          fail: 'Unable to save Repository {{name}}'
        delete:
          success: 'Repository {{name}} deleted'
          fail: 'Unable to delete Repository {{name}}'
    signup:
      submit: "Sign Up!"
      secondary: "I already have an account"
      errors:
        email:
          email: "Must be an email address, e.g user@mailhost.tld"
          required: "Email Required"
        password:
          required: "Password Required"
        passwordConfirmation:
          match: "Passwords do not match"
        acceptTermsAndConditions:
          required: "Please accept our terms and conditions"
      label:
        email: "Your Email"
        name: "Your Name"
        password: "Your Password"
        passwordConfirmation: "Please Confirm Your Password"
        acceptTermsAndConditions: "I agree with the <a href=\"https://harrow.io/terms-and-conditions/\" target=\"_blank\">Terms of Service</a>."
      placeholder:
        email: "Email Address"
        name: "Name"
        password: "Password"
        passwordConfirmation: "Confirm password"
      title: "Get started with the worlds most flexible CI tool..."
      flashes:
        success: "Your account was created successfully!"
        fail: "Signup failed, sorry, please check the form for errors!"
    gitTriggerForm:
      submit: "Save Git Trigger"
      label:
        name: "Name"
        triggerKind: "When a branch or tag"
        triggerRepository: "In repository"
        triggerMatch: "And the branch or tag name matches"
        taskUuid: "Trigger this task"
      flashes:
        success: "Git trigger saved successfully!"
        fail: "Failed to save Git trigger!"
      placeholder:
        name: "e.g \"Run tests on new tags\""
    webhookForm:
      submit: "Save Webhook"
      placeholder:
        name: "GitHub"
        slug: "commits"
        projectUuid: "Project"
      label:
        name: "Name"
        slug: "URL slug"
        projectUuid: "Project"
        taskUuid: "Task to trigger"
      flashes:
        success: "Webhook {{name}} saved"
        fail: "Unable to save Webhook {{name}}"
    wizard:
      create:
        label:
          organizationName: "Organization Name",
          projectName: "Project Name"
        placeholder:
          organizationName: "e.g. ACME Corp"
          projectName: "e.g. Meep Meep App"
        flashes:
          success: "Project {{name}} saved"
          fail: "Unable to save Project {{name}}"
      connect:
        label:
          url: "Repository URL"
          protocol: "Protocol"
          host: "Host"
          port: "Port"
          path: "Path"
          username: "Username"
          password: "Password or OAuth Token"
        placeholder:
          url: "e.g. git@github.com:capistrano/capistrano.git"
          protocol: "https"
          host: "github.com"
          port: "443"
          path: "/capistrano/capistrano.git"
          username: "git"
        flashes:
          success: "Repository saved"
          fail: "Unable to save Repository"
      stencils:
        flashes:
          success: "Stencil saved"
          fail: "Unable to save Stencil"
      users:
        label:
          email: "Email Address"
          name: "Name"
        placeholder:
          email: "e.g. joe.bloggs@example.com"
          name: "e.g. Joe Bloggs"
        flashes:
          success: "{count, plural, one{You've invited 1 member} other{You've invited # members}}"
          fail: "Unable to invite members"
      finished:
        header: 'Success! You&rsquo;ve created your {projects, plural, one{first} two{second} other{{ordinal}}} project'
    notifiers:
      slackNotifier:
        buttons:
          new: 'Add Slack Channel'
        legend: "Slack"
        collectionLegend: "{count, plural, one{1 Slack Channel} other{# Slack Channels}}"
        label:
          webhookUrl: 'Webhook URL'
          name: 'Post To Channel'
        placeholder:
          webhookUrl: "e.g. https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
          name: "e.g. #Channel"
        flashes:
          create:
            success: "Saved Slack Channel"
            fail: "Unable to save Slack Channel"
          delete:
            success: "Remove Slack Channel"
            fail: "Unable to remove Slack Channel"
      emailNotifier:
        buttons:
          new: 'Add Email Notifier'
        legend: "Email"
        collectionLegend: "{count, plural, one{1 Email Notifier} other{# Email Notifiers}}"
        label:
          recipient: "Recipient"
        placeholder:
          recipient: 'e.g. test@example.com'
        flashes:
          create:
            success: 'Saved Email Notifier'
            fail: 'Unable to save Email Notifier'
          delete:
            success: 'Removed Email Notifier'
            fail: 'Unable to remove Email Notifier'
      taskNotifier:
        buttons:
          new: 'Add Task Notifier'
        legend: 'Task Notifier'
        collectionLegend: "{count, plural, one{1 Task Notifier} other{# Task Notifiers}}"
        label:
          taskUuid: 'Task to Run'
          triggeredByTaskUuid: 'This Task'
          triggerAction: 'Run task When'
        options:
          triggerAction:
            operation:
              '*': "Success, Failure, and anything else"
              succeeded: 'Success'
              failed: 'Failure'
        flashes:
          create:
            success: 'Saved Task Notifier'
            fail: 'Unable to save Task Notifier'
          delete:
            success: 'Removed Task Notifier'
            fail: 'Unable to remove Task Notifier'
    tasks:
      notifierRules:
        label:
          operation:
            succeeded: "Success"
            failed: "Failure"
        flashes:
          success: "Saved Notification Rule"
          fail: "Unable to save Notification Rule"
    triggers:
      gitTrigger:
        button:
          new: "Add Git Trigger"
        legend: 'Git Trigger'
        collectionLegend: '{count, plural, one{1 Git Trigger} other{# Git Triggers}}'
        icon: 'icon-git'
        label:
          name: 'Trigger Name'
          changeType: 'Trigger when'
          matchRef: 'Branch or Tag'
          repositoryUuid: 'Repository'
          taskUuid: 'Run Task'
        placeholder:
          name: 'e.g. "Run tests on new tags"'
        options:
          changeType:
            change: 'Changed'
            add: 'Added'
            remove: 'Removed'
        flashes:
          success: 'Saved Git Trigger'
          fail: 'Unable to save Git Trigger'
      schedule:
        button:
          new: "Add Schedule"
        legend: 'Schedule'
        collectionLegend: '{count, plural, one{1 Schedule} other{# Schedules}}'
        icon: 'icon-clock'
        label:
          toggleTimespec: 'One-time Schedule'
          toggleCronspec: 'Recurring Schedule'
          cronspec: 'Cronspec'
          timespec: 'Timespec'
          description: 'Description'
          taskUuid: 'Run Task'
        placeholder:
          timespec: 'now + 5 minutes'
          cronspec: '0/5 * * * *'
          description: 'e.g. Run every 5 minutes'
        flashes:
          success: 'Saved Schedule'
          fail: 'Unable to save Schedule'
          delete:
            success: "Deleted Schedule"
            fail: "Unable to delete Schedule"
      webhook:
        button:
          new: "Add Webhook"
        legend: 'Webhook'
        collectionLegend: '{count, plural, one{1 Webhook} other{# Webhooks}}'
        icon: 'icon-webhooks'
        label:
          name: 'Webhook Name'
          taskUuid: 'Task to run'
        placeholder:
          name: 'e.g. Run Meep Meep App'
        flashes:
          success: 'Saved Webhook'
          fail: 'Unable to save Webhook'
      taskNotifier:
        button:
          new: "Add Task Trigger"
        legend: "Task Trigger"
        collectionLegend: '{count, plural, one{1 Task Notifier} other{# Task Notifiers}}'
        icon: 'icon-tasks'
        label:
          name: "Task Name"
          taskUuid: "Task to Run"
        flashes:
          success: "Saved Task Trigger"
          fail: "Unable to save Task Trigger"

  repositories:
    addDeployKey:
      title: "Add the Harrow Deploy Key to your GitHub Repository"
      text: "Oops! It looks like we couldn\"t reach your GitHub Repository. We might be able to fix that by adding the Harrow Deploy Key to it. It will show up as deploy@harrow.io, and can be removed at any time."
      button: "Do it"
      success: "Deploy Key added"
      fail: "Unable to add Deploy Key"
  oauth:
    githubCallback:
      error: "There was an error connecting to GitHub"
      title: "GitHub is processing your request"
      text: "Hang on, we are waiting for GitHub to process the request"
    github:
      organizations:
        title: "Your GitHub Organisations And Their Repositories"
      selectProject:
        title: "Select Project"
        desc: "Please select the Project you want to import Repositories into."
        filter: "Filter Projects"
        flash: "Project {{name}} selected"
      importGithubRepository:
        success: "Repository {{name}} successfully imported"
        fail: "Unable to import Repository {{name}}"
  invitations:
    show:
      accept: "Accept"
      refuse: "Refuse"
      description: "wants to invite you to"
      flashes:
        accept:
          fail: "Failed to accept the invitation."
          success: "Invitation accepted"
        refuse:
          fail: "Failed to refuse the invitation."
          success: "Invitation refused"
  operations:
    flashes:
      cancel:
        success: "Operation canceled"
        error: "Failed to cancel operation"
    status:
      success: "Succeeded"
      pending: "Pending"
      failed: "Failed"
      timedout: "Timed out"
      canceled: "Canceled"
  organizations:
    index:
      create: "Create Organisation"
      title: "Your Organisations"
      feed: "Activity Feed"
      schedule: "Schedule"
      dashboards: "Dashboards"
      tasks: "Tasks"
    show:
      createProject: "Create New Project"
      cancel: "Cancel"
      noMembers: "No Members"
      noProjects: "No Projects"
      editProject: "Edit"
      deleteProject: "Delete"
      delete: "Delete Organization"
      edit: "Edit Organization"
    flashes:
      deletion: "Organization {{name}} deleted"
      deletionFail: "Unable to delete Organization {{name}}"
  environmentsCreate:
    environmentPropertiesHint: "You can add variables, secrets and SSH keys after naming the environment"
  tasks:
    flashes:
      create:
        success: "Task created for {{environment.subject.name}}"
        fail: "Unable to create Task"
      runNow:
        success: "Task scheduled to run \"now\""
        fail: "Unable to schedule task"
      delete:
        success: "Task deleted"
        fail: "Unable to delete task"
  userSettings:
    title: "Your Settings"
    personalDetails:
      title: "Personal Information"
    TFA:
      title: "Two Factor Authentication"
    oauth:
      githubConnect: "Connect with GitHub"
      githubChecking: "Checking GitHub connectivity …"
      githubConnected: "✓ Connected"
  webhooks:
    prompts:
      regenerate: "This will generate a new URL for this webhook.  Do you want to proceed?"
    flashes:
      delete:
        success: "Webhook {{name}} deleted"
        failure: "Failed to delete Webhook {{name}}"
  gitTriggers:
    flashes:
      delete:
        success: "Git trigger {{name}} deleted"
        failure: "Failed to delete Git trigger {{name}}"
  prompts:
    "really?": "Are you sure?"
  projects:
    flashes:
      deletion: "Project {{name}} deleted"
      deletionFail: "Unable to delete project {{name}}"
  projectMemberships:
    prompts:
      promote: "Do you want to promote {{name}} to {{type}}?"
      remove: "Do you want to remove {{name}} from this project?"
  environmentList:
    flashes:
      deletion:
        success: "Environment {{name}} deleted successfully"
        failure: "Failed to delete Environment {{name}}"
  scriptList:
    selectEnvironment: "Please select an Environment to run your Script in:"
    noEnvironments: "You must have at least one Environment to run the Script in."
