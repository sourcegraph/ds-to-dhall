# ds-to-dhall

ds-to-dhall is a (for now internal) tool that lets us keep up with changes in
[deploy-sourcegraph](https://github.com/sourcegraph/deploy-sourcegraph) and friends while we are designing, engineering
and eventually migrating to a deploymemnt solution based on [Dhall](https://dhall-lang.org/#).

## Intro

The tool takes as input a directory tree of Kubernetes manifests and produces as output a Dhall expression of those
manifests.

The type of the produced Dhall expression is derived from the input directory tree of Kubernetes manifests.
Resources from that input are organized by component, kind and name. Kind and name are directly pulled from the manifests
themselves. Component is currently deduced from the input directory structure (ie resources inside the base/frontend
 subdirectory are assigned to the frontend component etc). This will later be improved by introducing labels in the
 manifests that directly declare components and thus eliminate the need to derive metadata from a directory structure 
 (which can be brittle and has exceptions that violate the location assumptions).
 
 Resources are placed in the result record by component -> kind -> name and given the appropriate Dhall type from the
  [Kubernetes Dhall schema](https://github.com/dhall-lang/dhall-kubernetes/blob/master/1.18/schemas.dhall).

## Usage

```shell script
ds-to-dhall -src ~/work/deploy-sourcegraph/base -dst ~/Desktop/record.dhall
```  

## Example schema snippet

```text
{ Gitserver :
      { Service :
          { gitserver :
              ( https://raw.githubusercontent.com/dhall-lang/dhall-kubernetes/f4bf4b9ddf669f7149ec32150863a93d6c4b3ef1/1.18/schemas.dhall
              ).Service.Type
          }
      }
  }
⩓ { Gitserver :
      { StatefulSet :
          { gitserver :
              ( https://raw.githubusercontent.com/dhall-lang/dhall-kubernetes/f4bf4b9ddf669f7149ec32150863a93d6c4b3ef1/1.18/schemas.dhall
              ).StatefulSet.Type
          }
      }
  }
⩓ { Indexed-Search :
      { Service :
          { indexed-search-indexer :
              ( https://raw.githubusercontent.com/dhall-lang/dhall-kubernetes/f4bf4b9ddf669f7149ec32150863a93d6c4b3ef1/1.18/schemas.dhall
              ).Service.Type
          }
      }
  }
⩓ { Indexed-Search :
      { Service :
          { indexed-search :
              ( https://raw.githubusercontent.com/dhall-lang/dhall-kubernetes/f4bf4b9ddf669f7149ec32150863a93d6c4b3ef1/1.18/schemas.dhall
              ).Service.Type
          }
      }
  }
⩓ { Indexed-Search :
      { StatefulSet :
          { indexed-search :
              ( https://raw.githubusercontent.com/dhall-lang/dhall-kubernetes/f4bf4b9ddf669f7149ec32150863a93d6c4b3ef1/1.18/schemas.dhall
              ).StatefulSet.Type
          }
      }
  }
⩓ { Prometheus :
      { ClusterRole :
          { prometheus :
              ( https://raw.githubusercontent.com/dhall-lang/dhall-kubernetes/f4bf4b9ddf669f7149ec32150863a93d6c4b3ef1/1.18/schemas.dhall
              ).ClusterRole.Type
          }
      }
  }

... (and so on and so forth)
```

## Example result snippet
 
```text
...
 , Gitserver =
  {
  ... 
  
  , StatefulSet.gitserver =
    { apiVersion = "apps/v1"
    , kind = "StatefulSet"
    , metadata =
      { annotations = Some
          ( toMap
              { description =
                  "Stores clones of repositories to perform Git operations."
              }
          )
      , clusterName = None Text
      , creationTimestamp = None Text
      , deletionGracePeriodSeconds = None Natural
      , deletionTimestamp = None Text
      , finalizers = None (List Text)
      , generateName = None Text
      , generation = None Natural
      , labels = Some
          ( toMap
              { sourcegraph-resource-requires = "no-cluster-admin"
              , deploy = "sourcegraph"
              }
          )
      , managedFields =
          None
            ( List
                { apiVersion : Text
                , fieldsType : Optional Text
                , fieldsV1 : Optional (List { mapKey : Text, mapValue : Text })
                , manager : Optional Text
                , operation : Optional Text
                , time : Optional Text
                }
            )
      , name = Some "gitserver"
      , namespace = None Text
      , ownerReferences =
          None
            ( List
                { apiVersion : Text
                , blockOwnerDeletion : Optional Bool
                , controller : Optional Bool
                , kind : Text
                , name : Text
                , uid : Text
                }
            )
      , resourceVersion = None Text
      , selfLink = None Text
      , uid = None Text
      }

...

```

Component -> Kind -> Name. Usually if there is only one resource of that kind in a component this gets collapsed
as Kind.Name. But we do have cases where there are multiple Services or ConfigMaps or Deployments in one component, so
you will see a subrecord of the kind and then fields for each by Name.
