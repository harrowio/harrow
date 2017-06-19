#!/bin/bash -e

main() {
  clone_repository
  analyze_repository
}

clone_repository() {
  mkdir -p ~/repositories
  {{with $repository := (index $.Repositories 0)}}
  (cd ~/repositories
   repository_name=$(basename "{{.Url}}" .git)
   report_status git.clone-started subject="Cloning {{.Url}} into $repository_name"
   if ! git clone -q "{{ $repository.CloneURL }}"; then
       exit 76 {{/* See: EX_PROTOCOL /usr/include/sysexits.h */}}
   fi
   report_status git.clone-finished subject="Finished cloning {{.Url}}"
   cd $repository_name
   install_git_checkout_hook {{ $repository.Uuid }}
   git checkout {{index $.Parameters.Checkout $repository.Uuid}} > /tmp/{{$repository.Uuid}}-checkout.log 2>&1
  )
  {{end}}
}

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

analyze_repository() {
  {{with $repo := (index $.Repositories 0)}}
  (cd ~/repositories/$(basename {{$repo.Url}} .git)
   git analyze-repository > /tmp/{{$repo.Uuid}}-analysis.json
   hevent analyze-repository repository={{$repo.Uuid}} data="$(cat /tmp/{{$repo.Uuid}}-analysis.json)"
  )
  {{end}}
}

report_status() {
  local type=$1
  shift
  hevent status type=$type "$@"
}

main
