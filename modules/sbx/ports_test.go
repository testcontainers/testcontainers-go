package sbx

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func testPortManager(t *testing.T, f *fakeSandbox) *portManager {
	t.Helper()
	m, err := newPortManager(context.Background(), f.runtime(t), &http.Client{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = m.close(ctx)
	})
	return m
}

func TestPortReservationPrecedesCreateAndUsesExactTuple(t *testing.T) {
	f := newFakeSandbox(t)
	m := testPortManager(t, f)
	prepared, err := m.prepareCreate(context.Background(), portBindings{"8080/tcp": {{HostIP: "127.0.0.1", HostPort: "39131"}}, "5353/udp": {{HostPort: "39132"}}})
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Bindings["8080/tcp"][0].HostPort != "39131" || prepared.Bindings["5353/udp"][0].HostPort != "39132" {
		t.Fatalf("unexpected forwarded bindings: %+v", prepared.Bindings)
	}
	f.record("docker POST /containers/create")
	steps := f.snapshot()
	for _, spec := range []string{"127.0.0.1:39131:39131/tcp4", "127.0.0.1:39132:39132/udp4"} {
		published := false
		for _, step := range steps {
			if strings.Contains(step, "--publish "+spec) {
				published = true
				break
			}
			if step == "docker POST /containers/create" {
				break
			}
		}
		if !published {
			t.Fatalf("%s not reserved before Docker create: %v", spec, steps)
		}
	}
	if err := m.bindCreate(prepared.Token, "container-1", "sample"); err != nil {
		t.Fatal(err)
	}
	finish := m.beginStop("container-1")
	if err := finish(context.Background(), true); err != nil {
		t.Fatal(err)
	}
	steps = f.snapshot()
	for _, spec := range []string{"127.0.0.1:39131:39131/tcp4", "127.0.0.1:39132:39132/udp4"} {
		found := false
		for _, step := range steps {
			if strings.Contains(step, "--unpublish "+spec) {
				found = true
			}
		}
		if !found {
			t.Fatalf("exact tuple not unpublished: %s; %v", spec, steps)
		}
	}
}

func TestRandomTCPAndUDPBindingsArePublishedBeforeCreate(t *testing.T) {
	f := newFakeSandbox(t)
	m := testPortManager(t, f)
	input := portBindings{"8080/tcp": {{HostPort: "0"}}, "5353/udp": {{HostPort: ""}}}
	prepared, err := m.prepareCreate(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"8080/tcp", "5353/udp"} {
		assigned := prepared.Bindings[key][0]
		if assigned.HostIP != "0.0.0.0" || assigned.HostPort == "" || assigned.HostPort == "0" {
			t.Fatalf("no usable host binding for %s: %+v", key, assigned)
		}
	}
	if input["8080/tcp"][0].HostPort != "0" || input["5353/udp"][0].HostPort != "" {
		t.Fatalf("caller bindings mutated: %+v", input)
	}
	f.mu.Lock()
	count := len(f.ports)
	f.mu.Unlock()
	if count != 2 {
		t.Fatalf("expected both physical host reservations before create; got %d", count)
	}
}

func TestRejectedBindingsHaveNoPublishSideEffect(t *testing.T) {
	cases := []portBindings{
		{"8080/tcp": {{HostIP: "192.0.2.4", HostPort: "5000"}}},
		{"8080/tcp": {{HostPort: "5000-5010"}}},
		{"8080/sctp": {{HostPort: "5000"}}},
		{"8080/tcp": {{HostPort: "5000"}, {HostPort: "5000"}}},
	}
	for _, bindings := range cases {
		t.Run(fmt.Sprint(bindings), func(t *testing.T) {
			f := newFakeSandbox(t)
			m := testPortManager(t, f)
			if _, err := m.prepareCreate(context.Background(), bindings); err == nil {
				t.Fatal("expected rejection")
			}
			for _, step := range f.snapshot() {
				if strings.Contains(step, "--publish") {
					t.Fatalf("invalid request published a port: %v", f.snapshot())
				}
			}
		})
	}
}

func TestLostCreateResponseRetainsReservation(t *testing.T) {
	f := newFakeSandbox(t)
	m := testPortManager(t, f)
	prepared, err := m.prepareCreate(context.Background(), portBindings{"80/tcp": {{HostPort: "39133"}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := m.abortCreate(context.Background(), prepared.Token, true); err != nil {
		t.Fatal(err)
	}
	if _, err := m.prepareCreate(context.Background(), portBindings{"81/tcp": {{HostPort: "39133"}}}); err == nil {
		t.Fatal("reused reservation after ambiguous create and empty list")
	}
	for _, step := range f.snapshot() {
		if strings.Contains(step, "--unpublish") {
			t.Fatalf("lost create response released port: %v", f.snapshot())
		}
	}
}

func TestStopUnpublishFailureKeepsPortOwned(t *testing.T) {
	f := newFakeSandbox(t)
	m := testPortManager(t, f)
	prepared, err := m.prepareCreate(context.Background(), portBindings{"80/tcp": {{HostPort: "39134"}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := m.bindCreate(prepared.Token, "container-1", "sample"); err != nil {
		t.Fatal(err)
	}
	f.unpublishErr = errors.New("policy denied")
	finish := m.beginStop("container-1")
	if err := finish(context.Background(), true); err == nil {
		t.Fatal("expected unpublish failure")
	}
	if _, err := m.prepareCreate(context.Background(), portBindings{"81/tcp": {{HostPort: "39134"}}}); err == nil {
		t.Fatal("reassigned orphaned port")
	}
	f.unpublishErr = nil
}

func TestStopThenStartRestoresImmutablePort(t *testing.T) {
	f := newFakeSandbox(t)
	m := testPortManager(t, f)
	prepared, err := m.prepareCreate(context.Background(), portBindings{"80/tcp": {{HostPort: "39135"}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := m.bindCreate(prepared.Token, "container-1", "sample"); err != nil {
		t.Fatal(err)
	}
	if err := m.beginStop("container-1")(context.Background(), true); err != nil {
		t.Fatal(err)
	}
	finish, err := m.beginStart(context.Background(), "container-1")
	if err != nil {
		t.Fatal(err)
	}
	if err := finish(context.Background(), true); err != nil {
		t.Fatal(err)
	}
	if _, ok := f.ports["127.0.0.1:39135:39135/tcp4"]; !ok {
		t.Fatal("restart did not restore port")
	}
	if err := m.beginStop("container-1")(context.Background(), true); err != nil {
		t.Fatal(err)
	}
	f.publishErr = errors.New("address already in use")
	if _, err := m.beginStart(context.Background(), "container-1"); err == nil {
		t.Fatal("start forwarded after fixed port collision")
	}
	f.publishErr = nil
}

func TestDelayedStopEventCannotReleaseConcurrentRestart(t *testing.T) {
	f := newFakeSandbox(t)
	m := testPortManager(t, f)
	prepared, err := m.prepareCreate(context.Background(), portBindings{"80/tcp": {{HostPort: "39137"}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := m.bindCreate(prepared.Token, "container-1", "sample"); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	f.containers["container-1"] = false
	f.mu.Unlock()
	if err := m.beginStop("container-1")(context.Background(), true); err != nil {
		t.Fatal(err)
	}
	finish, err := m.beginStart(context.Background(), "container-1")
	if err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	done := make(chan struct{})
	go func() {
		close(started)
		m.handleDockerEvent(context.Background(), dockerEvent{Status: "die", ID: "container-1"})
		close(done)
	}()
	<-started
	f.mu.Lock()
	f.containers["container-1"] = true
	f.mu.Unlock()
	if err := finish(context.Background(), true); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("delayed event did not finish")
	}
	f.mu.Lock()
	_, published := f.ports["127.0.0.1:39137:39137/tcp4"]
	f.mu.Unlock()
	if !published {
		t.Fatalf("late stop event released running container port: %v", f.snapshot())
	}
}

func TestOutOfBandContainerDeletionReleasesPort(t *testing.T) {
	f := newFakeSandbox(t)
	m := testPortManager(t, f)
	prepared, err := m.prepareCreate(context.Background(), portBindings{"80/tcp": {{HostPort: "39138"}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := m.bindCreate(prepared.Token, "container-1", "sample"); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	delete(f.containers, "container-1")
	f.mu.Unlock()
	if err := m.reconcile(context.Background()); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	_, published := f.ports["127.0.0.1:39138:39138/tcp4"]
	f.mu.Unlock()
	if published {
		t.Fatalf("deleted container kept host publication: %v", f.snapshot())
	}
	if _, err := m.prepareCreate(context.Background(), portBindings{"81/tcp": {{HostPort: "39138"}}}); err != nil {
		t.Fatalf("deleted container kept lease: %v", err)
	}
}

func TestUnpublishRequiresObservedRemoval(t *testing.T) {
	for _, mode := range []string{"exit-zero-but-still-published", "list-fails"} {
		t.Run(mode, func(t *testing.T) {
			f := newFakeSandbox(t)
			broken := false
			runtime := f.runtime(t)
			runtime.run = func(ctx context.Context, args ...string) ([]byte, error) {
				if len(args) > 2 && args[2] == "--unpublish" && broken {
					f.record("sbx " + strings.Join(args, " "))
					return []byte("[]"), nil // CLI exit 0, but publication remains.
				}
				if mode == "list-fails" && broken && len(args) > 2 && args[2] == "--format" {
					return nil, errors.New("list unavailable")
				}
				return f.run(ctx, args...)
			}
			m, err := newPortManager(context.Background(), runtime, &http.Client{})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				broken = false
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				_ = m.close(ctx)
			})
			prepared, err := m.prepareCreate(context.Background(), portBindings{"80/tcp": {{HostPort: "39201"}}})
			if err != nil {
				t.Fatal(err)
			}
			if err := m.bindCreate(prepared.Token, "container-1", "sample"); err != nil {
				t.Fatal(err)
			}
			broken = true
			if err := m.beginStop("container-1")(context.Background(), true); err == nil {
				t.Fatal("exit-zero unpublish was trusted without observing removal")
			}
			broken = false
			if _, err := m.prepareCreate(context.Background(), portBindings{"81/tcp": {{HostPort: "39201"}}}); err == nil {
				t.Fatal("uncertain port was reassigned")
			}
			if err := m.beginStop("container-1")(context.Background(), true); err != nil {
				t.Fatalf("cleanup retry: %v", err)
			}
			f.mu.Lock()
			_, stillPublished := f.ports["127.0.0.1:39201:39201/tcp4"]
			f.mu.Unlock()
			if stillPublished {
				t.Fatal("cleanup retry did not remove physical publication")
			}
			if _, err := m.prepareCreate(context.Background(), portBindings{"81/tcp": {{HostPort: "39201"}}}); err == nil {
				t.Fatal("stopped container lost its immutable restart reservation")
			}
			if err := m.beginDestroy("container-1")(context.Background(), true); err != nil {
				t.Fatalf("destroy after stopped cleanup: %v", err)
			}
			if _, err := m.prepareCreate(context.Background(), portBindings{"81/tcp": {{HostPort: "39201"}}}); err != nil {
				t.Fatalf("confirmed destruction did not free port: %v", err)
			}
		})
	}
}

func TestAmbiguousFixedPortRestartRequiresProofBeforeStart(t *testing.T) {
	f := newFakeSandbox(t)
	m := testPortManager(t, f)
	prepared, err := m.prepareCreate(context.Background(), portBindings{"80/tcp": {{HostPort: "39202"}}})
	if err != nil {
		t.Fatal(err)
	}
	if err := m.bindCreate(prepared.Token, "container-1", "sample"); err != nil {
		t.Fatal(err)
	}
	if err := m.beginStop("container-1")(context.Background(), true); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	f.publishErr = errors.New("publication response lost")
	f.mu.Unlock()
	if _, err := m.beginStart(context.Background(), "container-1"); err == nil {
		t.Fatal("ambiguous reacquisition forwarded start")
	}
	f.mu.Lock()
	f.publishErr = nil
	f.mu.Unlock()
	finish, err := m.beginStart(context.Background(), "container-1")
	if err != nil {
		t.Fatalf("second start did not reestablish same fixed port: %v", err)
	}
	if err := finish(context.Background(), true); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	_, live := f.ports["127.0.0.1:39202:39202/tcp4"]
	f.mu.Unlock()
	if !live {
		t.Fatal("second start returned without exact physical publication")
	}
	record, err := m.lockContainer("container-1")
	if err != nil {
		t.Fatalf("stable container ID lost: %v", err)
	}
	record.mu.Unlock()
	for _, step := range f.snapshot() {
		if strings.Contains(step, "--publish") && !strings.Contains(step, "127.0.0.1:39202:39202/tcp4") {
			t.Fatalf("fixed port reassigned: %v", f.snapshot())
		}
	}
}

func TestOrphanAddedDuringPortListSurvivesReconciliation(t *testing.T) {
	f := newFakeSandbox(t)
	listed := make(chan struct{})
	release := make(chan struct{})
	runtime := f.runtime(t)
	runtime.run = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) > 2 && args[2] == "--format" {
			data, err := f.run(ctx, args...)
			select {
			case <-listed:
			default:
				close(listed)
				<-release
			}
			return data, err
		}
		return f.run(ctx, args...)
	}
	// No background watcher: this test controls the list barrier exactly.
	m := &portManager{runtime: runtime, creations: map[string]*containerLease{}, containers: map[string]*containerLease{}}
	first := portPublication{HostIP: "127.0.0.1", HostPort: 39203, SandboxPort: 39203, Protocol: "tcp4"}
	second := portPublication{HostIP: "127.0.0.1", HostPort: 39204, SandboxPort: 39204, Protocol: "tcp4"}
	m.retainOrphans(&containerLease{token: "first", ports: []leasedPort{{desired: first, live: true, exact: true}}})
	done := make(chan error, 1)
	go func() { done <- m.reconcileOrphans(context.Background()) }()
	<-listed
	m.retainOrphans(&containerLease{token: "second", ports: []leasedPort{{desired: second, live: true, exact: true}}})
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	m.mu.Lock()
	forbidden := m.forbiddenPortsLocked()
	m.mu.Unlock()
	if !forbidden[39204] {
		t.Fatal("orphan appended after list snapshot was forgotten")
	}
}

func TestCanceledMultiPortPrepareUsesFreshCleanupContext(t *testing.T) {
	f := newFakeSandbox(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var cleanupContextValid bool
	runtime := f.runtime(t)
	runtime.run = func(commandCtx context.Context, args ...string) ([]byte, error) {
		if len(args) > 2 && args[2] == "--publish" && strings.Contains(args[3], "39206") {
			cancel()
			return nil, context.Canceled
		}
		if len(args) > 2 && args[2] == "--unpublish" {
			cleanupContextValid = commandCtx.Err() == nil
		}
		return f.run(commandCtx, args...)
	}
	m, err := newPortManager(context.Background(), runtime, &http.Client{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), time.Second)
		defer closeCancel()
		_ = m.close(closeCtx)
	})
	if _, err := m.prepareCreate(ctx, portBindings{"80/tcp": {{HostPort: "39205"}}, "81/tcp": {{HostPort: "39206"}}}); err == nil {
		t.Fatal("canceled second publication accepted")
	}
	if !cleanupContextValid {
		t.Fatal("first publication was not rolled back with fresh context")
	}
	f.mu.Lock()
	_, stillLive := f.ports["127.0.0.1:39205:39205/tcp4"]
	f.mu.Unlock()
	if stillLive {
		t.Fatal("first publication remained after canceled prepare")
	}
}

func TestOrphanCleanupRetriesUntilRemovalIsConfirmed(t *testing.T) {
	f := newFakeSandbox(t)
	publication := portPublication{HostIP: "127.0.0.1", HostPort: 39209, SandboxPort: 39209, Protocol: "tcp4"}
	f.mu.Lock()
	f.ports[publication.spec()] = publication
	f.unpublishErr = errors.New("temporary policy denial")
	f.mu.Unlock()
	m := &portManager{runtime: f.runtime(t), creations: map[string]*containerLease{}, containers: map[string]*containerLease{}}
	m.retainOrphans(&containerLease{token: "lost-create", ports: []leasedPort{{desired: publication, live: true, exact: true}}})
	if err := m.reconcileOrphans(context.Background()); err == nil {
		t.Fatal("failed orphan cleanup was treated as success")
	}
	m.mu.Lock()
	forbidden := m.forbiddenPortsLocked()
	m.mu.Unlock()
	if !forbidden[39209] {
		t.Fatal("failed orphan cleanup released ownership")
	}
	f.mu.Lock()
	f.unpublishErr = nil
	f.mu.Unlock()
	if err := m.reconcileOrphans(context.Background()); err != nil {
		t.Fatalf("orphan retry: %v", err)
	}
	m.mu.Lock()
	forbidden = m.forbiddenPortsLocked()
	m.mu.Unlock()
	if forbidden[39209] {
		t.Fatal("confirmed orphan removal kept reservation")
	}
}
