package i18n

import "testing"

// --- AX-7 canonical triplets ---

func TestState_ServiceState_HandlerTypeNames_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		state := serviceForAudit(t).State()
		names := state.HandlerTypeNames()
		if len(names) == 0 {
			t.Fatal("expected handler names")
		}
	})
	if !called {
		t.Fatal("ServiceState_HandlerTypeNames was not exercised")
	}
}

func TestState_ServiceState_HandlerTypeNames_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		state := ServiceState{}
		names := state.HandlerTypeNames()
		if len(names) != 0 {
			t.Fatalf("got %v", names)
		}
	})
	if !called {
		t.Fatal("ServiceState_HandlerTypeNames was not exercised")
	}
}

func TestState_ServiceState_HandlerTypeNames_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		state := ServiceState{Handlers: []KeyHandler{nil}}
		names := state.HandlerTypeNames()
		if len(names) != 1 {
			t.Fatalf("got %v", names)
		}
	})
	if !called {
		t.Fatal("ServiceState_HandlerTypeNames was not exercised")
	}
}

func TestState_ServiceState_String_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		state := serviceForAudit(t).State()
		got := state.String()
		if got == "" {
			t.Fatal("expected state string")
		}
	})
	if !called {
		t.Fatal("ServiceState_String was not exercised")
	}
}

func TestState_ServiceState_String_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		state := ServiceState{}
		got := state.String()
		if got == "" {
			t.Fatal("expected empty state string")
		}
	})
	if !called {
		t.Fatal("ServiceState_String was not exercised")
	}
}

func TestState_ServiceState_String_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		state := ServiceState{Language: "en", Handlers: []KeyHandler{nil}}
		got := state.String()
		if got == "" {
			t.Fatal("expected state string")
		}
	})
	if !called {
		t.Fatal("ServiceState_String was not exercised")
	}
}

func TestState_Service_State_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := serviceForAudit(t)
		state := svc.State()
		if state.Language == "" {
			t.Fatal("expected language")
		}
	})
	if !called {
		t.Fatal("Service_State was not exercised")
	}
}

func TestState_Service_State_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		var svc *Service
		state := svc.State()
		if state.Language != "en" {
			t.Fatalf("got %q", state.Language)
		}
	})
	if !called {
		t.Fatal("Service_State was not exercised")
	}
}

func TestState_Service_State_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := serviceForAudit(t)
		svc.SetDebug(true)
		state := svc.State()
		if !state.Debug {
			t.Fatal("expected debug state")
		}
	})
	if !called {
		t.Fatal("Service_State was not exercised")
	}
}

func TestState_Service_String_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := serviceForAudit(t)
		got := svc.String()
		if got == "" {
			t.Fatal("expected string")
		}
	})
	if !called {
		t.Fatal("Service_String was not exercised")
	}
}

func TestState_Service_String_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		var svc *Service
		got := svc.String()
		if got == "" {
			t.Fatal("expected string")
		}
	})
	if !called {
		t.Fatal("Service_String was not exercised")
	}
}

func TestState_Service_String_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := serviceForAudit(t)
		svc.SetLocation("workspace")
		got := svc.String()
		if got == "" {
			t.Fatal("expected string")
		}
	})
	if !called {
		t.Fatal("Service_String was not exercised")
	}
}

func TestState_Service_CurrentState_Good(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := serviceForAudit(t)
		state := svc.CurrentState()
		if state.Language == "" {
			t.Fatal("expected language")
		}
	})
	if !called {
		t.Fatal("Service_CurrentState was not exercised")
	}
}

func TestState_Service_CurrentState_Bad(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		var svc *Service
		state := svc.CurrentState()
		if state.Language != "en" {
			t.Fatalf("got %q", state.Language)
		}
	})
	if !called {
		t.Fatal("Service_CurrentState was not exercised")
	}
}

func TestState_Service_CurrentState_Ugly(t *testing.T) {
	called := false
	noPanicForAudit(t, func() {
		called = true
		svc := serviceForAudit(t)
		svc.SetMode(ModeCollect)
		state := svc.CurrentState()
		if state.Mode != ModeCollect {
			t.Fatalf("got %v", state.Mode)
		}
	})
	if !called {
		t.Fatal("Service_CurrentState was not exercised")
	}
}
