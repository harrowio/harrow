# Contributing to Harrow

* Use [Stack Overflow][so] for Capistrano-related how-to questions and support
* [**Don't** push your pull request](http://www.igvita.com/2011/12/19/dont-push-your-pull-requests/)
* [**Do** write a good commit message](http://365git.tumblr.com/post/3308646748/writing-git-commit-messages)

## Getting help

Most people will prefer to use the coud-hosted version of Harrow at harrow.io.

Paid plans include support which extends to debugging your environment and
setup.

If you're a paid user of harrow.io and you want to report an issue or send us a
patch please let us know when talking to us, it will help us prioritise the
issues we're dealing with.

If you're a paid user of Harrow.io please talk to us directly using the support
channels embedded within the web application.

## Tests

Most of Harrow is fairly well tested. Use of type-safe languages on the backend
(Go) eliminate the need for a whole class of unit tests, but we're unlikey to
be able to take a patch without corresponding tests.

Details of how to run tests are included in the README documents of the
respective component sub-directories.

## Coding guidelines

The backend is written in Go which includes a canonical formatter. Please make
sure to run the formatter and check for redundant imports before preparing your
PR. If you've changed the dependencies please be mindful to include the
relevant lockfile changes for the package manager.

The front-end, style-guide, configuration management etc don't have fixed
coding guidelines. Patches to introduce canonicalizing linters for these
sub-projects would be gladly accepted. In lieu of that, "whatever vim does" is
the expected format of files.

## Changelog

The changelog is maintained in Git notes (https://git-scm.com/docs/git-notes).
Please refer to the post at
https://harrow.io/blog/effortlessly-maintain-a-high-quality-change-log-with-little-known-git-tricks/
for more information.

GitHub doesn't support PRs for tags/notes and non-tree refs, so unfortunately
we can't make this part of the PR process, one of the core maintainers will
write the changelog entry if one is required.

## Breaking changes

We adhere to [semver](http://semver.org/) so breaking changes will require a
major release.  For breaking changes, it would normally be helpful to discuss
them by raising a 'Proposal' issue or PR with examples of the new API you're
proposing [before doing a lot of
work](https://www.igvita.com/2011/12/19/dont-push-your-pull-requests/).  Bear
in mind that breaking changes may require many hundreds / thousands of users to
update their code.
