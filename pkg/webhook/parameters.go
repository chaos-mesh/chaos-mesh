package webhook

// Parameters parameters
type Parameters struct {
	Addr                string // addr to serve on
	CertFile            string // path to the x509 certificate for https
	KeyFile             string // path to the x509 private key matching `CertFile`
	ConfigDirectory     string // path to sidecar injector configuration directory (contains yamls)
	AnnotationNamespace string // namespace used to scope annotations
}
