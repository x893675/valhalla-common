package maps

import (
	"fmt"
	"strings"
)

func ToSlice(m map[string]string) []string {
	slice := make([]string, 0, len(m))
	for k, v := range m {
		slice = append(slice, fmt.Sprintf("%s=%s", k, v))
	}
	return slice
}

func FromString(data string, sep string) map[string]string {
	list := strings.Split(data, sep)
	return FromSlice(list)
}

func FromSlice(data []string) map[string]string {
	m := make(map[string]string)
	for _, l := range data {
		if l != "" {
			kv := strings.SplitN(l, "=", 2)
			if len(kv) == 2 {
				m[kv[0]] = kv[1]
			}
		}
	}
	return m
}

func Merge(ms ...map[string]string) map[string]string {
	res := map[string]string{}
	for _, m := range ms {
		for k, v := range m {
			res[k] = v
		}
	}
	return res
}

func DeepMerge(dst, src *map[string]interface{}) {
	for srcK, srcV := range *src {
		dstV, ok := (*dst)[srcK]
		if !ok {
			continue
		}
		dV, ok := dstV.(map[string]interface{})
		// dstV is string type
		if !ok {
			(*dst)[srcK] = srcV
			continue
		}
		sV, ok := srcV.(map[string]interface{})
		if !ok {
			continue
		}
		DeepMerge(&dV, &sV)
		(*dst)[srcK] = dV
	}
}

func GetFromKeys(m map[string]string, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != "" {
			return v
		}
	}
	return ""
}

func SetKeys(m map[string]string, keys []string, value string) map[string]string {
	for _, v := range keys {
		m[v] = value
	}
	return m
}
