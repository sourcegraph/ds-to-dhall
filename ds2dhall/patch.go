package ds2dhall

import (
	"fmt"
	"sort"

	"ds-to-dhall/comkir"
)

func hasKeyWithStringValue(rec map[string]interface{}, key string) bool {
	v, ok := rec[key]
	if !ok {
		return false
	}

	_, ok = v.(string)
	return ok
}

func sortEnvVars(rec map[string]interface{}) {
	for k, v := range rec {
		lv, ok := v.([]interface{})
		if ok && len(lv) > 0 {
			xrec, ok := lv[0].(map[string]interface{})
			if k == "env" && ok && hasKeyWithStringValue(xrec, "name") {
				allGood := true
				for _, x := range lv {
					xrec, ok = x.(map[string]interface{})
					if !ok || !hasKeyWithStringValue(xrec, "name") {
						allGood = false
						break
					}
				}
				if allGood {
					sort.Slice(lv, func(i, j int) bool {
						xrec1 := lv[i].(map[string]interface{})
						name1 := xrec1["name"].(string)
						xrec2 := lv[j].(map[string]interface{})
						name2 := xrec2["name"].(string)

						return name1 < name2
					})
				}
			} else {
				for _, x := range lv {
					xrec, ok = x.(map[string]interface{})
					if ok {
						sortEnvVars(xrec)
					}
				}
			}
		} else {
			// do the recursive transform
			rv, ok := v.(map[string]interface{})
			if ok {
				sortEnvVars(rv)
			}
		}
	}
}

func patchResource(res *comkir.Resource, filename string) error {
	if res.Kind == "StatefulSet" {
		// patch statefulsets
		spec, ok := res.Contents["spec"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("resource %s is missing spec section", filename)
		}
		volumeClaimTemplates, ok := spec["volumeClaimTemplates"].([]interface{})
		if !ok {
			return fmt.Errorf("resource %s is missing volumeClaimTemplates section", filename)
		}
		for _, volumeClaimTemplate := range volumeClaimTemplates {
			vct, ok := volumeClaimTemplate.(map[string]interface{})
			if !ok {
				return fmt.Errorf("resource %s is missing volumeClaimTemplate section", filename)
			}
			vct["apiVersion"] = "v1"
			vct["kind"] = "PersistentVolumeClaim"
		}
	} else if res.Kind == "CronJob" {
		// patch cronjob
		spec, ok := res.Contents["spec"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("resource %s is missing spec section", filename)
		}
		jobTemplateSpec, ok := spec["jobTemplate"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("resource %s is missing jobTemplate section", filename)
		}

		_, ok = jobTemplateSpec["metadata"].(map[string]interface{})
		if !ok {
			jobTemplateSpec["metadata"] = make(map[string]interface{})
		}
	} else if res.Kind == "PersistentVolume" {
		// patch persistentvolume
		spec, ok := res.Contents["spec"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("resource %s is missing spec section", filename)
		}
		claimRef, ok := spec["claimRef"].(map[string]interface{})
		if ok {
			claimRef["apiVersion"] = "v1"
			claimRef["kind"] = "PersistentVolumeClaim"
		}
	}

	if res.Kind == "StatefulSet" || res.Kind == "Deployment" {
		sortEnvVars(res.Contents)
	}

	return nil
}
