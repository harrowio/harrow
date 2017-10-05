#!/bin/bash

HARROW_PREFIX=$HOME

# trap fatal errors and send hevent
set -o errtrace # to inherit traps in functions and subshells
function report_err() {
  if command -v hevent >/dev/null 2>&1; then
    hevent fatal script="$0" line="$1" func="$2" cmd="$3" status="$4"
  fi
}
trap 'report_err $LINENO "${FUNCNAME[0]}" "$BASH_COMMAND" $?' ERR

function main() {

  install_support_scripts
  source_bash_profile
  dot_ssh_dir_create

  ssh_agent_init
  ssh_agent_load_keys

  trap report_finish EXIT INT

  export_harrow_version
  export_gitvars

  {{ if .Environment }}export_env_vars{{ end }}
  {{ if .Secrets }}export_secret_vars{{ end }}

  {{ if .ShouldCloneRepos }}clone_repositories{{end}}

  discourage_badly_behaved_programs

  run_user_script

}

function report_finish() {
  send_git_logs
  {{ if .ShouldCloneRepos }}analyze_repositories{{end}}
  report_status vm.done subject="User script finished, shutting down VM"
}

function analyze_repositories() {
  true
  export RBENV_VERSION=2.2.2
  {{with $.Repositories}}
  {{range $repo := $.Repositories}}
  (cd ~/repositories/$(basename {{$repo.Url}} .git)
   git analyze-repository > /tmp/{{$repo.Uuid}}-analysis.json
   hevent analyze-repository repository={{$repo.Uuid}} data@/tmp/{{$repo.Uuid}}-analysis.json
  )
  {{end}}
  {{end}}
}

function report_status() {
  local type=$1
  shift

  if ! command -v hevent >/dev/null 2>&1; then
      return 0
  fi

  hevent status type=$type "$@"
}

function install_support_scripts() {
    sudo cp .bin/git-analyze-repository /usr/local/bin/
    sudo chmod +x /usr/local/bin/git-analyze-repository

    cat <<'END_OF_SCRIPT' > .bin/git-commit-to-json.awk
#!/usr/bin/awk -f

BEGIN {
    FS=":[ ]+"
    SAFE_KEYS["Parents"] = 1
    IN_BODY=0
    BODY=""
}

function json_quote(text) {
    split(text, lines, "\n")
    out=""
    for (line in lines) {
        gsub(/\\/, "\\\\", lines[line])
        gsub(/\t/, "\\t", lines[line])
        gsub(/"/, "\\\"", lines[line])
        out = out lines[line]
        if (length(lines) > 1) {
            out = out "\\n"
        }
    }
    return "\"" out "\""
}

function json_object(object) {
    i = 0
    n = length(object)
    out = "{"
    for (key in object) {
        i++
        out = out json_quote(key) ":"
        if (SAFE_KEYS[key]) {
            out = out object[key]
        } else {
            out = out json_quote(object[key])
        }
        if (i < n) { out = out "," }
    }
    out = out "}"

    return out
}

function json_array(array) {
    i = 0
    n = length(array)
    out = "["
    for (key in array) {
        i++
        out = out json_quote(array[key])
        if (i < n) { out = out "," }
    }
    out = out "]"

    return out
}

!IN_BODY && length($0) == 0 {
    IN_BODY = 1
    next
}

!IN_BODY && /^Parents:/ {
    split($2, PARENTS, " ")
    FIELDS["Parents"] = json_array(PARENTS)
    next
}

!IN_BODY && /^[-A-Za-z]+:/ {
    FIELDS[$1]=$2
    next
}

IN_BODY {
    BODY=BODY $0 "\n"
}

END {
    FIELDS["Body"] = BODY
    print json_object(FIELDS)
}

END_OF_SCRIPT
    cat <<'END_OF_SCRIPT' > .bin/git-commit-to-json
#!/bin/bash

CMD=cat

git_commit_format() {
    printf "%s: %%%s%%n" \
           Commit H \
           AuthorDate ai \
           Author aN \
           Parents P \
           AuthorEmail aE \
           Subject s \


    printf "%%n%%b"
}

usage() {
    cat <<EOF
$0 [-c CMD] <commit-id> <to-id>

Outputs information about <commit-id> as a json object with the
following keys:

    Commit       [string] the commit id
    Author       [string] the author of the commit
    AuthorEmail  [string] the author's email address
    AuthorDate   [string] the commit's authoring date (ISO 8601)
    Subject      [string] the commit's subject line
    Body         [string] the commit message's body
    Parents      [array ] the commit's parent hashes

If <to-id> is provided, the commits which are reachable only reachable
from <to-id> but not from <from-id> are displayed.  Note that this is
a stream of JSON objects, not a JSON array.


If <-c CMD> is provided, CMD will be executed for each commit, passing
the serialized commit on stdin.

EOF

}

format_commit() {
    awk -f ${0}.awk
}

select_commits() {
    local from to
    from=$1
    to=$2

    if [ -z "$to" ]; then
        git rev-list -n 5 $from
    else
        git rev-list $from..$to
    fi
}

main() {
    for arg in "$@"; do
        case $arg in
            -c)
                shift
                CMD="$1"
                shift
        esac
    done

    if [ -z "$1" ]; then
        usage
        return 1
    fi

    select_commits "$1" "$2" |
        while read commit_id; do
            git show --no-patch --format="$(git_commit_format)%n" "$commit_id" |
                format_commit |
                eval "${CMD}"
        done
}

main "$@"


END_OF_SCRIPT
    chmod +x .bin/git-commit-to-json
}

function send_git_logs() {
    true
    {{with .PreviousOperation}}
    {{range $repo := $.Repositories}}
    (cd ~/repositories/$(basename ${{$repo.CloneURL}} .git)
    $HARROW_PREFIX/.bin/git-commit-to-json -c 'tee git-commit.json >/dev/null; hevent git-log repository={{$repo.Uuid}} entry@git-commit.json' {{$.PreviousOperation.RepositoryCheckouts.Hash $repo.Uuid}} HEAD)
    {{end}}
    {{end}}
}

function source_bash_profile() {
    if [ -e $HARROW_PREFIX/.bash_profile ]; then
        source $HARROW_PREFIX/.bash_profile
    fi
    if [ -e /etc/profile ]; then
        source /etc/profile
    fi
}

function run_user_script() {
  report_status vm.run-user-script subject="Running user script"
  # do not treat errors in the user script as fatal failures
  trap - ERR
  DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
  ${DIR%%/}/script
  exit
}

function dot_ssh_dir_create() {
  local USER=$(whoami)
  mkdir -p ~/.ssh
  chmod 700 ~/.ssh
  chmod -R 600 ~/.ssh/*
  sudo chown -R $USER:$USER ~/.ssh
}

function ssh_agent_init() {
  eval $(ssh-agent -s) > /dev/null 2>&1
}

function export_harrow_version() {
  export HARROW_VERSION="1.0.0"
}

function export_gitvars() {
  export GIT_ASKPASS="/bin/echo"
  export GIT_SSH="$HARROW_PREFIX/.bin/git-ssh"
}

function discourage_badly_behaved_programs() {
  export DEBIAN_FRONTEND="noninteractive"
}

function ssh_agent_load_keys() {
  true
  {{ if .Keys }}{{ range $key := .Keys }}{{ if (not $key.BelongsToRepository) }}
  ssh-add ".ssh/{{ $key.Filename }}" >/dev/null 2>&1
  {{ end }}{{ end }}{{ end }}
}

{{ if .ShouldCloneRepos }}function clone_repositories() {
  mkdir -p repositories/
  {{with .Repositories}}(
    {{ range . }}clone_repository_{{ .Url | asciiSafe }}
    {{ end }}
  ){{end}}
}

{{ range $repository := .Repositories }}function clone_repository_{{ $repository.Url | asciiSafe }}() {
  (
    cd repositories/
    repository_name=$(basename "{{.Url}}" .git)

    report_status git.clone-started subject="Cloning {{.PublicURL}} into $repository_name"
    if [ -d "$repository_name" ]; then
        report_status git.clone-failed subject="Cloning {{.PublicURL}} failed: another repository named $repository_name exists"
        exit 0
    fi

    if ! git clone -q "{{ $repository.CloneURL }}"; then
        exit 76 {{/* See: EX_PROTOCOL /usr/include/sysexits.h */}}
    fi
    report_status git.clone-finished subject="Finished cloning {{$repository.PublicURL}}"
    cd $repository_name
    install_git_checkout_hook {{ $repository.Uuid }}
    checkout="{{index $.Parameters.Checkout $repository.Uuid}}"
    if ! git symbolic-ref -q "$checkout"; then
        checkout=$(git branch --list | grep -F '*' | cut -b 3-)
    fi

    git checkout ${checkout#*/*/} > /tmp/{{$repository.Uuid}}-checkout.log 2>&1
  )
}

{{ end }}{{ end }}

install_git_checkout_hook() {
    local repository_uuid=$1
    cat <<EOF > .git/hooks/post-checkout
#!/bin/bash
new_head=\$2
repository=$repository_uuid
ref="\$(git symbolic-ref HEAD)"
hevent checkout repository=\$repository hash=\$new_head ref="\$ref"
EOF
    chmod +x .git/hooks/post-checkout
}

{{ if .Environment }}function export_env_vars() {
  {{if .Environment.Variables.M}}
  {{ range $key, $value := .Environment.Variables.M }}export {{$key}}="{{$value}}"
  {{ end }}
  {{else}}:{{end}}
}{{ end }}

{{ if .Secrets }}function export_secret_vars() {
  {{ range $secret := .Secrets }}export {{$secret.Name}}="{{$secret.Value}}"
  {{ end }}
} {{ end }}


main
