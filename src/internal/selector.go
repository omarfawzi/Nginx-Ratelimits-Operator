package internal

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func selectorToString(sel metav1.LabelSelector) string {
	data, _ := json.Marshal(sel)
	return string(data)
}
