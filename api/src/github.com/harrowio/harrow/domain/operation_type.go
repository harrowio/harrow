package domain

const (
	OperationTypeGitAccessCheck         string = "git.check.access"
	OperationTypeGitEnumerationBranches string = "git.enumeration.branches"
	OperationTypeGitEnumerationCommits  string = "git.enumeration.commits"
	OperationTypeJobScheduled           string = "job.scheduled"
	OperationTypeJobWebhooked           string = "job.webhooked"
	OperationTypeJobRetried             string = "job.retried"
	OperationTypeNotifierInvoke         string = "notifier.invoke"
)
