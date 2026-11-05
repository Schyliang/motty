package workbench

import (
	"context"
	"errors"
	"testing"

	"motty/internal/desktop"
	"motty/internal/desktop/rdp"
	"motty/internal/sshconn"
)

func TestVNCTLSCertificateTrustSurvivesBackupAndRestart(t *testing.T) {
	certificate := []byte("TLS certificate")
	open := func(ctx context.Context, request desktopRequest) (desktop.Session, error) {
		if err := request.Verify("pi.example.com:5900", certificate); err != nil {
			return nil, err
		}
		return &encryptedVNC{newPictureOnly()}, nil
	}
	source, host := realVNCState(t, "alice", open)
	openHost(source)
	waitUntil(t, source, "TLS identity question", func() bool { return source.modal == modalHostKey })
	source.Click(hitHostKeyTrust)
	session, _ := source.sessionByIDPointer(source.activeSessionID)
	waitUntil(t, source, "TLS desktop", func() bool { return session.desk != nil })
	document, err := source.backupDocument()
	if err != nil {
		t.Fatal(err)
	}
	source.Close()

	options := testOptions(t)
	options.Invalidate = func() {}
	target := newTestState(t, options)
	if _, err := target.applyBackup(document); err != nil {
		t.Fatal(err)
	}
	target.Close()
	state := newTestState(t, options)
	state.connectionReconnect, state.dialDesktop = reconnectOff, open
	saved, ok := state.hostByCatalogID(host.CatalogID)
	if !ok || saved.vncFingerprint("pi.example.com:5900", vncCertificateType) != rdp.Fingerprint(certificate) {
		t.Fatal("backup and disk reload lost the TLS certificate pin")
	}
	state.selectedHostID = saved.ID
	state.Click(hitHosts)
	openHost(state)
	reconnected, _ := state.sessionByIDPointer(state.activeSessionID)
	waitUntil(t, state, "TLS desktop after restart", func() bool { return reconnected.desk != nil })
	if state.modal == modalHostKey {
		t.Fatal("saved TLS certificate prompted again after restart")
	}

	certificate = []byte("changed TLS certificate")
	reconnected.desk.session.Close()
	state.reconnect(reconnected)
	waitUntil(t, state, "changed TLS certificate after restart", func() bool { return reconnected.phase == sshconn.PhaseFailed })
	if !errors.Is(reconnected.err, sshconn.ErrHostKeyChanged) || state.modal == modalHostKey {
		t.Fatalf("changed TLS certificate was not refused: %v", reconnected.err)
	}
	saved, ok = state.hostByCatalogID(host.CatalogID)
	if !ok || saved.vncFingerprint("pi.example.com:5900", vncCertificateType) != rdp.Fingerprint([]byte("TLS certificate")) {
		t.Fatal("refusing the changed certificate replaced the saved pin")
	}
}

func TestVNCTLSCertificateTrustOnceDoesNotSurviveRestart(t *testing.T) {
	open := func(ctx context.Context, request desktopRequest) (desktop.Session, error) {
		if err := request.Verify("pi.example.com:5900", []byte("TLS certificate")); err != nil {
			return nil, err
		}
		return &encryptedVNC{newPictureOnly()}, nil
	}
	state, host := realVNCState(t, "alice", open)
	if err := state.saveWithResult(); err != nil {
		t.Fatal(err)
	}
	options := state.options
	openHost(state)
	waitUntil(t, state, "TLS identity question", func() bool { return state.modal == modalHostKey })
	state.Click(hitHostKeyOnce)
	session, _ := state.sessionByIDPointer(state.activeSessionID)
	waitUntil(t, state, "one-time TLS desktop", func() bool { return session.desk != nil })
	state.Close()

	restarted := newTestState(t, options)
	restarted.connectionReconnect, restarted.dialDesktop = reconnectOff, open
	saved, ok := restarted.hostByCatalogID(host.CatalogID)
	if !ok || len(saved.VNCKeys) != 0 {
		t.Fatal("one-time TLS trust saved an identity pin")
	}
	restarted.selectedHostID = saved.ID
	restarted.Click(hitHosts)
	openHost(restarted)
	waitUntil(t, restarted, "TLS question after one-time trust", func() bool { return restarted.modal == modalHostKey })
	restarted.Click(hitHostKeyReject)
}

func TestVNCTLSCertificateSaveFailurePreventsAuthentication(t *testing.T) {
	for _, scenario := range []string{"save blocked", "host deleted", "another certificate saved"} {
		t.Run(scenario, func(t *testing.T) {
			started := make(chan struct{}, 1)
			state, host := realVNCState(t, "alice", func(ctx context.Context, request desktopRequest) (desktop.Session, error) {
				if err := request.Verify("pi.example.com:5900", []byte("TLS certificate")); err != nil {
					return nil, err
				}
				started <- struct{}{}
				return &encryptedVNC{newPictureOnly()}, nil
			})
			if err := state.saveWithResult(); err != nil {
				t.Fatal(err)
			}
			openHost(state)
			session, _ := state.sessionByIDPointer(state.activeSessionID)
			waitUntil(t, state, "TLS identity question", func() bool { return state.modal == modalHostKey })
			if scenario == "save blocked" {
				state.pauseSave++
				defer func() { state.pauseSave-- }()
			} else {
				other := newTestState(t, state.options)
				current, ok := other.hostByCatalogID(host.CatalogID)
				if !ok {
					t.Fatal("pending TLS host is absent in another window")
				}
				if scenario == "host deleted" {
					for index := len(other.sessions) - 1; index >= 0; index-- {
						if other.sessions[index].host.CatalogID == current.CatalogID {
							other.closeSession(other.sessions[index].ID)
						}
					}
					other.hostActionIDs = []uint32{current.ID}
					other.confirmDeleteHosts()
					if _, ok := other.hostByCatalogID(host.CatalogID); ok {
						t.Fatalf("host deletion failed: %q", other.dialogError)
					}
				} else {
					current.VNCKeys = []vncHostKey{{Address: "pi.example.com:5900", Kind: vncCertificateType, Fingerprint: rdp.Fingerprint([]byte("another certificate"))}}
					other.hosts[other.hostIndex(current.ID)] = current
					if err := other.saveWithResult(); err != nil {
						t.Fatal(err)
					}
				}
				other.Close()
			}
			state.Click(hitHostKeyTrust)
			waitUntil(t, state, "failed TLS identity save", func() bool {
				current, ok := state.sessionByIDPointer(session.ID)
				return !ok || current.phase == sshconn.PhaseFailed || current.phase == sshconn.PhaseClosed
			})
			if scenario == "save blocked" && !errors.Is(session.err, sshconn.ErrHostKeyRejected) {
				t.Fatalf("failed certificate save returned %v", session.err)
			}
			select {
			case <-started:
				t.Fatal("authentication started after the TLS identity save failed")
			default:
			}
			if saved, ok := state.hostByCatalogID(host.CatalogID); ok && saved.vncFingerprint("pi.example.com:5900", vncCertificateType) == rdp.Fingerprint([]byte("TLS certificate")) {
				t.Fatal("failed save retained the pending TLS certificate pin")
			}
			if scenario == "another certificate saved" {
				saved, ok := state.hostByCatalogID(host.CatalogID)
				if !ok || saved.vncFingerprint("pi.example.com:5900", vncCertificateType) != rdp.Fingerprint([]byte("another certificate")) {
					t.Fatal("failed save replaced the other window's TLS certificate pin")
				}
			}
		})
	}
}
