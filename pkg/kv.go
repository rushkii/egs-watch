package pkg

type KeyValue struct {
	Key   string
	Value string
}

func GetKVFromArray(key string, objects []KeyValue) string {
	for _, obj := range objects {
		if obj.Key == key {
			return obj.Value
		}
	}

	return ""
}
