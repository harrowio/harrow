Github
======

Github webhooks can be configured to fire on a variety of events such as new
commits, tags, branches, pull request, etc.

When the webhook body refers to a commit, Harrow will pre-checkout the
repository (or, repositories) at the correct, matching branch.

Given the example that you have a frontend, and backend repository both added
to Harrow in a given project, and you wish to trigger integration tests on the
combined codebases when new branches appear in either one.

One could simply ensure that branch naming is consistent betweeen the
repositories, e.g. ``1234-our-new-feature-branch``, then attaching the webhook
at Github, and configuring it to be sent on all new branch events, when running
your operation on Harrow the correct feature branch would be checked out in
both repositories (although the webhook fired only in one of them).
