package services

import (
	"testing"
)

var svc *PterodactylService

func init() {
	svc = NewPterodactylService("ptla_ItoKeY0gvzIfhaLxcJxNNogOTBJKUGshAclNpYQynSF")
}

func TestPterodactylService_HasAccount(t *testing.T) {
	if !svc.HasAccount(1) {
		t.Error("expected true for ExtID=1")
	}
}

func TestPterodactylService_GetNodes(t *testing.T) {
	_, err := svc.GetNodes()
	if err != nil {
		t.Error(err)
	}
}
