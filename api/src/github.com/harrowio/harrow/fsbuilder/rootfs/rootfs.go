package rootfs

//go:generate go-bindata -o templates.go -pkg rootfs -nocompress=true ./templates/

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"

	redis "gopkg.in/redis.v2"

	"github.com/harrowio/harrow/bus/logevent"
	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/fsbuilder"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/stores"
)

type rootFsFile struct {
	path     string
	contents []byte
}

func NewBuilder(c *fsbuilder.Config) *rootFs {
	return &rootFs{
		config:       c,
		systemConfig: config.GetConfig(),
	}
}

type anyNotifier struct {
	Type     string      `json:"type"`
	Notifier interface{} `json:"notifier"`
}

type rootFs struct {
	systemConfig  *config.Config
	config        *fsbuilder.Config
	OperationCtxt *domain.OperationSetupScriptCtxt

	operation *domain.Operation
	project   *domain.Project
	wsbi      *domain.WorkspaceBaseImage

	triggeringActivity *domain.Activity

	job        *domain.Job
	task       *domain.Task
	repository *domain.Repository

	log logger.Logger

	notifier anyNotifier
}

func (self *rootFs) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self *rootFs) SetLogger(l logger.Logger) {
	self.log = l
}

func (self *rootFs) Build(operationUuid string) (io.Reader, error) {

	if err := self.loadOperation(operationUuid); err != nil {
		return nil, err
	}

	if err := self.constructContext(); err != nil {
		return nil, err
	}

	if r, err := self.buildFilesystem(); err != nil {
		return nil, err
	} else {
		return r, nil
	}

}

func (self *rootFs) loadOperation(operationUuid string) error {

	opStore := stores.NewDbOperationStore(self.config.Tx())

	if operation, err := opStore.FindByUuid(operationUuid); err != nil {
		return fmt.Errorf("Can't load operation #uuid %s: %s", operationUuid, err)
	} else {
		self.operation = operation
	}

	return nil
}

func (self *rootFs) constructContext() error {

	// Special Storage
	secretKvStore := self.config.Secrets()

	// Database Storage
	projStore := stores.NewDbProjectStore(self.config.Tx())
	repositoryCredentialStore := stores.NewRepositoryCredentialStore(secretKvStore, self.config.Tx())
	repositoryStore := stores.NewDbRepositoryStore(self.config.Tx())
	wsbiStore := stores.NewDbWorkspaceBaseImageStore(self.config.Tx())
	taskStore := stores.NewDbTaskStore(self.config.Tx())
	jobStore := stores.NewDbJobStore(self.config.Tx())
	deliveryStore := stores.NewDbDeliveryStore(self.config.Tx())
	secretStore := stores.NewSecretStore(secretKvStore, self.config.Tx())
	credentialStore := stores.NewRepositoryCredentialStore(secretKvStore, self.config.Tx())
	activityStore := stores.NewDbActivityStore(self.config.Tx())
	environmentStore := stores.NewDbEnvironmentStore(self.config.Tx())
	operationStore := stores.NewDbOperationStore(self.config.Tx())
	notifierStore := stores.NewDbNotifierStore(self.config.Tx())
	if canProceed, err := self.operation.IsReady(repositoryStore, credentialStore, environmentStore, secretStore); err != nil {
		return fmt.Errorf("Error checking operation readiness: %s", err)
	} else {
		if canProceed == false {
			return fmt.Errorf("Unable to build rootfs for operation, missing pieces (operation uuid %s)", self.operation.Uuid)
		}
	}

	if self.operation.NotifierUuid != nil && self.operation.NotifierType != nil {
		notifier, err := notifierStore.FindByUuidAndType(*self.operation.NotifierUuid, *self.operation.NotifierType)
		if err != nil {
			return err
		}

		self.notifier.Notifier = notifier
		self.notifier.Type = *self.operation.NotifierType
	}

	if project, err := self.operation.FindProject(projStore); err != nil {
		return fmt.Errorf("Can't load project for operation #uuid %s: %s", self.operation.Uuid, err)
	} else {
		self.project = project
	}

	if wsbi, err := self.operation.FindWorkspaceBaseImage(wsbiStore); err != nil {
		return fmt.Errorf("Can't load operation workspace base iamge %s: %s\n", self.operation.WorkspaceBaseImageUuid, err)
	} else {
		self.wsbi = wsbi
	}

	if jobUuid := self.operation.JobUuid; jobUuid != nil {
		if job, err := jobStore.FindByUuid(*jobUuid); err != nil {
			self.Log().Warn().Msgf("Cannot find job %q: %s", *jobUuid, err)
		} else {
			self.job = job
		}
	}

	// Optional Fields, one or the other should always be absent
	if self.operation.IsUserJob() {
		if task, err := self.operation.Task(taskStore); err != nil {
			return fmt.Errorf("Can't load operation task %s: %s\n", self.operation.JobUuid, err)
		} else {
			self.task = task
		}
	} else if self.operation.Category() == "repository" {
		repository, err := repositoryStore.FindByUuid(*self.operation.RepositoryUuid)
		if err != nil {
			return fmt.Errorf("Can't load operation repository %s: %s\n", *self.operation.RepositoryUuid, err)
		}
		self.repository = repository
		basicCredential, err := repositoryCredentialStore.FindByRepositoryUuidAndType(self.repository.Uuid, domain.RepositoryCredentialBasic)
		if err == nil && basicCredential != nil {
			self.repository.SetCredential(basicCredential)
		}
	}

	OperationCtxt, err := self.operation.NewSetupScriptCtxt(self.wsbi, self.project)
	if err != nil {
		return fmt.Errorf("Can't initialize setup script context: %s", err)
	}
	self.OperationCtxt = OperationCtxt

	parameters := self.operation.Parameters
	repositories, err := self.operation.Repositories(repositoryStore)
	if err != nil {
		return fmt.Errorf("Can't load operation repositories: %s", err)
	} else {
		fmt.Fprintf(os.Stderr, "found some repos %#v\n", repositories)
		self.OperationCtxt.Repositories = repositories
	}

	activity, err := activityStore.FindActivityById(parameters.TriggeredByActivityId)
	if _, notFound := err.(*domain.NotFoundError); err != nil && !notFound {
		return fmt.Errorf("Can't load operation triggering activity: %s", err)
	} else if activity != nil {
		self.triggeringActivity = activity
		operation, ok := activity.Payload.(*domain.Operation)
		if ok {
			if err := self.loadOperationLogs(operation); err != nil {
				self.Log().Error().Msgf("Failed to load operation logs for operation %q: %s", operation.Uuid, err)
			}
			self.triggeringActivity.Payload = operation
		}
		if ok && operation.JobUuid != nil {
			job, err := jobStore.FindByUuid(*operation.JobUuid)
			if err == nil {
				self.job = job
			}
		}
	}

	for _, repository := range repositories {

		if _, found := parameters.Checkout[repository.Uuid]; !found {
			parameters.Checkout[repository.Uuid] = "master"
		}

		basicCredential, err := repositoryCredentialStore.FindByRepositoryUuidAndType(repository.Uuid, domain.RepositoryCredentialBasic)
		if err == nil && basicCredential != nil {
			repository.SetCredential(basicCredential)
		}

		repositoryCredential, err := repositoryCredentialStore.FindByRepositoryUuidAndType(repository.Uuid, domain.RepositoryCredentialSsh)
		if err != nil && !domain.IsNotFound(err) {
			return fmt.Errorf("Can't load credentials for repository %s: %s", repository.Uuid, err)
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			fmt.Fprintf(os.Stderr, "repoCreden: %s\n", repositoryCredential)
			continue
		}

		fmt.Fprintf(os.Stderr, "do we get here?\n")

		sshRepositoryCredential, err := domain.AsSshRepositoryCredential(repositoryCredential)
		if err != nil {
			self.Log().Warn().Msgf("Can't load RepositoryCredential: %s", err)
			continue
		}

		fmt.Fprintf(os.Stderr, "getting handle on Git repo\n")
		if gitRepo, err := repository.Git(); err == nil {
			if gitRepo.UsesSSH() {
				self.OperationCtxt.AddKey("repository", repositoryCredential.Name, sshRepositoryCredential.PrivateKey, sshRepositoryCredential.PublicKey)
				fmt.Fprintf(os.Stderr, "repouses ssh credentials are %v", repositoryCredential)
				self.OperationCtxt.AddSshConfig(repository, repositoryCredential)
			} else {
				fmt.Fprintf(os.Stderr, "repo claims not use ssh")
			}
		} else {
			self.Log().Warn().Msgf("Can't parse repository URL, so skipping generating keys: %s", err)
		}
	}
	self.OperationCtxt.Parameters = parameters

	secrets, err := self.operation.Secrets(environmentStore, secretStore)
	if err != nil {
		return fmt.Errorf("Can't load secrets for operation %s: %s", self.operation.Uuid, err)
	}

	if err := self.OperationCtxt.LoadWebhookBody(deliveryStore); err != nil {
		self.Log().Warn().Msgf("Can't load delivered webhook body: %s\n", err)
	}

	if err := self.OperationCtxt.LoadPreviousOperation(operationStore); err != nil {
		self.Log().Warn().Msgf("Can't load previous operation: %s\n", err)
	}

	for _, secret := range secrets {
		if secret.IsSsh() {
			sshSecret, err := domain.AsSshSecret(secret)
			if err != nil {
				self.Log().Warn().Msgf("Can't load SshSecret: %s", err)
				continue
			}
			self.OperationCtxt.AddKey("environment", secret.Name, sshSecret.PrivateKey, sshSecret.PublicKey)
		} else {
			envSecret, err := domain.AsEnvironmentSecret(secret)
			if err != nil {
				self.Log().Warn().Msgf("Can't load EnvironmentSecret: %s", err)
				continue
			}
			self.OperationCtxt.Secrets = append(self.OperationCtxt.Secrets, envSecret)
		}
	}

	environment, err := self.operation.Environment(environmentStore)
	if err != nil {
		return fmt.Errorf("Can't load environment for operation %s: %s", self.operation.Uuid, err)
	} else if environment != nil {
		self.OperationCtxt.Environment = environment
	}

	return nil
}

func (self *rootFs) loadOperationLogs(operation *domain.Operation) error {
	redisClient := redis.NewTCPClient(self.systemConfig.RedisConnOpts(0))
	fileTransport := logevent.NewFileTransport(self.systemConfig, self.log)
	messages, err := fileTransport.Consume(operation.Uuid)
	if err != nil {
		messages, err = logevent.NewRedisTransport(redisClient, self.log).Consume(operation.Uuid)
	}
	if err != nil {
		return err
	}

	for message := range messages {
		operation.AddLogEvent(message)
	}

	return nil
}

func (self *rootFs) buildFilesystem() (io.Reader, error) {

	var rootFsBuffer bytes.Buffer
	var gzWriter *gzip.Writer = gzip.NewWriter(&rootFsBuffer)
	var tarGzWriter *tar.Writer = tar.NewWriter(gzWriter)

	// Write SSH Keys
	var sshKeyFiles []rootFsFile
	for _, sshKey := range self.OperationCtxt.Keys {
		sshPublicKey := rootFsFile{
			path:     fmt.Sprintf(".ssh/%s.pub", sshKey.Filename()),
			contents: []byte(sshKey.Public),
		}
		sshPrivateKey := rootFsFile{
			path:     fmt.Sprintf(".ssh/%s", sshKey.Filename()),
			contents: []byte(sshKey.Private),
		}
		sshKeyFiles = append(sshKeyFiles, sshPublicKey)
		sshKeyFiles = append(sshKeyFiles, sshPrivateKey)
	}

	for _, sshKeyFile := range sshKeyFiles {
		hdr := &tar.Header{
			Name:    sshKeyFile.path,
			Mode:    0600,
			ModTime: time.Now().UTC(),
			Size:    int64(len(sshKeyFile.contents)),
		}
		if err := tarGzWriter.WriteHeader(hdr); err != nil {
			return nil, fmt.Errorf("Can't add ssh key header %s to filesystem: %s", sshKeyFile.path, err)
		}
		if _, err := tarGzWriter.Write([]byte(sshKeyFile.contents)); err != nil {
			return nil, fmt.Errorf("Can't add ssh key contents %s to filesystem: %s", sshKeyFile.path, err)
		}
	}

	self.writeWebhookBodyToArchive(tarGzWriter)

	// Write SSH config
	sshConfigPath := ".ssh/config"
	var sshConfig bytes.Buffer
	err := self.template("ssh_config").Execute(&sshConfig, self.OperationCtxt)
	if err != nil {
		self.Log().Warn().Msgf("Error rendering ssh_config: %s\n", err)
		return nil, err
	}
	hdr := &tar.Header{
		Name:    sshConfigPath,
		Mode:    0600,
		ModTime: time.Now().UTC(),
		Size:    int64(sshConfig.Len()),
	}
	if err := tarGzWriter.WriteHeader(hdr); err != nil {
		return nil, fmt.Errorf("Can't add ssh config header %s to filesystem: %s", sshConfigPath, err)
	}
	if _, err := tarGzWriter.Write(sshConfig.Bytes()); err != nil {
		return nil, fmt.Errorf("Can't add ssh config contents %s to filesystem: %s", sshConfigPath, err)
	}

	// Write binary and script executables
	var setupShBuffer bytes.Buffer
	err = self.template("setup.sh").Execute(&setupShBuffer, self.OperationCtxt)
	if err != nil {
		self.Log().Warn().Msgf("Error rendering setup.sh: %s\n", err)
		return nil, err
	}

	gitAnalyzeRepository := MustAsset("templates/git-analyze-repository")

	var executableFiles []rootFsFile = []rootFsFile{
		rootFsFile{path: ".bin/git-analyze-repository", contents: gitAnalyzeRepository},
		rootFsFile{path: ".bin/setup", contents: setupShBuffer.Bytes()},
		rootFsFile{path: ".bin/git-ssh", contents: []byte("#!/bin/sh -e\nexec /usr/bin/ssh -o LogLevel=quiet -o PasswordAuthentication=no -o StrictHostKeyChecking=no \"$@\"\n")},
	}

	// Check what kind of operation we're running
	if self.operation.IsUserJob() {
		scriptRootFsFile := rootFsFile{path: ".bin/script", contents: []byte(self.task.Body)}
		executableFiles = append(executableFiles, scriptRootFsFile)
	} else if self.operation.IsGitAccessCheck() {
		var scriptBuffer bytes.Buffer
		if err != nil {
			self.Log().Warn().Msgf("Error making git.Repository from domain.Repository: %s\n", err)
			return nil, err
		}
		fmt.Fprintf(&scriptBuffer, "#!/bin/bash -e\ngit ls-remote -h \"%s\"\n", self.repository.Url)
		scriptRootFsFile := rootFsFile{path: ".bin/script", contents: scriptBuffer.Bytes()}
		executableFiles = append(executableFiles, scriptRootFsFile)
	} else if self.operation.IsGitMetadataCollect() {
		collectMetadataBody := bytes.Buffer{}
		err := self.template("collect-metadata.sh").Execute(&collectMetadataBody, self.OperationCtxt)
		if err != nil {
			self.Log().Warn().Msgf("Error rendering collect-metadata.sh: %s\n", err)
			return nil, err
		}
		scriptRootFsFile := rootFsFile{path: ".bin/script", contents: collectMetadataBody.Bytes()}
		executableFiles = append(executableFiles, scriptRootFsFile)
	}

	if err := self.writeContextData(tarGzWriter); err != nil {
		return nil, err
	}

	for _, executableFile := range executableFiles {
		hdr := &tar.Header{
			Name:    executableFile.path,
			Mode:    0744,
			ModTime: time.Now().UTC(),
			Size:    int64(len(executableFile.contents)),
		}
		if err := tarGzWriter.WriteHeader(hdr); err != nil {
			return nil, fmt.Errorf("Can't add executable file header %s to filesystem: %s", executableFile.path, err)
		}
		if _, err := tarGzWriter.Write([]byte(executableFile.contents)); err != nil {
			return nil, fmt.Errorf("Can't add executable file contents %s to filesystem: %s", executableFile.path, err)
		}
	}

	if err := tarGzWriter.Close(); err != nil {
		return nil, fmt.Errorf("Error calling tarGzWriter.Close(): %s", err)
	}

	if err := gzWriter.Close(); err != nil {
		return nil, fmt.Errorf("Error calling to gzWriter.Close(): %s", err)
	}

	return &rootFsBuffer, nil

}

func (self *rootFs) writeWebhookBodyToArchive(w *tar.Writer) {
	if self.OperationCtxt.WebhookBody == nil {
		return
	}

	body := self.OperationCtxt.WebhookBody

	hdr := &tar.Header{
		Name:    "harrow-webhook-body",
		Mode:    0644,
		ModTime: time.Now().UTC(),
		Size:    int64(len(body)),
	}

	if err := w.WriteHeader(hdr); err != nil {
		self.Log().Warn().Msgf("Cannot write tar header for webhook body: %s\n", err)
		return
	}

	if _, err := w.Write(body); err != nil {
		self.Log().Warn().Msgf("Cannot write webhook body to archive: %s\n", err)
		return
	}
}

func (self *rootFs) template(fileName string) *template.Template {
	return self.templates()[fileName]
}

func (self *rootFs) templates() map[string]*template.Template {
	return map[string]*template.Template{
		"setup.sh":            template.Must(template.New("setup.sh").Funcs(self.templateFuncMap()).Parse(string(MustAsset("templates/setup.sh")))),
		"collect-metadata.sh": template.Must(template.New("collect-metadata.sh").Funcs(self.templateFuncMap()).Parse(string(MustAsset("templates/collect-metadata.sh")))),
		"ssh_config":          template.Must(template.New("ssh_config").Funcs(self.templateFuncMap()).Parse(string(MustAsset("templates/ssh_config")))),
	}
}

func (self *rootFs) templateFuncMap() template.FuncMap {
	// Aggressively strips input to make safe filenames.
	// e.g: git@bitbucket.org:harrowio/harrow-cli.git
	// becomes: git_bitbucket_org_harrowio_harrow_cli_git
	return template.FuncMap{"asciiSafe": asciiSafe}
}

func asciiSafe(input string) string {

	ascii_safe_re := regexp.MustCompile("[^a-zA-Z0-9]")
	single_separator_re := regexp.MustCompile("_+")

	ascii_safe := ascii_safe_re.ReplaceAllString(strings.ToLower(input), "_")
	single_separator := single_separator_re.ReplaceAllString(ascii_safe, "_")

	return single_separator
}

func (self *rootFs) writeContextData(w *tar.Writer) error {
	if err := writeAsJSON(w, "harrow/operation.json", self.operation); err != nil {
		return err
	}

	for _, repository := range self.OperationCtxt.Repositories {
		if err := writeAsJSON(w, fmt.Sprintf("harrow/repositories/%s.json", repository.Uuid), repository); err != nil {
			return err
		}
	}

	if err := writeAsJSON(w, "harrow/activity.json", self.triggeringActivity); err != nil {
		return err
	}

	if err := writeAsJSON(w, "harrow/project.json", self.project); err != nil {
		return err
	}

	if err := writeAsJSON(w, "harrow/notifier.json", self.notifier); err != nil {
		return err
	}

	if self.job != nil {
		if err := writeAsJSON(w, "harrow/job.json", self.job); err != nil {

		}
	}

	return nil
}

func writeAsJSON(w *tar.Writer, path string, data interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	header := &tar.Header{
		Name:    path,
		Mode:    0644,
		ModTime: time.Now().UTC(),
		Size:    int64(len(body)),
	}

	if err := w.WriteHeader(header); err != nil {
		return fmt.Errorf("writeAsJSON: WriteHeader: %s: %s", path, err)
	}

	if _, err := w.Write(body); err != nil {
		return fmt.Errorf("writeAsJSON: Write: %s: %s", path, err)
	}

	return nil
}
