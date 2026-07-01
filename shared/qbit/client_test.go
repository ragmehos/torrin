package qbit

import "testing"

func TestStatePredicates(t *testing.T) {
	if !IsComplete(&Torrent{State: "uploading"}) {
		t.Error("uploading should be complete")
	}
	if !IsComplete(&Torrent{State: "downloading", Progress: 1.0}) {
		t.Error("progress 1.0 should be complete")
	}
	if IsComplete(&Torrent{State: "downloading", Progress: 0.5}) {
		t.Error("downloading 0.5 should not be complete")
	}
	if !IsStalled(&Torrent{State: "stalledDL", DlSpeed: 0}) {
		t.Error("stalledDL with 0 speed should be stalled")
	}
	if IsStalled(&Torrent{State: "stalledDL", DlSpeed: 100}) {
		t.Error("stalledDL with speed should not be stalled")
	}
	if !IsError(&Torrent{State: "error"}) {
		t.Error("error state should be error")
	}
}
