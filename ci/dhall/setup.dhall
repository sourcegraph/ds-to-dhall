let GitHubActions = (./imports.dhall).GitHubActions

let Checkout =
      GitHubActions.Step::{
      , name = None Text
      , uses = Some "actions/checkout@v2"
      }

let ASDFSteps =
      [ GitHubActions.Step::{
        , name = Some "[setup] Install asdf"
        , uses = Some "asdf-vm/actions/setup@v1.0.0"
        }
      , GitHubActions.Step::{
        , name = Some "[setup] Install asdf plugins"
        , run = Some "./ci/asdf-add-plugins.sh"
        }
      , GitHubActions.Step::{
        , name = Some "[setup] Install asdf .tool-versions"
        , run = Some "./ci/asdf-install.sh"
        }
      , GitHubActions.Step::{
        , name = Some "[setup] prepare go environment"
        , uses = Some "./ci/asdf-setup-go.sh"
        }
      ]

let SetupSteps = [ Checkout ] # ASDFSteps

let Job =
      { Type = GitHubActions.Job.Type
      , default =
          GitHubActions.Job.default
        with runs-on = GitHubActions.RunsOn.Type.ubuntu-latest
        with steps = Some SetupSteps
      }

in  { SetupSteps, Job }
