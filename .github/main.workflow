workflow "Build master release draft" {
  on = "push"
  resolves = [
    "Add darwin 386 binary to draft",
    "Add darwin amd64 binary to draft",
    "Add linux 386 binary to draft",
    "Add linux amd64 binary to draft",
    "Add windows 386 binary to draft",
    "Add windows amd64 binary to draft"
  ]
}

action "Filter for master" {
  uses = "actions/bin/filter@master"
  args = "branch master"
}

action "Build" {
  needs = ["Filter for master"]
  uses = "docker://golang:1.12-stretch"
  runs = ["sh", "-c", "$GITHUB_WORKSPACE/.github/build.sh"]
}

action "Draft" {
  needs = ["Build"]
  uses = "swinton/httpie.action@master"
  args = ["--auth-type=jwt", "--auth=$GITHUB_TOKEN", "POST", "api.github.com/repos/$GITHUB_REPOSITORY/releases", "tag_name=`cat $GITHUB_WORKSPACE/.version`", "name=`cat $GITHUB_WORKSPACE/.version`", "draft:=true"]
  secrets = ["GITHUB_TOKEN"]
}

action "Add darwin 386 binary to draft" {
  needs = ["Draft"]
  uses = "swinton/httpie.action@master"
  args = ["--auth-type=jwt", "--auth=$GITHUB_TOKEN", "POST", "`jq .upload_url $HOME/Draft.response.body --raw-output | sed 's/{.*}//g'`?name=eager-`cat $GITHUB_WORKSPACE/.version`_darwin-386.tar.gz", "@$GITHUB_WORKSPACE/.dist/eager-darwin-386.tar.gz"]
  secrets = ["GITHUB_TOKEN"]
}

action "Add darwin amd64 binary to draft" {
  needs = ["Draft"]
  uses = "swinton/httpie.action@master"
  args = ["--auth-type=jwt", "--auth=$GITHUB_TOKEN", "POST", "`jq .upload_url $HOME/Draft.response.body --raw-output | sed 's/{.*}//g'`?name=eager-`cat $GITHUB_WORKSPACE/.version`_darwin-amd64.tar.gz", "@$GITHUB_WORKSPACE/.dist/eager-darwin-amd64.tar.gz"]
  secrets = ["GITHUB_TOKEN"]
}

action "Add linux 386 binary to draft" {
  needs = ["Draft"]
  uses = "swinton/httpie.action@master"
  args = ["--auth-type=jwt", "--auth=$GITHUB_TOKEN", "POST", "`jq .upload_url $HOME/Draft.response.body --raw-output | sed 's/{.*}//g'`?name=eager-`cat $GITHUB_WORKSPACE/.version`_linux-386.tar.gz", "@$GITHUB_WORKSPACE/.dist/eager-linux-386.tar.gz"]
  secrets = ["GITHUB_TOKEN"]
}

action "Add linux amd64 binary to draft" {
  needs = ["Draft"]
  uses = "swinton/httpie.action@master"
  args = ["--auth-type=jwt", "--auth=$GITHUB_TOKEN", "POST", "`jq .upload_url $HOME/Draft.response.body --raw-output | sed 's/{.*}//g'`?name=eager-`cat $GITHUB_WORKSPACE/.version`_linux-amd64.tar.gz", "@$GITHUB_WORKSPACE/.dist/eager-linux-amd64.tar.gz"]
  secrets = ["GITHUB_TOKEN"]
}

action "Add windows 386 binary to draft" {
  needs = ["Draft"]
  uses = "swinton/httpie.action@master"
  args = ["--auth-type=jwt", "--auth=$GITHUB_TOKEN", "POST", "`jq .upload_url $HOME/Draft.response.body --raw-output | sed 's/{.*}//g'`?name=eager-`cat $GITHUB_WORKSPACE/.version`_windows-386.tar.gz", "@$GITHUB_WORKSPACE/.dist/eager-windows-386.tar.gz"]
  secrets = ["GITHUB_TOKEN"]
}

action "Add windows amd64 binary to draft" {
  needs = ["Draft"]
  uses = "swinton/httpie.action@master"
  args = ["--auth-type=jwt", "--auth=$GITHUB_TOKEN", "POST", "`jq .upload_url $HOME/Draft.response.body --raw-output | sed 's/{.*}//g'`?name=eager-`cat $GITHUB_WORKSPACE/.version`_windows-amd64.tar.gz", "@$GITHUB_WORKSPACE/.dist/eager-windows-amd64.tar.gz"]
  secrets = ["GITHUB_TOKEN"]
}
