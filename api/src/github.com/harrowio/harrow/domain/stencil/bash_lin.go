package stencil

import (
	"fmt"

	"github.com/harrowio/harrow/domain"
)

// BashLinux sets up defaults for explaining how the user can interract with
// Harrow unconnected to a single language or framework.
type BashLinux struct {
	conf *Configuration
}

func NewBashLinux(configuration *Configuration) *BashLinux {
	return &BashLinux{
		conf: configuration,
	}
}

// Create creates the following objects for this stencil:
//
// Environments: demo, staging, production
//
// Tasks (in test): "....", ".....", "....."
//
// Tasks (in staging, production): "....."
func (self *BashLinux) Create() error {
	errors := NewError()
	project, err := self.conf.Projects.FindProject(self.conf.ProjectUuid)
	if err != nil {
		return errors.Add("FindProject", project, err)
	}

	sandboxEnvironment := project.NewEnvironment("Sandbox")
	if err := self.conf.Environments.CreateEnvironment(sandboxEnvironment); err != nil {
		errors.Add("CreateEnvironment", sandboxEnvironment, err)
	}

	stagingEnvironment := project.NewEnvironment("Staging").
		Set("REDIS_VERSION", "unstable")
	if err := self.conf.Environments.CreateEnvironment(stagingEnvironment); err != nil {
		errors.Add("CreateEnvironment", stagingEnvironment, err)
	}

	productionEnvironment := project.NewEnvironment("Production").
		Set("REDIS_VERSION", "stable").
		Set("DOCKERHUB_USER", "user").
		Set("DOCKERHUB_PASS", "Correct Horse Battery Staple").
		Set("DOCKERHUB_EMAIL", "user@example.org")
	if err := self.conf.Environments.CreateEnvironment(productionEnvironment); err != nil {
		errors.Add("CreateEnvironment", productionEnvironment, err)
	}

	buildAndPushDockerContainerTask := project.NewTask("Build & Push Docker Container", self.BuildAndPushDockerContainerTaskBody())
	if err := self.conf.Tasks.CreateTask(buildAndPushDockerContainerTask); err != nil {
		errors.Add("CreateTask", buildAndPushDockerContainerTask, err)
	}

	cowsayTask := project.NewTask("Cowsay Greeting", self.CowsayTaskBody())
	if err := self.conf.Tasks.CreateTask(cowsayTask); err != nil {
		errors.Add("CreateTask", cowsayTask, err)
	}

	deployFilesRsyncTask := project.NewTask("Deploy Files With Rsync", self.DeployFilesRsyncTaskBody())
	if err := self.conf.Tasks.CreateTask(deployFilesRsyncTask); err != nil {
		errors.Add("CreateTask", deployFilesRsyncTask, err)
	}

	downloadAndCompileRedisTask := project.NewTask("Download & Compile Redis", self.DownloadCompileRedisBody())
	if err := self.conf.Tasks.CreateTask(downloadAndCompileRedisTask); err != nil {
		errors.Add("CreateTask", downloadAndCompileRedisTask, err)
	}

	lintCodeTask := project.NewTask("Lint Scripts", self.LintScriptsTaskBody())
	if err := self.conf.Tasks.CreateTask(lintCodeTask); err != nil {
		errors.Add("CreateTask", lintCodeTask, err)
	}

	cowsayJob := &domain.Job{}
	lintCodeJob := &domain.Job{}
	rsyncDeployStagingJob := &domain.Job{}
	rsyncDeployProductionJob := &domain.Job{}
	buildAndPushDockerContainerProductionJob := &domain.Job{}
	downloadAndCompileRedisStagingJob := &domain.Job{}
	downloadAndCompileRedisProductionJob := &domain.Job{}
	jobsToCreate := []struct {
		Task        *domain.Task
		Environment *domain.Environment
		SaveAs      **domain.Job
	}{
		{cowsayTask, sandboxEnvironment, &cowsayJob},
		{lintCodeTask, stagingEnvironment, &lintCodeJob},
		{deployFilesRsyncTask, stagingEnvironment, &rsyncDeployStagingJob},
		{deployFilesRsyncTask, productionEnvironment, &rsyncDeployProductionJob},
		{buildAndPushDockerContainerTask, productionEnvironment, &buildAndPushDockerContainerProductionJob},
		{downloadAndCompileRedisTask, stagingEnvironment, &downloadAndCompileRedisStagingJob},
		{downloadAndCompileRedisTask, productionEnvironment, &downloadAndCompileRedisProductionJob},
	}

	for _, jobToCreate := range jobsToCreate {
		jobName := fmt.Sprintf("%s - %s", jobToCreate.Environment.Name, jobToCreate.Task.Name)
		job := project.NewJob(jobName, jobToCreate.Task.Uuid, jobToCreate.Environment.Uuid)
		if err := self.conf.Jobs.CreateJob(job); err != nil {
			errors.Add("CreateJob", job, err)
		}
		if jobToCreate.SaveAs != nil {
			*jobToCreate.SaveAs = job
		}
	}

	productionDeployKey := productionEnvironment.NewSecret("Deploy key", domain.SecretSsh)
	if err := self.conf.Secrets.CreateSecret(productionDeployKey); err != nil {
		errors.Add("CreateSecret", productionDeployKey, err)
	}

	return errors.ToError()
}

func (self *BashLinux) BuildAndPushDockerContainerTaskBody() string {
	return `#!/bin/bash -e

#
# The trick "${MY_VARAIBLE:?Need to set variable in environment}" ensures
# that we don't continue unless the variables are set properly in the
# environment. You can learn more at the Bash documentation:
#
#  * http://tldp.org/LDP/Bash-Beginners-Guide/html/sect_03_04.html
#
hfold "Signing Into Docker Hub"
sudo docker login --username="${DOCKERHUB_USER:?Variable is not set, check the environment}" \
                  --password="${DOCKERHUB_PASS:?Variable is not set, check the environment}" \
                  --email="${DOCKERHUB_EMAIL:?Variable is not set, check the environment}"
hfold --end

#
# Given a Docker file, from a repository or pulled from the internet, it
# is easy to build and deploy that, and to push the container to a remote
# container registry.
#
cat > Dockerfile <<-EOF
#
# A basic apache server. To use either add
# or bind mount content under /var/www
#
# Dockerfile from https://github.com/kstaken/dockerfile-examples
#
FROM ubuntu:12.04

MAINTAINER Kimbro Staken version: 0.1

RUN apt-get update && apt-get install -y apache2 && apt-get clean && rm -rf /var/lib/apt/lists/*

ENV APACHE_RUN_USER www-data
ENV APACHE_RUN_GROUP www-data
ENV APACHE_LOG_DIR /var/log/apache2

EXPOSE 80

CMD ["/usr/sbin/apache2", "-D", "FOREGROUND"]
EOF

#
# Dockerfile in the current directory so we can
#
hfold "Building apache2 Docker Container" sudo docker build -t apache2 .

#
# Docker will try and ask if we *really* want to push a container image
# to a public registry '--force=true' means we won't be asked.
#
hfold "Pushing Docker Image" docker push --force=true apache2

# We can print a pretty banner, if we like 'tput' lets
# us change the way the output is printed if the terminal
# supports the changes we ask for.
tput setab 2 # Set the background color to green
tput setaf 0 # Set the foreground color to white
tput bold   # With bold fonts
printf "╒═══════════════════════════════════════════════════════════════════╕\n"
printf "│ Success!                                                          │\n"
printf "╘═══════════════════════════════════════════════════════════════════╛\n"
echo
`
}

func (self *BashLinux) LintScriptsTaskBody() string {
	return `#!/bin/bash -e

#
# Harrow allows the use of sudo to install packages
# and libraries
#
hfold "Install ShellCheck" sudo apt-get install shellcheck

#
# Here we're counting the number of shell scripts in the
# repositories directory with a combination of 'find'
# and 'wc', incase we don't find any, we can print
# a warning to standard out and let people know what
# to do, and then stop the script.
#
numberOfShellScripts=$(find repositories/ -type f -name '*.sh' | wc -l)
if [ $numberOfShellScripts -eq 0 ]; then
		{
			printf "Looks like you have no shell scripts in your "
			printf "repositories/ directory, this may be because "
			printf "you haven't added any repositories, or there "
			printf "simply no shell scripts! \n"
			printf "Add a repository with shell scripts to experiment "
			printf "with this task!\n"
		} 1>&2
    exit 123
fi

find repositories/ -type f -name '*.sh'  | while read file; do
    printf "Shellchecking $file\n";
    if shellcheck "$file"; then
        printf "\tNo problems found with %s\n" $file
    fi
done
`
}

func (self *BashLinux) DownloadCompileRedisBody() string {
	return `#!/bin/bash -e

# Being able to compile and install cutting edge software for testing,
# or to maintain your own nightly binaries, perhaps distributing them
# via S3 can be a huge time saver.
#
# This script uses the simple build instructions for the Redis data structure
# store to compile and test either the stable or unstable version depending
# on the $REDIS_VERSION environment variable.

hfold "Downloading Redis"
if [ "$REDIS_VERSION" = "stable" ]; then
    wget http://download.redis.io/redis-stable.tar.gz
    tar -xvzf "redis-stable.tar.gz"
    redis_dir="redis-stable"
else
    wget https://github.com/antirez/redis/archive/unstable.tar.gz
    tar -xvzf "unstable.tar.gz"
    redis_dir="redis-unstable"
fi
hfold --end

# We can make it easy to follow which directory the script is
# working in by using a sub-shell for the part where we change
# the working directory
(
    cd "$redis_dir/"
    hfold "Compile Redis" make
    hfold "Install test dependencies" sudo apt-get -y install tk8.5 tcl8.5
    hfold "Test Redis" make test
)

# Congratulations, Redis compiled and installed to our VM
# now, we could easily deploy it, or build a package for
# it and deploy the package...

# We can print a pretty banner, if we like 'tput' lets
# us change the way the output is printed if the terminal
# supports the changes we ask for.
tput setab 2 # Set the background color to green
tput setaf 0 # Set the foreground color to white
tput bold   # With bold fonts
printf "╒═══════════════════════════════════════════════════════════════════╕\n"
printf "│ Success!                                                          │\n"
printf "╘═══════════════════════════════════════════════════════════════════╛\n"
echo
`
}

// CowsayTaskBody returns the body of the task for displaying a greeting
// using cowsay
func (self *BashLinux) CowsayTaskBody() string {
	return `#!/bin/bash -e

# Folds tell our log viewer to make collapsible regions. You can either wrap
# your commands in hfold blocks like this...
hfold "Greet the world"
printf "Hello, %s!\n" "World"
hfold --end

# ... or you can specify a single command to run as the second argument to
# hfold:
hfold "Install cowsay" sudo DEBIAN_FRONTEND=noninteractive apt-get install -y cowsay

/usr/games/cowsay -f tux "Greetings from the Harrow virtual machine!"
`
}

// DeployFilesRsyncTaskBody returns the body of the task for deploying
// a simple directory of files using Rsync.
func (self *BashLinux) DeployFilesRsyncTaskBody() string {
	return `#!/bin/bash -e

#
# Check early that the required environment variables are set
# and print a warning if not
#
{$RSYNC_HOST_PORT:?Don't know which host:port to use, check the environment}

#
# Harrow doesn't trust unknown SSH hosts by default, this is a way
# to quickly bypass an annoying error! The ${...%\;*} construct splits
# the variable on : returning the first half
#
ssh-keyscan -4 ${RSYNC_HOST_PORT%\:*} ~/.ssh/known_hosts

#
# A better way to trust ah SSH host is to verify the key(s) manually
# and add them to the environment, and add a variable to your environment
# with the host key.
#
# Better yet is to use tools which allow you to append configuration
# such as trusted known hosts to the system-wide settings by adding
# them to the project-local configuration as part of your repository.
#
# Capistrano and SSHKit are examples of software that work this way.
#

#
# Using a sub-shell when changing directory is a nice way to visually
# group related commands, whilst making sure it's easy to follow the
# flow of the program.
#
(
	cd repositories/"{$REPO_NAME:?Don't know which repository to deploy, check the environment}"

	#
	# When deploying anything out of a Git repository it's critical that you
	# do *not* deploy the Git directory itself. The Git directory would be a
	# huge risk for security if you were to make it available on the web!
	#
	# Excluding that directory using rsync is trivial.
	#
	rsync -a --exclude='.git/' ./* "$RSYNC_HOST_PORT":/var/www/your-project/
)
`
}
