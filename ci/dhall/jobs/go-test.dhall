let GitHubActions = (../imports.dhall).GitHubActions

let Setup = ../setup.dhall

let SetupSteps = Setup.SetupSteps

let Job = Setup.Job

in  Job::{
    , name = Some "go-test"
    , steps =
          SetupSteps
        # [ GitHubActions.Step::{
            , name = Some "go-test"
            , run = Some "ci/go-test.sh"
            }
          ]
    }
