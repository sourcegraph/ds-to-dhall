{ Frontend =
  { Deployment.sourcegraph-frontend =
    { apiVersion = "apps/v1"
    , kind = "Deployment"
    , metadata =
      { annotations.description =
          "Serves the frontend of Sourcegraph via HTTP(S)."
      , labels =
        { deploy = "sourcegraph"
        , sourcegraph-component = "frontend"
        , sourcegraph-resource-requires = "no-cluster-admin"
        }
      , name = "sourcegraph-frontend"
      }
    , spec =
      { minReadySeconds = 10
      , replicas = 1
      , revisionHistoryLimit = 10
      , selector.matchLabels.app = "sourcegraph-frontend"
      , strategy =
        { rollingUpdate = { maxSurge = 2, maxUnavailable = 0 }
        , type = "RollingUpdate"
        }
      , template =
        { metadata.labels =
          { app = "sourcegraph-frontend", deploy = "sourcegraph" }
        , spec =
          { containers =
            [ { args = [ "serve" ]
              , env =
                [ { name = "PGDATABASE"
                  , value = Some "sg"
                  , valueFrom =
                      None
                        { fieldRef :
                            { apiVersion : Optional Text, fieldPath : Text }
                        }
                  }
                , { name = "PGHOST"
                  , value = Some "pgsql"
                  , valueFrom =
                      None
                        { fieldRef :
                            { apiVersion : Optional Text, fieldPath : Text }
                        }
                  }
                , { name = "PGPORT"
                  , value = Some "5432"
                  , valueFrom =
                      None
                        { fieldRef :
                            { apiVersion : Optional Text, fieldPath : Text }
                        }
                  }
                , { name = "PGSSLMODE"
                  , value = Some "disable"
                  , valueFrom =
                      None
                        { fieldRef :
                            { apiVersion : Optional Text, fieldPath : Text }
                        }
                  }
                , { name = "PGUSER"
                  , value = Some "sg"
                  , valueFrom =
                      None
                        { fieldRef :
                            { apiVersion : Optional Text, fieldPath : Text }
                        }
                  }
                , { name = "SRC_GIT_SERVERS"
                  , value = Some "gitserver-0.gitserver:3178"
                  , valueFrom =
                      None
                        { fieldRef :
                            { apiVersion : Optional Text, fieldPath : Text }
                        }
                  }
                , { name = "POD_NAME"
                  , value = None Text
                  , valueFrom = Some
                    { fieldRef =
                      { apiVersion = None Text, fieldPath = "metadata.name" }
                    }
                  }
                , { name = "CACHE_DIR"
                  , value = Some "/mnt/cache/\$(POD_NAME)"
                  , valueFrom =
                      None
                        { fieldRef :
                            { apiVersion : Optional Text, fieldPath : Text }
                        }
                  }
                , { name = "GRAFANA_SERVER_URL"
                  , value = Some "http://grafana:30070"
                  , valueFrom =
                      None
                        { fieldRef :
                            { apiVersion : Optional Text, fieldPath : Text }
                        }
                  }
                , { name = "JAEGER_SERVER_URL"
                  , value = Some "http://jaeger-query:16686"
                  , valueFrom =
                      None
                        { fieldRef :
                            { apiVersion : Optional Text, fieldPath : Text }
                        }
                  }
                , { name = "PRECISE_CODE_INTEL_BUNDLE_MANAGER_URL"
                  , value = Some "http://precise-code-intel-bundle-manager:3187"
                  , valueFrom =
                      None
                        { fieldRef :
                            { apiVersion : Optional Text, fieldPath : Text }
                        }
                  }
                , { name = "PROMETHEUS_URL"
                  , value = Some "http://prometheus:30090"
                  , valueFrom =
                      None
                        { fieldRef :
                            { apiVersion : Optional Text, fieldPath : Text }
                        }
                  }
                ]
              , image =
                  "index.docker.io/sourcegraph/frontend:3.19.2@sha256:776606b680d7ce4a5d37451831ef2414ab10414b5e945ed5f50fe768f898d23f"
              , livenessProbe = Some
                { httpGet =
                  { path = "/healthz", port = "http", scheme = "HTTP" }
                , initialDelaySeconds = 300
                , timeoutSeconds = 5
                }
              , name = "frontend"
              , ports =
                [ { containerPort = 3080
                  , name = Some "http"
                  , protocol = None Text
                  }
                , { containerPort = 3090
                  , name = Some "http-internal"
                  , protocol = None Text
                  }
                ]
              , readinessProbe = Some
                { httpGet =
                  { path = "/healthz", port = "http", scheme = "HTTP" }
                , periodSeconds = 5
                , timeoutSeconds = 5
                }
              , resources =
                { limits = { cpu = "2", memory = "4G" }
                , requests = { cpu = "2", memory = "2G" }
                }
              , terminationMessagePolicy = Some "FallbackToLogsOnError"
              , volumeMounts = Some
                [ { mountPath = "/mnt/cache", name = "cache-ssd" } ]
              }
            , { args =
                [ "--reporter.grpc.host-port=jaeger-collector:14250"
                , "--reporter.type=grpc"
                ]
              , env =
                [ { name = "POD_NAME"
                  , value = None Text
                  , valueFrom = Some
                    { fieldRef =
                      { apiVersion = Some "v1", fieldPath = "metadata.name" }
                    }
                  }
                ]
              , image =
                  "index.docker.io/sourcegraph/jaeger-agent:3.19.2@sha256:e757094c04559780dba1ded3475ee5f0e4e5330aa6bbc8a7398e7277b0e450fe"
              , livenessProbe =
                  None
                    { httpGet : { path : Text, port : Text, scheme : Text }
                    , initialDelaySeconds : Natural
                    , timeoutSeconds : Natural
                    }
              , name = "jaeger-agent"
              , ports =
                [ { containerPort = 5775
                  , name = None Text
                  , protocol = Some "UDP"
                  }
                , { containerPort = 5778
                  , name = None Text
                  , protocol = Some "TCP"
                  }
                , { containerPort = 6831
                  , name = None Text
                  , protocol = Some "UDP"
                  }
                , { containerPort = 6832
                  , name = None Text
                  , protocol = Some "UDP"
                  }
                ]
              , readinessProbe =
                  None
                    { httpGet : { path : Text, port : Text, scheme : Text }
                    , periodSeconds : Natural
                    , timeoutSeconds : Natural
                    }
              , resources =
                { limits = { cpu = "1", memory = "500M" }
                , requests = { cpu = "100m", memory = "100M" }
                }
              , terminationMessagePolicy = None Text
              , volumeMounts = None (List { mountPath : Text, name : Text })
              }
            ]
          , securityContext.runAsUser = 0
          , serviceAccountName = "sourcegraph-frontend"
          , volumes = [ { emptyDir = {=}, name = "cache-ssd" } ]
          }
        }
      }
    }
  , Ingress.sourcegraph-frontend =
    { apiVersion = "networking.k8s.io/v1beta1"
    , kind = "Ingress"
    , metadata =
      { annotations =
        { `kubernetes.io/ingress.class` = "nginx"
        , `nginx.ingress.kubernetes.io/proxy-body-size` = "150m"
        }
      , labels =
        { app = "sourcegraph-frontend"
        , deploy = "sourcegraph"
        , sourcegraph-component = "frontend"
        , sourcegraph-resource-requires = "no-cluster-admin"
        }
      , name = "sourcegraph-frontend"
      }
    , spec.rules =
      [ { http.paths =
          [ { backend =
              { serviceName = "sourcegraph-frontend", servicePort = 30080 }
            , path = "/"
            }
          ]
        }
      ]
    }
  , Role.sourcegraph-frontend =
    { apiVersion = "rbac.authorization.k8s.io/v1"
    , kind = "Role"
    , metadata =
      { labels =
        { category = "rbac"
        , deploy = "sourcegraph"
        , sourcegraph-component = "frontend"
        , sourcegraph-resource-requires = "cluster-admin"
        }
      , name = "sourcegraph-frontend"
      }
    , rules =
      [ { apiGroups = [ "" ]
        , resources = [ "endpoints", "services" ]
        , verbs = [ "get", "list", "watch" ]
        }
      ]
    }
  , RoleBinding.sourcegraph-frontend =
    { apiVersion = "rbac.authorization.k8s.io/v1"
    , kind = "RoleBinding"
    , metadata =
      { labels =
        { category = "rbac"
        , deploy = "sourcegraph"
        , sourcegraph-component = "frontend"
        , sourcegraph-resource-requires = "cluster-admin"
        }
      , name = "sourcegraph-frontend"
      }
    , roleRef = { apiGroup = "", kind = "Role", name = "sourcegraph-frontend" }
    , subjects = [ { kind = "ServiceAccount", name = "sourcegraph-frontend" } ]
    }
  , Service =
    { sourcegraph-frontend =
      { apiVersion = "v1"
      , kind = "Service"
      , metadata =
        { annotations =
          { `prometheus.io/port` = "6060"
          , `sourcegraph.prometheus/scrape` = "true"
          }
        , labels =
          { app = "sourcegraph-frontend"
          , deploy = "sourcegraph"
          , sourcegraph-component = "frontend"
          , sourcegraph-resource-requires = "no-cluster-admin"
          }
        , name = "sourcegraph-frontend"
        }
      , spec =
        { ports = [ { name = "http", port = 30080, targetPort = "http" } ]
        , selector.app = "sourcegraph-frontend"
        , type = "ClusterIP"
        }
      }
    , sourcegraph-frontend-internal =
      { apiVersion = "v1"
      , kind = "Service"
      , metadata =
        { labels =
          { app = "sourcegraph-frontend"
          , deploy = "sourcegraph"
          , sourcegraph-component = "frontend"
          , sourcegraph-resource-requires = "no-cluster-admin"
          }
        , name = "sourcegraph-frontend-internal"
        }
      , spec =
        { ports =
          [ { name = "http-internal", port = 80, targetPort = "http-internal" }
          ]
        , selector.app = "sourcegraph-frontend"
        , type = "ClusterIP"
        }
      }
    }
  , ServiceAccount.sourcegraph-frontend =
    { apiVersion = "v1"
    , imagePullSecrets = [ { name = "docker-registry" } ]
    , kind = "ServiceAccount"
    , metadata =
      { labels =
        { category = "rbac"
        , deploy = "sourcegraph"
        , sourcegraph-component = "frontend"
        , sourcegraph-resource-requires = "no-cluster-admin"
        }
      , name = "sourcegraph-frontend"
      }
    }
  }
, Indexed-Search =
  { Service =
    { indexed-search =
      { apiVersion = "v1"
      , kind = "Service"
      , metadata =
        { annotations =
          { description =
              "Headless service that provides a stable network identity for the indexed-search stateful set."
          , `prometheus.io/port` = "6070"
          , `sourcegraph.prometheus/scrape` = "true"
          }
        , labels =
          { app = "indexed-search"
          , deploy = "sourcegraph"
          , sourcegraph-component = "indexed-search"
          , sourcegraph-resource-requires = "no-cluster-admin"
          }
        , name = "indexed-search"
        }
      , spec =
        { clusterIP = "None"
        , ports = [ { port = 6070 } ]
        , selector.app = "indexed-search"
        , type = "ClusterIP"
        }
      }
    , indexed-search-indexer =
      { apiVersion = "v1"
      , kind = "Service"
      , metadata =
        { annotations =
          { description =
              "Headless service that provides a stable network identity for the indexed-search stateful set."
          , `prometheus.io/port` = "6072"
          , `sourcegraph.prometheus/scrape` = "true"
          }
        , labels =
          { app = "indexed-search-indexer"
          , deploy = "sourcegraph"
          , sourcegraph-component = "indexed-search"
          , sourcegraph-resource-requires = "no-cluster-admin"
          }
        , name = "indexed-search-indexer"
        }
      , spec =
        { clusterIP = "None"
        , ports = [ { port = 6072, targetPort = 6072 } ]
        , selector.app = "indexed-search"
        , type = "ClusterIP"
        }
      }
    }
  , StatefulSet.indexed-search =
    { apiVersion = "apps/v1"
    , kind = "StatefulSet"
    , metadata =
      { annotations.description = "Backend for indexed text search operations."
      , labels =
        { deploy = "sourcegraph"
        , sourcegraph-component = "indexed-search"
        , sourcegraph-resource-requires = "no-cluster-admin"
        }
      , name = "indexed-search"
      }
    , spec =
      { replicas = 1
      , revisionHistoryLimit = 10
      , selector.matchLabels.app = "indexed-search"
      , serviceName = "indexed-search"
      , template =
        { metadata.labels = { app = "indexed-search", deploy = "sourcegraph" }
        , spec =
          { containers =
            [ { env = None <>
              , image =
                  "index.docker.io/sourcegraph/indexed-searcher:3.19.2@sha256:d2e87635cf48c4c5d744962540751022013359bc00a9fb8e1ec2464cc6a0a2b8"
              , name = "zoekt-webserver"
              , ports = [ { containerPort = 6070, name = "http" } ]
              , readinessProbe = Some
                { failureThreshold = 3
                , httpGet =
                  { path = "/healthz", port = "http", scheme = "HTTP" }
                , periodSeconds = 5
                , timeoutSeconds = 5
                }
              , resources =
                { limits = { cpu = "2", memory = "4G" }
                , requests = { cpu = "500m", memory = "2G" }
                }
              , terminationMessagePolicy = "FallbackToLogsOnError"
              , volumeMounts = [ { mountPath = "/data", name = "data" } ]
              }
            , { env = None <>
              , image =
                  "index.docker.io/sourcegraph/search-indexer:3.19.2@sha256:7ddeb4d06a89e086506f08d9a114186260c7fa6c242e59be8c28b505d506047a"
              , name = "zoekt-indexserver"
              , ports = [ { containerPort = 6072, name = "index-http" } ]
              , readinessProbe =
                  None
                    { failureThreshold : Natural
                    , httpGet : { path : Text, port : Text, scheme : Text }
                    , periodSeconds : Natural
                    , timeoutSeconds : Natural
                    }
              , resources =
                { limits = { cpu = "8", memory = "8G" }
                , requests = { cpu = "4", memory = "4G" }
                }
              , terminationMessagePolicy = "FallbackToLogsOnError"
              , volumeMounts = [ { mountPath = "/data", name = "data" } ]
              }
            ]
          , securityContext.runAsUser = 0
          , volumes = [ { name = "data" } ]
          }
        }
      , updateStrategy.type = "RollingUpdate"
      , volumeClaimTemplates =
        [ { apiVersion = "apps/v1"
          , kind = "PersistentVolumeClaim"
          , metadata = { labels.deploy = "sourcegraph", name = "data" }
          , spec =
            { accessModes = [ "ReadWriteOnce" ]
            , resources.requests.storage = "200Gi"
            , storageClassName = "sourcegraph"
            }
          }
        ]
      }
    }
  }
}
