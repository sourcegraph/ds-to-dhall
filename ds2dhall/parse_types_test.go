package ds2dhall

import (
	"fmt"
	"testing"
)

func TestParseTypes(t *testing.T) {
	typesStr := `
{ MutatingWebhook =
    ./types/io.k8s.api.admissionregistration.v1.MutatingWebhook.dhall sha256:efd9982e7e8a60db4df6ba7347a877073fa6efaed429e99569978d4d0b1cc630
, MutatingWebhookConfiguration =
    ./types/io.k8s.api.admissionregistration.v1.MutatingWebhookConfiguration.dhall sha256:9282892ce2600b6bc5e0e63635d8997d0bf87016af4e4fe9d7582fb8a186992c
, MutatingWebhookConfigurationList =
    ./types/io.k8s.api.admissionregistration.v1.MutatingWebhookConfigurationList.dhall sha256:4e079c1d2053458b54f772a5725b5358bd92ee14e0e95afc09f108fe150a17a6
, RuleWithOperations =
    ./types/io.k8s.api.admissionregistration.v1.RuleWithOperations.dhall sha256:ee46b465a52a8f80e0d085792749aaf74494d084bcd97b5f9a3f656a5bcce700
, ValidatingWebhook =
    ./types/io.k8s.api.admissionregistration.v1.ValidatingWebhook.dhall
, ValidatingWebhookConfiguration =
    ./types/io.k8s.api.admissionregistration.v1.ValidatingWebhookConfiguration.dhall sha256:a201dd0fc624d6d7a3b8729335c24c70373138e5451f5edee9da1143ed7855dc
, ValidatingWebhookConfigurationList =
    ./types/io.k8s.api.admissionregistration.v1.ValidatingWebhookConfigurationList.dhall sha256:2efc92527d8b2793b1bb31ebe887acfb13927dedf08701a9ba299af1e9577df6
, ControllerRevision =
    ./types/io.k8s.api.apps.v1.ControllerRevision.dhall sha256:2e67e44517534e2baef1133c1b47e8953a9bed4a27315c483b51d960bebe4554
, ControllerRevisionList =
    ./types/io.k8s.api.apps.v1.ControllerRevisionList.dhall sha256:209723e7a750dc2b67fb355e9847f70439fecc3d68e97ae2d20d9b0f33f1c643
, DaemonSet =
    ./types/io.k8s.api.apps.v1.DaemonSet.dhall sha256:d7c19aff4e38a8fa4f1f0a65710a21d49767f8349011039a0fcebf678201db55
, DaemonSetCondition =
    ./types/io.k8s.api.apps.v1.DaemonSetCondition.dhall sha256:10de5e5aed3f6e1721f79bd8e2f9ffcecb92658fbe7442e6eaf74c6780b4779d
, DaemonSetList =
    ./types/io.k8s.api.apps.v1.DaemonSetList.dhall
, DaemonSetSpec =
    ./types/io.k8s.api.apps.v1.DaemonSetSpec.dhall sha256:55d2f0b4500a542979ea5a83a03c49aa75b1fae6fc859ed92ed6527b271cf882
, DaemonSetStatus =
    ./types/io.k8s.api.apps.v1.DaemonSetStatus.dhall sha256:8df7934b5710e2cd7436b009df0b714ede641037aaa5d5757e257f4f285515e8
}
`
	res, err := parseTypes([]byte(typesStr))
	if err != nil {
		t.Errorf("failed to parse: %v", err)
	}

	fmt.Println(res)
}
