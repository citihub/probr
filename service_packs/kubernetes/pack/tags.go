package kubernetes_pack

var (
	tags = map[string][]string{
		"@probes/kubernetes":                           []string{"k-cra", "k-gen", "k-iam", "k-iaf", "k-psp"},
		"@probes/kubernetes/general":                   []string{"k-gen"},
		"@probes/kubernetes/iam":                       []string{"k-iam"},
		"@probes/kubernetes/internet_access":           []string{"k-iaf"},
		"@standard/citihub/CHC2-APPDEV135":             []string{"k-cra"},
		"@standard/citihub/CHC2-ITS120":                []string{"k-cra"},
		"@control_type/preventative":                   []string{"k-cra-001", "k-cra-002", "k-cra-003", "k-iam-001", "k-iam-002", "k-iam-003", "k-iaf-001", "k-psp-001", "k-psp-002", "k-psp-003", "k-psp-004", "k-psp-005", "k-psp-006", "k-psp-007", "k-psp-008", "k-psp-009", "k-psp-010", "k-psp-011", "k-psp-012", "k-psp-013"},
		"@standard/cis":                                []string{"k-gen", "k-psp"},
		"@standard/cis/gke":                            []string{"k-gen", "k-psp"},
		"@standard/cis/gke/v1.6.0/5.1.3":               []string{"k-gen-001", "k-iam-004"},
		"@standard/cis/gke/v1.6.0/5.6.3":               []string{"k-gen-002"},
		"@standard/cis/gke/v1.6.0/6":                   []string{"k-cra"},
		"@standard/cis/gke/v1.6.0/6.1":                 []string{"k-cra"},
		"@standard/cis/gke/v1.6.0/6.1.3":               []string{"k-cra-001"},
		"@standard/cis/gke/v1.6.0/6.1.4":               []string{"k-cra-002"},
		"@standard/cis/gke/v1.6.0/6.1.5":               []string{"k-cra-003"},
		"@standard/cis/gke/v1.6.0/6.10.1":              []string{"k-gen-003"},
		"@csp/any":                                     []string{"k-cra", "k-gen", "k-iam", "k-psp"},
		"@csp/azure":                                   []string{"k-iam-001", "k-iam-002", "k-iam-003"},
		"@probes/kubernetes/container_registry_access": []string{"k-cra"},
		"@control_type/inspection":                     []string{"k-gen-001", "k-gen-002", "k-gen-003", "k-iam-004"},
		"@standard/citihub/CHC2-IAM105":                []string{"k-gen-001", "k-iam", "k-psp", "k-iam-004"},
		"@standard/citihub/CHC2-ITS115":                []string{"k-gen-003"},
		"@standard/citihub/CHC2-SVD010":                []string{"k-iaf"},
		"@category/iam":                                []string{"k-iam"},
		"@category/internet_access":                    []string{"k-iaf"},
		"@standard/citihub":                            []string{"k-iam", "k-iaf", "k-psp"},
		"@probes/kubernetes/pod_security_policy":       []string{"k-psp"},
		"@category/pod_security_policy":                []string{"k-psp"},
		"@standard/cis/gke/v1.6.0/5":                   []string{"k-psp"},
		"@standard/cis/gke/v1.6.0/5.2":                 []string{"k-psp"},
		"@standard/cis/gke/v1.6.0/5.2.1":               []string{"k-psp-001"},
		"@standard/cis/gke/v1.6.0/5.2.2":               []string{"k-psp-002"},
		"@standard/cis/gke/v1.6.0/5.2.3":               []string{"k-psp-003"},
		"@standard/cis/gke/v1.6.0/5.2.4":               []string{"k-psp-004"},
		"@standard/cis/gke/v1.6.0/5.2.5":               []string{"k-psp-005"},
		"@standard/cis/gke/v1.6.0/5.2.6":               []string{"k-psp-006"},
		"@standard/cis/gke/v1.6.0/5.2.7":               []string{"k-psp-007", "k-psp-013"},
		"@standard/cis/gke/v1.6.0/5.2.8":               []string{"k-psp-008"},
		"@standard/cis/gke/v1.6.0/5.2.9":               []string{"k-psp-009"},
		"@standard/none/PSP-0.1":                       []string{"k-psp-012"},
	}
)
