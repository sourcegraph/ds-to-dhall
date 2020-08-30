let GitHubActions = (./imports.dhall).GitHubActions

let ASDFSteps =
      [ GitHubActions.Step::{
        , name = Some "Install asdf"
        , uses = Some "asdf-vm/actions/setup@v1.0.0"
        }
      , GitHubActions.Step::{
        , name = Some "Install asdf plugins"
        , run = Some "./ci/asdf-add-plugins.sh"
        }
      , GitHubActions.Step::{
        , name = Some "Install asdf tools versions"
        , run = Some "./ci/asdf-install.sh"
        }
      , GitHubActions.Step::{
        , name = Some "setup go"
        , uses = Some "actions/setup-go@v1"
        , `with` = Some (toMap { go-version = "1.15" })
        }
      ]

let Checkout =
      GitHubActions.Step::{
      , name = None Text
      , uses = Some "actions/checkout@v2"
      }

let SetupSteps = [ Checkout ] # ASDFSteps

let Job =
      { Type = GitHubActions.Job.Type
      , default =
          GitHubActions.Job.default
        with runs-on = GitHubActions.RunsOn.Type.ubuntu-latest
        with steps = Some SetupSteps
      }

in  { SetupSteps, Job }
