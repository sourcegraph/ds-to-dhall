{ Frontend :
    { Deployment :
        { sourcegraph-frontend :
            { apiVersion : Text
            , kind : Text
            , metadata :
                { annotations : { description : Text }
                , labels :
                    { `app.kubernetes.io/component` : Text
                    , deploy : Text
                    , sourcegraph-resource-requires : Text
                    }
                , name : Text
                }
            , spec :
                { minReadySeconds : Natural
                , replicas : Natural
                , revisionHistoryLimit : Natural
                , selector : { matchLabels : { app : Text } }
                , strategy :
                    { rollingUpdate :
                        { maxSurge : Natural, maxUnavailable : Natural }
                    , type : Text
                    }
                , template :
                    { metadata : { labels : { app : Text, deploy : Text } }
                    , spec :
                        { containers :
                            { frontend :
                                { args : List Text
                                , env :
                                    { CACHE_DIR : { name : Text, value : Text }
                                    , GRAFANA_SERVER_URL :
                                        { name : Text, value : Text }
                                    , JAEGER_SERVER_URL :
                                        { name : Text, value : Text }
                                    , PGDATABASE : { name : Text, value : Text }
                                    , PGHOST : { name : Text, value : Text }
                                    , PGPORT : { name : Text, value : Text }
                                    , PGSSLMODE : { name : Text, value : Text }
                                    , PGUSER : { name : Text, value : Text }
                                    , POD_NAME :
                                        { name : Text
                                        , valueFrom :
                                            { fieldRef : { fieldPath : Text } }
                                        }
                                    , PRECISE_CODE_INTEL_BUNDLE_MANAGER_URL :
                                        { name : Text, value : Text }
                                    , PROMETHEUS_URL :
                                        { name : Text, value : Text }
                                    , SRC_GIT_SERVERS :
                                        { name : Text, value : Text }
                                    }
                                , image :
                                    < asText : Text
                                    | asRecord :
                                        { name : Text
                                        , registry : Text
                                        , sha256 : Text
                                        , version : Text
                                        }
                                    >
                                , livenessProbe :
                                    { httpGet :
                                        { path : Text
                                        , port : Text
                                        , scheme : Text
                                        }
                                    , initialDelaySeconds : Natural
                                    , timeoutSeconds : Natural
                                    }
                                , name : Text
                                , ports :
                                    { http :
                                        { containerPort : Natural, name : Text }
                                    , http-internal :
                                        { containerPort : Natural, name : Text }
                                    }
                                , readinessProbe :
                                    { httpGet :
                                        { path : Text
                                        , port : Text
                                        , scheme : Text
                                        }
                                    , periodSeconds : Natural
                                    , timeoutSeconds : Natural
                                    }
                                , resources :
                                    { limits : { cpu : Text, memory : Text }
                                    , requests : { cpu : Text, memory : Text }
                                    }
                                , terminationMessagePolicy : Text
                                , volumeMounts :
                                    { cache-ssd :
                                        { mountPath : Text, name : Text }
                                    }
                                }
                            , jaeger-agent :
                                { args : List Text
                                , env :
                                    { POD_NAME :
                                        { name : Text
                                        , valueFrom :
                                            { fieldRef :
                                                { apiVersion : Text
                                                , fieldPath : Text
                                                }
                                            }
                                        }
                                    }
                                , image :
                                    { name : Text
                                    , registry : Text
                                    , sha256 : Text
                                    , version : Text
                                    }
                                , name : Text
                                , ports :
                                    List
                                      { containerPort : Natural
                                      , protocol : Text
                                      }
                                , resources :
                                    { limits : { cpu : Text, memory : Text }
                                    , requests : { cpu : Text, memory : Text }
                                    }
                                }
                            }
                        , securityContext : { runAsUser : Natural }
                        , serviceAccountName : Text
                        , volumes :
                            { cache-ssd : { emptyDir : {}, name : Text } }
                        }
                    }
                }
            }
        }
    , Ingress :
        { sourcegraph-frontend :
            { apiVersion : Text
            , kind : Text
            , metadata :
                { annotations :
                    { `kubernetes.io/ingress.class` : Text
                    , `nginx.ingress.kubernetes.io/proxy-body-size` : Text
                    }
                , labels :
                    { app : Text
                    , `app.kubernetes.io/component` : Text
                    , deploy : Text
                    , sourcegraph-resource-requires : Text
                    }
                , name : Text
                }
            , spec :
                { rules :
                    List
                      { http :
                          { paths :
                              List
                                { backend :
                                    { serviceName : Text
                                    , servicePort : Natural
                                    }
                                , path : Text
                                }
                          }
                      }
                }
            }
        }
    , Role :
        { sourcegraph-frontend :
            { apiVersion : Text
            , kind : Text
            , metadata :
                { labels :
                    { `app.kubernetes.io/component` : Text
                    , category : Text
                    , deploy : Text
                    , sourcegraph-resource-requires : Text
                    }
                , name : Text
                }
            , rules :
                List
                  { apiGroups : List Text
                  , resources : List Text
                  , verbs : List Text
                  }
            }
        }
    , RoleBinding :
        { sourcegraph-frontend :
            { apiVersion : Text
            , kind : Text
            , metadata :
                { labels :
                    { `app.kubernetes.io/component` : Text
                    , category : Text
                    , deploy : Text
                    , sourcegraph-resource-requires : Text
                    }
                , name : Text
                }
            , roleRef : { apiGroup : Text, kind : Text, name : Text }
            , subjects : { sourcegraph-frontend : { kind : Text, name : Text } }
            }
        }
    , Service :
        { sourcegraph-frontend :
            { apiVersion : Text
            , kind : Text
            , metadata :
                { annotations :
                    { `prometheus.io/port` : Text
                    , `sourcegraph.prometheus/scrape` : Text
                    }
                , labels :
                    { app : Text
                    , `app.kubernetes.io/component` : Text
                    , deploy : Text
                    , sourcegraph-resource-requires : Text
                    }
                , name : Text
                }
            , spec :
                { ports :
                    { http : { name : Text, port : Natural, targetPort : Text }
                    }
                , selector : { app : Text }
                , type : Text
                }
            }
        , sourcegraph-frontend-internal :
            { apiVersion : Text
            , kind : Text
            , metadata :
                { labels :
                    { app : Text
                    , `app.kubernetes.io/component` : Text
                    , deploy : Text
                    , sourcegraph-resource-requires : Text
                    }
                , name : Text
                }
            , spec :
                { ports :
                    { http-internal :
                        { name : Text, port : Natural, targetPort : Text }
                    }
                , selector : { app : Text }
                , type : Text
                }
            }
        }
    , ServiceAccount :
        { sourcegraph-frontend :
            { apiVersion : Text
            , imagePullSecrets : { docker-registry : { name : Text } }
            , kind : Text
            , metadata :
                { labels :
                    { `app.kubernetes.io/component` : Text
                    , category : Text
                    , deploy : Text
                    , sourcegraph-resource-requires : Text
                    }
                , name : Text
                }
            }
        }
    }
, Indexed-Search :
    { Service :
        { indexed-search :
            { apiVersion : Text
            , kind : Text
            , metadata :
                { annotations :
                    { description : Text
                    , `prometheus.io/port` : Text
                    , `sourcegraph.prometheus/scrape` : Text
                    }
                , labels :
                    { app : Text
                    , `app.kubernetes.io/component` : Text
                    , deploy : Text
                    , sourcegraph-resource-requires : Text
                    }
                , name : Text
                }
            , spec :
                { clusterIP : Text
                , ports : List { port : Natural }
                , selector : { app : Text }
                , type : Text
                }
            }
        , indexed-search-indexer :
            { apiVersion : Text
            , kind : Text
            , metadata :
                { annotations :
                    { description : Text
                    , `prometheus.io/port` : Text
                    , `sourcegraph.prometheus/scrape` : Text
                    }
                , labels :
                    { app : Text
                    , `app.kubernetes.io/component` : Text
                    , deploy : Text
                    , sourcegraph-resource-requires : Text
                    }
                , name : Text
                }
            , spec :
                { clusterIP : Text
                , ports :
                    List (Optional { port : Natural, targetPort : Natural })
                , selector : { app : Text }
                , type : Text
                }
            }
        }
    , StatefulSet :
        { indexed-search :
            { apiVersion : Text
            , kind : Text
            , metadata :
                { annotations : { description : Text }
                , labels :
                    { `app.kubernetes.io/component` : Text
                    , deploy : Text
                    , sourcegraph-resource-requires : Text
                    }
                , name : Text
                }
            , spec :
                { replicas : Natural
                , revisionHistoryLimit : Natural
                , selector : { matchLabels : { app : Text } }
                , serviceName : Text
                , template :
                    { metadata : { labels : { app : Text, deploy : Text } }
                    , spec :
                        { containers :
                            { zoekt-indexserver :
                                { env : Optional <>
                                , image :
                                    { name : Text
                                    , registry : Text
                                    , sha256 : Text
                                    , version : Text
                                    }
                                , name : Text
                                , ports :
                                    { index-http :
                                        { containerPort : Natural, name : Text }
                                    }
                                , resources :
                                    { limits : { cpu : Text, memory : Text }
                                    , requests : { cpu : Text, memory : Text }
                                    }
                                , terminationMessagePolicy : Text
                                , volumeMounts :
                                    { data : { mountPath : Text, name : Text } }
                                }
                            , zoekt-webserver :
                                { env : Optional <>
                                , image :
                                    { name : Text
                                    , registry : Text
                                    , sha256 : Text
                                    , version : Text
                                    }
                                , name : Text
                                , ports :
                                    { http :
                                        { containerPort : Natural, name : Text }
                                    }
                                , readinessProbe :
                                    { failureThreshold : Natural
                                    , httpGet :
                                        { path : Text
                                        , port : Text
                                        , scheme : Text
                                        }
                                    , periodSeconds : Natural
                                    , timeoutSeconds : Natural
                                    }
                                , resources :
                                    { limits : { cpu : Text, memory : Text }
                                    , requests : { cpu : Text, memory : Text }
                                    }
                                , terminationMessagePolicy : Text
                                , volumeMounts :
                                    { data : { mountPath : Text, name : Text } }
                                }
                            }
                        , securityContext : { runAsUser : Natural }
                        , volumes : { data : { name : Text } }
                        }
                    }
                , updateStrategy : { type : Text }
                , volumeClaimTemplates :
                    List
                      { apiVersion : Text
                      , kind : Text
                      , metadata : { labels : { deploy : Text }, name : Text }
                      , spec :
                          { accessModes : List Text
                          , resources : { requests : { storage : Text } }
                          , storageClassName : Text
                          }
                      }
                }
            }
        }
    }
}
