package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

func main() {

	files := os.Args[1:]

	if len(files) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			files = append(files, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			panic(err)
		}
	}

	baseFile, others := files[0], files[1:]

	b, err := ioutil.ReadFile(baseFile)
	if err != nil {
		panic(err)
	}

	var base map[interface{}]interface{}
	if err = yaml.Unmarshal(b, &base); err != nil {
		panic(err)
	}
	stringBase := turnToString(base)

	for _, file := range others {
		b, err = ioutil.ReadFile(file)
		if err != nil {
			panic(err)
		}

		var patch map[interface{}]interface{}
		if err = yaml.Unmarshal(b, &patch); err != nil {
			panic(err)
		}
		stringPatch := turnToString(patch)

		applyPatch(stringBase, stringPatch)
	}

	b, err = yaml.Marshal(stringBase)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}

func turnToString(patch map[interface{}]interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	for k, v := range patch {
		switch vt := v.(type) {
		case []interface{}:
			for i, vv := range vt {
				switch vvt := vv.(type) {
				case map[interface{}]interface{}:
					vt[i] = turnToString(vvt)
				default:
					vt[i] = vvt
				}
			}
			m[k.(string)] = vt
		case map[interface{}]interface{}:
			m[k.(string)] = turnToString(vt)
		default:
			m[k.(string)] = vt
		}
	}
	return m
}

func applyPatch(base, patch map[string]interface{}) {
	for k, v := range patch {
		switch vt := v.(type) {
		case []interface{}:
			vv, ok := base[k].([]interface{})
			if !ok {
				vv = make([]interface{}, len(vt))
			}
			for i, vvv := range vt {
				switch vvvt := vvv.(type) {
				case map[string]interface{}:
					var vvvv map[string]interface{}
					if len(vv) <= i {
						vvvv = make(map[string]interface{})
						vv = append(vv, vvvv)
					} else {
						vvvv, ok = vv[i].(map[string]interface{})
						if !ok {
							vvvv = make(map[string]interface{})
							vv[i] = vvvv
						}
					}
					applyPatch(vvvv, vvvt)
				default:
					if len(vv) <= i {
						vv = append(vv, vvvt)
					} else {
						vv[i] = vvvt
					}
				}
			}
			base[k] = vv
		case map[string]interface{}:
			vv, ok := base[k].(map[string]interface{})
			if !ok {
				vv = make(map[string]interface{})
				base[k] = vv
			}
			applyPatch(vv, vt)
		default:
			base[k] = vt
		}
	}
}
