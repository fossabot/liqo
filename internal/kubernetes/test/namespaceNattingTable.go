package test

import (
	v1 "github.com/liqoTech/liqo/api/namespaceNattingTable/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateNamespaceNattingTable() *v1.NamespaceNattingTable {
	return &v1.NamespaceNattingTable{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: ForeignClusterId,
		},
		Spec: v1.NamespaceNattingTableSpec{
			ClusterId: ForeignClusterId,
			NattingTable: map[string]string{
				Namespace: NattedNamespace,
			},
			DeNattingTable: map[string]string{
				NattedNamespace: Namespace,
			},
		},
	}
}
