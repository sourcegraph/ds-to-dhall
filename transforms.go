package main

import (
	"fmt"
	"strings"
)

func hasKeyWithStringValue(rec map[string]interface{}, key string) bool {
	v, ok := rec[key]
	if !ok {
		return false
	}

	_, ok = v.(string)
	return ok
}

func str2dockerImage(dstr string) map[string]interface{} {
	// example: index.docker.io/sourcegraph/frontend:3.19.2@sha256:776606b680d7ce4a5d37451831ef2414ab10414b5e945ed5f50fe768f898d23fa
	// parts: registry: index.docker.io
	//        name: sourcegraph/frontend
	//        version: 3.19.2
	//        sha256: 776606b680d7ce4a5d37451831ef2414ab10414b5e945ed5f50fe768f898d23fa

	di := make(map[string]interface{})
	xs := strings.Split(dstr, "@sha256:")

	if len(xs) == 2 {
		di["sha256"] = xs[1]
		dstr = xs[0]
	}

	xs = strings.Split(dstr, ":")
	if len(xs) == 2 {
		di["version"] = xs[1]
		dstr = xs[0]
	}

	xs = strings.Split(dstr, "/")
	if len(xs) > 1 {
		di["registry"] = xs[0]
		di["name"] = strings.Join(xs[1:], "/")
	} else {
		di["name"] = dstr
	}

	return di
}

func transformDockerImageSpec(rec map[string]interface{}) map[string]interface{} {
	mrec := make(map[string]interface{})
	for k, v := range rec {
		mrec[k] = v
		dstr, ok := v.(string)

		if k == "image" && ok {
			mrec[k] = str2dockerImage(dstr)
		}

		rv, ok := mrec[k].(map[string]interface{})
		if ok {
			xrec := transformDockerImageSpec(rv)
			mrec[k] = xrec
		}
	}

	return mrec
}

func transformList2Record(rec map[string]interface{}) (map[string]interface{}, error) {
	mrec := make(map[string]interface{})

	for k, v := range rec {
		mrec[k] = v

		// check if it's a list with records in it that have "name" key
		lv, ok := v.([]interface{})
		if ok && len(lv) > 0 {
			xrec, ok := lv[0].(map[string]interface{})
			if ok && hasKeyWithStringValue(xrec, "name") {
				rv := make(map[string]interface{})
				for _, x := range lv {
					xrec, ok = x.(map[string]interface{})
					if !ok || !hasKeyWithStringValue(xrec, "name") {
						return nil, fmt.Errorf("expected list of records that have `name` key %+v", lv)
					}
					name := xrec["name"].(string)
					rv[name] = xrec
				}
				mrec[k] = rv
			}
		}

		// do the recursive transform
		rv, ok := mrec[k].(map[string]interface{})
		if ok {
			xrec, err := transformList2Record(rv)
			if err != nil {
				return nil, err
			}
			mrec[k] = xrec
		}
	}
	return mrec, nil
}
