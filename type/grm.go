package _type

import "github.com/xdmybl/gate-type/pkg/api/gate.xdmybl.io/v1"

type Snapshot struct {
	CaList       []v1.CaCertificate
	CertList     []v1.Certificate
	UpstreamList []v1.Upstream
	FilterList   []v1.Filter
	GatewayList  []v1.Gateway
}

func (s Snapshot) GetCaByName(name string) *v1.CaCertificate {
	for i := 0; i < len(s.CaList); i++ {
		if name == s.CaList[i].GetName() {
			return s.CaList[i].DeepCopy()
		}
	}
	return nil
}

func (s Snapshot) GetCertByName(name string) *v1.Certificate {
	for i := 0; i < len(s.CertList); i++ {
		if name == s.CertList[i].GetName() {
			return s.CertList[i].DeepCopy()
		}
	}
	return nil
}
