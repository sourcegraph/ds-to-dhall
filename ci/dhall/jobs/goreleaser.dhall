let GitHubActions = (../imports.dhall).GitHubActions

let Setup = ../setup.dhall

let SetupSteps = Setup.SetupSteps

let Job = Setup.Job

in  Job::{
    , name = Some "goreleaser"
    , steps =
          SetupSteps
        # [ GitHubActions.Step::{
            , name = Some "goreleaser"
            , uses = Some "goreleaser/goreleaser-action@v2"
            , `if` = Some "startsWith(github.ref, 'refs/tags/')"
            , `with` = Some
                (toMap { version = "latest", args = "release --rm-dist" })
            , env = Some (toMap { GITHUB_TOKEN = "\${{ secrets.GH_TOKEN }}" })
            }
          ]
    }
