package sbx

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
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

func TestPlanBindingsPreservesRequestedFixedPorts(t *testing.T) {
	plan := func(key, protocol string, port int, ephemeral bool) portPlan {
		return portPlan{
			containerPort: key, ephemeral: ephemeral,
			publication: portPublication{HostIP: "127.0.0.1", HostPort: port, SandboxPort: port, Protocol: protocol + "4"},
		}
	}
	cases := []struct {
		name       string
		bindings   portBindings
		protocol   string
		candidates []int
		forbidden  map[portReservation]bool
		want       []portPlan
		wantErr    bool
	}{
		{
			name:     "empty TCP before later key",
			bindings: portBindings{"80/tcp": {{HostPort: ""}}, "81/tcp": {{HostPort: "40000"}}},
			protocol: "tcp", candidates: []int{40000, 40001},
			want: []portPlan{plan("80/tcp", "tcp", 40001, true), plan("81/tcp", "tcp", 40000, false)},
		},
		{
			name:     "zero UDP before later key",
			bindings: portBindings{"80/udp": {{HostPort: "0"}}, "81/udp": {{HostPort: "40000"}}},
			protocol: "udp", candidates: []int{40000, 40001},
			want: []portPlan{plan("80/udp", "udp", 40001, true), plan("81/udp", "udp", 40000, false)},
		},
		{
			name:     "zero TCP before later binding",
			bindings: portBindings{"80/tcp": {{HostPort: "0"}, {HostPort: "40000"}}},
			protocol: "tcp", candidates: []int{40000, 40001},
			want: []portPlan{plan("80/tcp", "tcp", 40001, true), plan("80/tcp", "tcp", 40000, false)},
		},
		{
			name:     "fixed UDP allows same TCP port",
			bindings: portBindings{"80/tcp": {{HostPort: ""}}, "81/udp": {{HostPort: "40000"}}},
			protocol: "tcp", candidates: []int{40000},
			want: []portPlan{plan("80/tcp", "tcp", 40000, true), plan("81/udp", "udp", 40000, false)},
		},
		{
			name:      "external owner rejects fixed port",
			bindings:  portBindings{"80/tcp": {{HostPort: "40000"}}},
			forbidden: map[portReservation]bool{{protocol: "tcp", port: 40000}: true},
			wantErr:   true,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			original := make(portBindings, len(tt.bindings))
			for key, bindings := range tt.bindings {
				original[key] = append([]portBinding(nil), bindings...)
			}
			var originalForbidden map[portReservation]bool
			if tt.forbidden != nil {
				originalForbidden = make(map[portReservation]bool, len(tt.forbidden))
				for reservation, owned := range tt.forbidden {
					originalForbidden[reservation] = owned
				}
			}
			candidates := tt.candidates
			got, err := planBindingsUsing(context.Background(), tt.bindings, tt.forbidden, func(protocol string) (int, error) {
				if protocol != tt.protocol {
					return 0, fmt.Errorf("unexpected candidate protocol %q", protocol)
				}
				if len(candidates) == 0 {
					return 0, errors.New("candidate sequence exhausted")
				}
				port := candidates[0]
				candidates = candidates[1:]
				return port, nil
			})
			if (err != nil) != tt.wantErr {
				t.Fatalf("plan error = %v, want error %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("plans = %+v, want %+v", got, tt.want)
			}
			if len(candidates) != 0 {
				t.Fatalf("unused candidates: %v", candidates)
			}
			if !reflect.DeepEqual(tt.bindings, original) || !reflect.DeepEqual(tt.forbidden, originalForbidden) {
				t.Fatalf("planner mutated caller inputs: bindings=%+v forbidden=%+v", tt.bindings, tt.forbidden)
			}
		})
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
	if !forbidden[second.reservation()] {
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
	if !forbidden[publication.reservation()] {
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
	if forbidden[publication.reservation()] {
		t.Fatal("confirmed orphan removal kept reservation")
	}
}

func TestPortReservationsIncludeProtocol(t *testing.T) {
	f := newFakeSandbox(t)
	m := testPortManager(t, f)

	prepared, err := m.prepareCreate(context.Background(), portBindings{
		"53/tcp": {{HostPort: "1053"}},
		"53/udp": {{HostPort: "1053"}},
	})
	if err != nil {
		t.Fatalf("reserve matching TCP and UDP host ports: %v", err)
	}
	if prepared.Bindings["53/tcp"][0].HostPort != "1053" || prepared.Bindings["53/udp"][0].HostPort != "1053" {
		t.Fatalf("unexpected rewritten bindings: %+v", prepared.Bindings)
	}
	for _, spec := range []string{"127.0.0.1:1053:1053/tcp4", "127.0.0.1:1053:1053/udp4"} {
		f.mu.Lock()
		_, published := f.ports[spec]
		f.mu.Unlock()
		if !published {
			t.Fatalf("missing protocol-specific publication %s", spec)
		}
	}

	if err := m.bindCreate(prepared.Token, "container-both", "both"); err != nil {
		t.Fatal(err)
	}
	tcp, err := m.prepareCreate(context.Background(), portBindings{"80/tcp": {{HostPort: "1054"}}})
	if err != nil {
		t.Fatalf("reserve active TCP lease: %v", err)
	}
	if err := m.bindCreate(tcp.Token, "container-tcp", "tcp"); err != nil {
		t.Fatal(err)
	}
	if _, err := m.prepareCreate(context.Background(), portBindings{"80/udp": {{HostPort: "1054"}}}); err != nil {
		t.Fatalf("cross-protocol active lease reuse: %v", err)
	}
	if _, err := m.prepareCreate(context.Background(), portBindings{"81/tcp": {{HostPort: "1054"}}}); err == nil {
		t.Fatal("same-protocol active lease reuse was accepted")
	}
}

func TestEnginePortReservationIsTCPOnly(t *testing.T) {
	f := newFakeSandbox(t)
	m := testPortManager(t, f)
	if _, err := m.prepareCreate(context.Background(), portBindings{"80/udp": {{HostPort: "2375"}}}); err != nil {
		t.Fatalf("UDP port 2375 should remain available: %v", err)
	}
	if _, err := m.prepareCreate(context.Background(), portBindings{"80/tcp": {{HostPort: "2375"}}}); err == nil {
		t.Fatal("TCP port 2375 was not reserved for the inner Engine")
	}
}

func TestDockerPortDiscoveryReservesOnlyPublicProtocolTuple(t *testing.T) {
	f := newFakeSandbox(t)
	f.list = []dockerContainer{{Ports: []dockerPort{
		{PrivatePort: 8080, PublicPort: 40000, Type: "tcp"},
		{PrivatePort: 9090, Type: "tcp"},
	}}}
	m := testPortManager(t, f)

	if _, err := m.prepareCreate(context.Background(), portBindings{"80/tcp": {{HostPort: "8080"}}}); err != nil {
		t.Fatalf("private container port incorrectly reserved on host: %v", err)
	}
	if _, err := m.prepareCreate(context.Background(), portBindings{"81/tcp": {{HostPort: "9090"}}}); err != nil {
		t.Fatalf("private-only exposure incorrectly reserved on host: %v", err)
	}
	if _, err := m.prepareCreate(context.Background(), portBindings{"82/tcp": {{HostPort: "40000"}}}); err == nil {
		t.Fatal("published TCP host port was not reserved")
	}
	if _, err := m.prepareCreate(context.Background(), portBindings{"82/udp": {{HostPort: "40000"}}}); err != nil {
		t.Fatalf("published TCP host port incorrectly reserved UDP tuple: %v", err)
	}
}

func TestFailedOrphanCleanupRetainsOnlyItsProtocol(t *testing.T) {
	f := newFakeSandbox(t)
	tcpPublication := portPublication{HostIP: "127.0.0.1", HostPort: 39210, SandboxPort: 39210, Protocol: "tcp4"}
	f.mu.Lock()
	f.ports[tcpPublication.spec()] = tcpPublication
	f.unpublishErr = errors.New("temporary policy denial")
	f.mu.Unlock()
	m := &portManager{
		runtime:    f.runtime(t),
		engine:     newDockerEngine(mustParseURL(t, f.server.URL), &http.Client{}),
		creations:  map[string]*containerLease{},
		containers: map[string]*containerLease{},
	}
	m.retainOrphans(&containerLease{token: "lost-tcp", ports: []leasedPort{{desired: tcpPublication, live: true, exact: true}}})

	udp, err := m.prepareCreate(context.Background(), portBindings{"80/udp": {{HostPort: "39210"}}})
	if err != nil {
		t.Fatalf("TCP orphan incorrectly reserved UDP tuple: %v", err)
	}
	if err := m.bindCreate(udp.Token, "udp-container", "udp"); err != nil {
		t.Fatal(err)
	}
	if err := m.reconcileOrphans(context.Background()); err == nil {
		t.Fatal("failed orphan unpublish was treated as success")
	}

	m.mu.Lock()
	forbidden := m.forbiddenPortsLocked()
	orphanAmbiguous := len(m.orphans) == 1 && !m.orphans[0].port.exact
	m.mu.Unlock()
	if !forbidden[tcpPublication.reservation()] {
		t.Fatal("failed TCP orphan cleanup released its reservation")
	}
	if !orphanAmbiguous {
		t.Fatal("failed TCP orphan cleanup was not retained as ambiguous")
	}
	udpReservation := portReservation{protocol: "udp", port: 39210}
	if !forbidden[udpReservation] {
		t.Fatal("failed TCP orphan cleanup erased the UDP container lease")
	}
	f.mu.Lock()
	_, udpPublished := f.ports["127.0.0.1:39210:39210/udp4"]
	f.mu.Unlock()
	if !udpPublished {
		t.Fatal("failed TCP orphan cleanup unpublished the UDP tuple")
	}
}

func TestConcurrentOrphanReconciliationCannotUnpublishReusedReservation(t *testing.T) {
	f := newFakeSandbox(t)
	publication := portPublication{HostIP: "127.0.0.1", HostPort: 39212, SandboxPort: 39212, Protocol: "tcp4"}
	f.mu.Lock()
	f.ports[publication.spec()] = publication
	f.mu.Unlock()

	firstListed := make(chan struct{})
	releaseFirstList := make(chan struct{})
	concurrentList := make(chan struct{})
	staleUnpublish := make(chan struct{})
	releaseStaleUnpublish := make(chan struct{})
	var releaseFirstOnce sync.Once
	var releaseStaleOnce sync.Once
	var controlMu sync.Mutex
	firstListSeen := false
	firstListHeld := false
	concurrentListSeen := false
	unpublishCount := 0
	runtime := f.runtime(t)
	runtime.run = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) > 2 && args[2] == "--format" {
			data, err := f.run(ctx, args...)
			controlMu.Lock()
			if !firstListSeen {
				firstListSeen, firstListHeld = true, true
				close(firstListed)
				controlMu.Unlock()
				select {
				case <-releaseFirstList:
				case <-ctx.Done():
					return nil, ctx.Err()
				}
				controlMu.Lock()
				firstListHeld = false
				controlMu.Unlock()
				return data, err
			}
			if firstListHeld && !concurrentListSeen {
				concurrentListSeen = true
				close(concurrentList)
			}
			controlMu.Unlock()
			return data, err
		}
		if len(args) > 2 && args[2] == "--unpublish" {
			controlMu.Lock()
			unpublishCount++
			stale := unpublishCount == 2
			if stale {
				close(staleUnpublish)
			}
			controlMu.Unlock()
			if stale {
				select {
				case <-releaseStaleUnpublish:
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
		}
		return f.run(ctx, args...)
	}
	m := &portManager{
		runtime:    runtime,
		engine:     newDockerEngine(runtime.dockerURL, &http.Client{}),
		creations:  map[string]*containerLease{},
		containers: map[string]*containerLease{},
	}
	m.retainOrphans(&containerLease{token: "old", ports: []leasedPort{{desired: publication, live: true, exact: true}}})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	done := make(chan error, 2)
	go func() { done <- m.reconcileOrphans(ctx) }()
	select {
	case <-firstListed:
	case <-ctx.Done():
		t.Fatal("first orphan reconciliation did not reach its list barrier")
	}
	go func() { done <- m.reconcileOrphans(ctx) }()

	concurrentSnapshot := false
	select {
	case <-concurrentList:
		concurrentSnapshot = true
	case <-time.After(200 * time.Millisecond):
		// A serialized second pass cannot snapshot the orphan while the first
		// pass is still holding the allocation boundary.
	}

	var errs []error
	if concurrentSnapshot {
		select {
		case err := <-done:
			errs = append(errs, err)
		case <-ctx.Done():
			errs = append(errs, errors.New("concurrent orphan pass did not finish"))
		}
	}
	prepared, prepareErr := preparedCreate{}, error(nil)
	if concurrentSnapshot {
		prepared, prepareErr = m.prepareCreate(ctx, portBindings{"80/tcp": {{HostPort: "39212"}}})
	}
	releaseFirstOnce.Do(func() { close(releaseFirstList) })

	if concurrentSnapshot {
		select {
		case <-staleUnpublish:
		case <-ctx.Done():
			errs = append(errs, errors.New("stale orphan pass did not reach unpublish"))
		}
		releaseStaleOnce.Do(func() { close(releaseStaleUnpublish) })
		select {
		case err := <-done:
			errs = append(errs, err)
		case <-ctx.Done():
			errs = append(errs, errors.New("stale orphan pass did not finish"))
		}
	} else {
		for range 2 {
			select {
			case err := <-done:
				errs = append(errs, err)
			case <-ctx.Done():
				errs = append(errs, errors.New("serialized orphan pass did not finish"))
			}
		}
		prepared, prepareErr = m.prepareCreate(ctx, portBindings{"80/tcp": {{HostPort: "39212"}}})
	}
	releaseStaleOnce.Do(func() { close(releaseStaleUnpublish) })

	if err := errors.Join(errs...); err != nil {
		t.Fatal(err)
	}
	if prepareErr != nil {
		t.Fatalf("reuse reservation after orphan cleanup: %v", prepareErr)
	}
	if prepared.Bindings["80/tcp"][0].HostPort != "39212" {
		t.Fatalf("unexpected reused binding: %+v", prepared.Bindings)
	}
	f.mu.Lock()
	_, published := f.ports[publication.spec()]
	f.mu.Unlock()
	if !published {
		t.Fatal("stale orphan pass unpublished the newly reused tuple")
	}
}

func TestAbortCleanupExcludesReuseThroughOrphanTransfer(t *testing.T) {
	for _, releaseFails := range []bool{false, true} {
		t.Run(fmt.Sprintf("release-fails-%t", releaseFails), func(t *testing.T) {
			f := newFakeSandbox(t)
			unpublishEntered := make(chan struct{})
			releaseUnpublish := make(chan struct{})
			var releaseOnce sync.Once
			release := func() { releaseOnce.Do(func() { close(releaseUnpublish) }) }
			defer release()
			var interceptMu sync.Mutex
			intercept := true
			runtime := f.runtime(t)
			runtime.run = func(ctx context.Context, args ...string) ([]byte, error) {
				if len(args) > 2 && args[2] == "--unpublish" {
					interceptMu.Lock()
					hold := intercept
					intercept = false
					interceptMu.Unlock()
					if hold {
						data, err := f.run(ctx, args...)
						close(unpublishEntered)
						select {
						case <-releaseUnpublish:
						case <-ctx.Done():
							return nil, ctx.Err()
						}
						if releaseFails {
							return nil, errors.New("unpublish response lost")
						}
						return data, err
					}
				}
				return f.run(ctx, args...)
			}
			m := &portManager{
				runtime:    runtime,
				engine:     newDockerEngine(runtime.dockerURL, &http.Client{}),
				creations:  map[string]*containerLease{},
				containers: map[string]*containerLease{},
			}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			prepared, err := m.prepareCreate(ctx, portBindings{"80/tcp": {{HostPort: "39213"}}})
			if err != nil {
				t.Fatal(err)
			}
			abortDone := make(chan error, 1)
			go func() { abortDone <- m.abortCreate(ctx, prepared.Token, false) }()
			select {
			case <-unpublishEntered:
			case <-ctx.Done():
				release()
				select {
				case <-abortDone:
				case <-time.After(time.Second):
				}
				t.Fatal("abort did not reach controlled unpublish")
			}

			prepareDone := make(chan error, 1)
			go func() {
				_, prepareErr := m.prepareCreate(ctx, portBindings{"81/tcp": {{HostPort: "39213"}}})
				prepareDone <- prepareErr
			}()
			select {
			case err := <-prepareDone:
				release()
				abortErr := waitPortTestError(t, ctx, abortDone, "abort after premature prepare")
				t.Fatalf("fresh creation passed abort cleanup boundary: prepare=%v abort=%v", err, abortErr)
			case <-time.After(100 * time.Millisecond):
			}
			release()
			abortErr := waitPortTestError(t, ctx, abortDone, "abort cleanup")
			prepareErr := waitPortTestError(t, ctx, prepareDone, "concurrent prepare")
			if releaseFails {
				if abortErr == nil {
					t.Fatal("lost unpublish response was treated as confirmed release")
				}
				if prepareErr == nil {
					t.Fatal("fresh creation reused a tuple transferred to orphan ownership")
				}
				if err := m.reconcileOrphans(ctx); err != nil {
					t.Fatalf("reconcile absent orphan: %v", err)
				}
			} else {
				if abortErr != nil {
					t.Fatalf("abort cleanup: %v", abortErr)
				}
				if prepareErr != nil {
					t.Fatalf("reuse after confirmed abort cleanup: %v", prepareErr)
				}
				// The concurrent prepare already installed the later owner.
				if err := m.abortCreate(ctx, prepared.Token, false); err != nil {
					t.Fatalf("duplicate abort: %v", err)
				}
			}
			if releaseFails {
				if _, err := m.prepareCreate(ctx, portBindings{"81/tcp": {{HostPort: "39213"}}}); err != nil {
					t.Fatalf("reuse after orphan absence was proven: %v", err)
				}
			}
			f.mu.Lock()
			_, published := f.ports["127.0.0.1:39213:39213/tcp4"]
			f.mu.Unlock()
			if !published {
				t.Fatal("stale or duplicate abort removed the later owner's forwarding")
			}
		})
	}
}

func TestAbortLosesToBoundCreateWithoutStaleCleanup(t *testing.T) {
	f := newFakeSandbox(t)
	m := testPortManager(t, f)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	prepared, err := m.prepareCreate(ctx, portBindings{"80/tcp": {{HostPort: "39214"}}})
	if err != nil {
		t.Fatal(err)
	}
	m.mu.Lock()
	record := m.creations[prepared.Token]
	m.mu.Unlock()
	record.mu.Lock()
	m.allocationMu.Lock()
	recordLocked, allocationLocked := true, true
	defer func() {
		if recordLocked {
			record.mu.Unlock()
		}
		if allocationLocked {
			m.allocationMu.Unlock()
		}
	}()
	abortDone := make(chan error, 1)
	go func() { abortDone <- m.abortCreate(ctx, prepared.Token, false) }()

	// A pre-fix abort deletes creation ownership before waiting on record.mu.
	// The fixed abort cannot take its authoritative snapshot while allocation is held.
	oldDeletion := false
	deadline := time.Now().Add(100 * time.Millisecond)
	for time.Now().Before(deadline) {
		m.mu.Lock()
		oldDeletion = m.creations[prepared.Token] == nil
		m.mu.Unlock()
		if oldDeletion {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if oldDeletion {
		record.mu.Unlock()
		recordLocked = false
		m.allocationMu.Unlock()
		allocationLocked = false
		_ = waitPortTestError(t, ctx, abortDone, "pre-bound abort")
		t.Fatal("abort claimed creation before entering allocation boundary")
	}
	if err := m.bindCreate(prepared.Token, "bound-container", "bound"); err != nil {
		record.mu.Unlock()
		recordLocked = false
		m.allocationMu.Unlock()
		allocationLocked = false
		_ = waitPortTestError(t, ctx, abortDone, "abort after bind failure")
		t.Fatal(err)
	}
	record.mu.Unlock()
	recordLocked = false
	m.allocationMu.Unlock()
	allocationLocked = false
	if err := waitPortTestError(t, ctx, abortDone, "losing abort"); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	_, published := f.ports["127.0.0.1:39214:39214/tcp4"]
	f.mu.Unlock()
	if !published {
		t.Fatal("losing abort unpublished the bound container's forwarding")
	}
}

func TestRemovedRecordReconciliationCannotUnpublishReusedTuple(t *testing.T) {
	f := newFakeSandbox(t)
	runtime := f.runtime(t)
	publication := portPublication{HostIP: "127.0.0.1", HostPort: 39215, SandboxPort: 39215, Protocol: "tcp4"}
	f.mu.Lock()
	f.ports[publication.spec()] = publication
	f.mu.Unlock()
	record := &containerLease{
		token: "removed-owner", id: "removed-container",
		ports: []leasedPort{{containerPort: "80/tcp", desired: publication, live: true, exact: true}},
	}
	m := &portManager{
		runtime: runtime, engine: newDockerEngine(runtime.dockerURL, &http.Client{}),
		creations: map[string]*containerLease{}, containers: map[string]*containerLease{"removed-container": record},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	record.mu.Lock()
	recordLocked := true
	defer func() {
		if recordLocked {
			record.mu.Unlock()
		}
	}()
	reconcileStarted := make(chan struct{})
	reconcileDone := make(chan error, 1)
	go func() {
		close(reconcileStarted)
		reconcileDone <- m.reconcileRecord(ctx, record)
	}()
	select {
	case <-reconcileStarted:
	case <-ctx.Done():
		record.mu.Unlock()
		recordLocked = false
		select {
		case <-reconcileDone:
		case <-time.After(time.Second):
		}
		t.Fatal("stale reconciliation did not start")
	}
	m.removeRecord(record)
	if err := m.reconcileOrphans(ctx); err != nil {
		record.mu.Unlock()
		recordLocked = false
		_ = waitPortTestError(t, ctx, reconcileDone, "stale reconciliation after orphan failure")
		t.Fatalf("reclaim removed owner's orphan: %v", err)
	}
	prepared, err := m.prepareCreate(ctx, portBindings{"81/tcp": {{HostPort: "39215"}}})
	if err != nil {
		record.mu.Unlock()
		recordLocked = false
		_ = waitPortTestError(t, ctx, reconcileDone, "stale reconciliation after prepare failure")
		t.Fatalf("reuse reclaimed tuple: %v", err)
	}
	if err := m.bindCreate(prepared.Token, "new-container", "new"); err != nil {
		record.mu.Unlock()
		recordLocked = false
		_ = waitPortTestError(t, ctx, reconcileDone, "stale reconciliation after bind failure")
		t.Fatal(err)
	}
	record.mu.Unlock()
	recordLocked = false
	if err := waitPortTestError(t, ctx, reconcileDone, "stale reconciliation"); err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	_, published := f.ports[publication.spec()]
	f.mu.Unlock()
	if !published {
		t.Fatal("stale removed-record reconciliation unpublished the new owner's tuple")
	}
}

func TestEphemeralRetryPreservesCreateAndEngineReservations(t *testing.T) {
	f := newFakeSandbox(t)
	f.list = []dockerContainer{{Ports: []dockerPort{{PrivatePort: 80, PublicPort: 40000, Type: "tcp"}}}}
	runtime := f.runtime(t)
	publishCalls := 0
	runtime.run = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) > 2 && args[2] == "--publish" {
			publishCalls++
			if publishCalls == 2 || publishCalls == 5 || publishCalls == 6 {
				return nil, errors.New("address already in use")
			}
		}
		return f.run(ctx, args...)
	}
	m := &portManager{
		runtime: runtime, engine: newDockerEngine(runtime.dockerURL, &http.Client{}),
		creations: map[string]*containerLease{}, containers: map[string]*containerLease{},
	}
	// Planning consumes the first two candidates before either publication retries.
	candidates := []int{39990, 39991, 40000, 40001, 40003, 40004, 40004, 40005, 40005, 40002}
	m.portCandidate = func(protocol string) (int, error) {
		if protocol != "tcp" {
			return 0, fmt.Errorf("unexpected candidate protocol %q", protocol)
		}
		if len(candidates) == 0 {
			return 0, errors.New("candidate sequence exhausted")
		}
		port := candidates[0]
		candidates = candidates[1:]
		return port, nil
	}
	prepared, err := m.prepareCreate(context.Background(), portBindings{
		"53/udp": {{HostPort: "40002"}},
		"79/tcp": {{HostPort: "0"}},
		"80/tcp": {{HostPort: "40001"}},
		"81/tcp": {{HostPort: "0"}},
		"82/tcp": {{HostPort: "40003"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := prepared.Bindings["79/tcp"][0].HostPort; got != "40004" {
		t.Fatalf("first retry did not use expected replacement tuple: got %s, want 40004", got)
	}
	if got := prepared.Bindings["81/tcp"][0].HostPort; got != "40002" {
		t.Fatalf("retry did not preserve protocol-aware exclusions: got %s, want 40002", got)
	}
	if len(candidates) != 0 {
		t.Fatalf("retry accepted an Engine or same-create TCP reservation; unused candidates: %v", candidates)
	}
	if publishCalls != 8 {
		t.Fatalf("unexpected publish attempts: got %d, want 8", publishCalls)
	}
}

func TestForbiddenPortEnumerationIgnoresMutableLeaseFlags(t *testing.T) {
	desired := portPublication{HostIP: "127.0.0.1", HostPort: 39211, SandboxPort: 39211, Protocol: "tcp4"}
	record := &containerLease{ports: []leasedPort{{desired: desired, live: true, exact: true}}}
	m := &portManager{creations: map[string]*containerLease{"token": record}, containers: map[string]*containerLease{}}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			record.mu.Lock()
			record.ports[0].live = i%2 == 0
			record.ports[0].orphan = i%3 == 0
			record.ports[0].exact = i%5 == 0
			record.mu.Unlock()
		}
	}()
	for i := 0; i < 10000; i++ {
		m.mu.Lock()
		forbidden := m.forbiddenPortsLocked()
		m.mu.Unlock()
		if !forbidden[desired.reservation()] {
			t.Fatal("immutable desired reservation disappeared")
		}
	}
	wg.Wait()
}

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func waitPortTestError(t *testing.T, ctx context.Context, result <-chan error, operation string) error {
	t.Helper()
	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		t.Fatalf("%s did not finish: %v", operation, ctx.Err())
		return ctx.Err()
	}
}
