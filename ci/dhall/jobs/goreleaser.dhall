let GitHubActions = (../imports.dhall).GitHubActions

let Setup = ../setup.dhall

let SetupSteps = Setup.SetupSteps

let Job = Setup.Job

let isTag = "startsWith(github.ref, 'refs/tags/')"

let base =
      GitHubActions.Step::{
      , name = Some "goreleaser"
      , uses = Some "goreleaser/goreleaser-action@v2"
      }

let publish
    : GitHubActions.Step.Type
    =   base
      ⫽ { `if` = Some isTag
        , name = Some "build go binaries for release"
        , `with` = Some
            (toMap { version = "latest", args = "release --rm-dist" })
        , env = Some (toMap { GITHUB_TOKEN = "\${{ secrets.GH_TOKEN }}" })
        }

let test_publish
    : GitHubActions.Step.Type
    =   base
      ⫽ { `if` = Some "! ${isTag}"
        , name = Some "test building go binaries"
        , `with` = Some
            ( toMap
                { version = "latest"
                , args = "--snapshot --skip-publish --rm-dist"
                }
            )
        }

in  Job::{
    , name = Some "goreleaser"
    , steps = SetupSteps # [ publish, test_publish ]
    }
