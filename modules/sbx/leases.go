package sbx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	portOperationTimeout = 15 * time.Second
	portReconcilePeriod  = 10 * time.Second
)

var errUnmanagedContainer = errors.New("container has no managed port lease")

type leasedPort struct {
	containerPort string
	desired       portPublication
	live          bool
	orphan        bool
	exact         bool
}

type containerLease struct {
	mu                  sync.Mutex
	token               string
	id                  string
	name                string
	ports               []leasedPort
	firstStartAttempted bool
	ambiguousCreate     bool
	removed             bool
}

type orphanLease struct {
	id    uint64
	token string
	port  leasedPort
}

type portManager struct {
	runtime sandboxRuntime
	client  *http.Client
	engine  *dockerEngine

	allocationMu  sync.Mutex
	closeMu       sync.Mutex
	mu            sync.Mutex
	creations     map[string]*containerLease
	containers    map[string]*containerLease
	orphans       []orphanLease
	nextOrphanID  uint64
	closed        bool
	portCandidate func(string) (int, error)

	watchCtx    context.Context
	watchCancel context.CancelFunc
	watchWG     sync.WaitGroup
}

func newPortManager(ctx context.Context, runtime sandboxRuntime, client *http.Client) (*portManager, error) {
	if runtime.name == "" || runtime.run == nil || runtime.dockerURL == nil {
		return nil, errors.New("initialize sbx port manager: incomplete sandbox runtime")
	}
	if client == nil {
		client = &http.Client{}
	}
	watchCtx, cancel := context.WithCancel(context.Background())
	m := &portManager{
		runtime: runtime, client: client, engine: newDockerEngine(runtime.dockerURL, client),
		creations: make(map[string]*containerLease), containers: make(map[string]*containerLease),
		watchCtx: watchCtx, watchCancel: cancel,
	}
	initialCtx, initialCancel := boundedContext(ctx, portOperationTimeout)
	_, err := m.engine.knownPorts(initialCtx)
	if err == nil {
		err = m.reconcile(initialCtx)
	}
	initialCancel()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("initial Docker port reconciliation for sandbox %q: %w", runtime.name, err)
	}
	m.startWatchers()
	return m, nil
}

func (m *portManager) prepareCreate(ctx context.Context, bindings portBindings) (preparedCreate, error) {
	m.allocationMu.Lock()
	defer m.allocationMu.Unlock()

	token, err := newLeaseToken()
	if err != nil {
		return preparedCreate{}, err
	}
	known, err := m.engine.knownPorts(ctx)
	if err != nil {
		return preparedCreate{}, fmt.Errorf("sandbox %q creation %s: inspect Docker ports before reservation: %w", m.runtime.name, token, err)
	}
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return preparedCreate{}, errors.New("sbx port manager is closed")
	}
	for reservation := range m.forbiddenPortsLocked() {
		known[reservation] = true
	}
	m.mu.Unlock()

	candidate := m.portCandidate
	if candidate == nil {
		candidate = availableHostPort
	}
	plans, err := planBindingsUsing(ctx, bindings, known, candidate)
	if err != nil {
		return preparedCreate{}, fmt.Errorf("sandbox %q creation %s: %w", m.runtime.name, token, err)
	}
	for _, plan := range plans {
		known[plan.publication.reservation()] = true
	}
	record := &containerLease{token: token, ports: make([]leasedPort, 0, len(plans))}
	for _, plan := range plans {
		publication, exact, publishErr := m.publish(ctx, token, plan, known)
		if publishErr != nil {
			if publication.HostPort != 0 {
				record.ports = append(record.ports, leasedPort{containerPort: plan.containerPort, desired: publication, live: true, orphan: true, exact: exact})
			}
			cleanupCtx, cleanupCancel := freshPortOperationContext()
			cleanupErr := m.releaseRecord(cleanupCtx, record)
			cleanupCancel()
			m.retainOrphans(record)
			return preparedCreate{}, errors.Join(publishErr, cleanupErr)
		}
		record.ports = append(record.ports, leasedPort{containerPort: plan.containerPort, desired: publication, live: true, exact: true})
		known[publication.reservation()] = true
	}

	rewritten := cloneBindings(bindings)
	indices := make(map[string]int)
	for _, lease := range record.ports {
		index := indices[lease.containerPort]
		rewritten[lease.containerPort][index] = portBinding{HostIP: "0.0.0.0", HostPort: strconv.Itoa(lease.desired.HostPort)}
		indices[lease.containerPort] = index + 1
	}
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		cleanupCtx, cleanupCancel := freshPortOperationContext()
		cleanupErr := m.releaseRecord(cleanupCtx, record)
		cleanupCancel()
		m.retainOrphans(record)
		return preparedCreate{}, errors.Join(errors.New("sbx port manager is closed"), cleanupErr)
	}
	m.creations[token] = record
	m.mu.Unlock()
	return preparedCreate{Token: token, Bindings: rewritten}, nil
}

func cloneBindings(bindings portBindings) portBindings {
	clone := make(portBindings, len(bindings))
	for key, entries := range bindings {
		clone[key] = append([]portBinding(nil), entries...)
	}
	return clone
}

func (m *portManager) bindCreate(token, containerID, containerName string) error {
	if token == "" || containerID == "" {
		return errors.New("bind sbx port lease: token and container ID are required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return errors.New("sbx port manager is closed")
	}
	record := m.creations[token]
	if record == nil {
		if existing := m.containers[containerID]; existing != nil && existing.token == token {
			return nil
		}
		return fmt.Errorf("sandbox %q container %q: unknown port lease token %q", m.runtime.name, containerID, token)
	}
	if existing := m.containers[containerID]; existing != nil && existing != record {
		return fmt.Errorf("sandbox %q container %q already has a port lease", m.runtime.name, containerID)
	}
	delete(m.creations, token)
	record.id, record.name = containerID, strings.TrimPrefix(containerName, "/")
	m.containers[containerID] = record
	return nil
}

func (m *portManager) abortCreate(ctx context.Context, token string, ambiguous bool) error {
	if ambiguous {
		m.mu.Lock()
		record := m.creations[token]
		if record == nil {
			m.mu.Unlock()
			return nil
		}
		record.ambiguousCreate = true
		m.mu.Unlock()
		containers, err := m.engine.listByLease(ctx, token)
		if err != nil {
			return fmt.Errorf("sandbox %q creation %s: resolve ambiguous Docker create: %w", m.runtime.name, token, err)
		}
		switch len(containers) {
		case 0:
			// Docker may finish the create after the lost response and this list.
			// Keep the reservation until an event/reconciliation finds its owner or
			// whole-sandbox cleanup makes the ambiguity irrelevant.
			return nil
		case 1:
			return m.bindCreate(token, containers[0].ID, containers[0].Name)
		default:
			return fmt.Errorf("sandbox %q creation %s: ambiguous Docker create matched %d containers; reservations retained", m.runtime.name, token, len(containers))
		}
	}
	m.allocationMu.Lock()
	defer m.allocationMu.Unlock()
	m.mu.Lock()
	record := m.creations[token]
	if record == nil {
		m.mu.Unlock()
		return nil
	}
	// Claim cleanup while creation ownership is authoritative. bindCreate and
	// observed-container binding either transfer the record first or lose to
	// this deletion; no stale abort may clean a subsequently bound owner.
	delete(m.creations, token)
	m.mu.Unlock()
	record.mu.Lock()
	err := m.releaseRecordLocked(ctx, record)
	record.mu.Unlock()
	m.mu.Lock()
	m.retainOrphansLocked(record)
	m.mu.Unlock()
	return err
}

func (m *portManager) beginStart(ctx context.Context, containerRef string) (func(context.Context, bool) error, error) {
	m.allocationMu.Lock()
	defer m.allocationMu.Unlock()
	m.mu.Lock()
	closed := m.closed
	m.mu.Unlock()
	if closed {
		return nil, errors.New("sbx port manager is closed")
	}

	record, err := m.lockContainer(containerRef)
	if errors.Is(err, errUnmanagedContainer) {
		return func(context.Context, bool) error { return nil }, nil
	}
	if err != nil {
		return nil, err
	}
	record.firstStartAttempted = true
	newlyPublished := make([]int, 0)
	var published []portPublication
	listedPublished := false
	for i := range record.ports {
		lease := &record.ports[i]
		if lease.live {
			if lease.exact {
				continue
			}
			if !listedPublished {
				published, err = m.listPublished(ctx)
				listedPublished = true
			}
			if err != nil {
				cleanupErr := m.releaseIndices(ctx, record, newlyPublished)
				record.mu.Unlock()
				return nil, errors.Join(fmt.Errorf("sandbox %q container %q port %s: verify uncertain publication before start: %w", m.runtime.name, containerRef, lease.desired.spec(), err), cleanupErr)
			}
			if containsPublication(published, lease.desired) {
				lease.exact, lease.orphan = true, false
				continue
			}
			// The authoritative list proves the earlier ambiguous publish did
			// not leave this tuple behind. Retry the immutable desired port.
			lease.live, lease.orphan = false, false
		}
		publication, exact, publishErr := m.publish(ctx, record.token, portPlan{containerPort: lease.containerPort, publication: lease.desired}, nil)
		if publishErr != nil || publication != lease.desired {
			var unexpectedErr error
			if publication.HostPort != 0 && publication == lease.desired {
				lease.live, lease.orphan, lease.exact = true, true, exact
			} else if publication.HostPort != 0 {
				unexpected := &containerLease{token: record.token, ports: []leasedPort{{desired: publication, live: true, exact: exact}}}
				cleanupCtx, cancel := freshPortOperationContext()
				unexpectedErr = m.releaseRecord(cleanupCtx, unexpected)
				cancel()
				m.retainOrphans(unexpected)
			}
			cleanupErr := m.releaseIndices(ctx, record, newlyPublished)
			record.mu.Unlock()
			if publishErr == nil {
				publishErr = fmt.Errorf("sbx changed immutable desired port from %s to %s", lease.desired.spec(), publication.spec())
			}
			return nil, errors.Join(fmt.Errorf("sandbox %q container %q port %s: restore before start: %w", m.runtime.name, containerRef, lease.desired.spec(), publishErr), unexpectedErr, cleanupErr)
		}
		lease.live, lease.orphan, lease.exact = true, false, true
		newlyPublished = append(newlyPublished, i)
	}
	return m.startFinish(record), nil
}

func (m *portManager) startFinish(record *containerLease) func(context.Context, bool) error {
	var once sync.Once
	return func(ctx context.Context, started bool) (retErr error) {
		once.Do(func() {
			defer record.mu.Unlock()
			if started {
				return
			}
			retErr = m.releaseIfConfirmedStoppedLocked(ctx, record)
		})
		return retErr
	}
}

func (m *portManager) beginStop(containerRef string) func(context.Context, bool) error {
	record, err := m.lockContainer(containerRef)
	if errors.Is(err, errUnmanagedContainer) {
		return func(context.Context, bool) error { return nil }
	}
	if err != nil {
		return func(context.Context, bool) error { return err }
	}
	var once sync.Once
	return func(ctx context.Context, stopped bool) (retErr error) {
		once.Do(func() {
			defer record.mu.Unlock()
			if stopped {
				retErr = m.releaseRecordLocked(ctx, record)
				return
			}
			retErr = m.releaseIfConfirmedStoppedLocked(ctx, record)
		})
		return retErr
	}
}

func (m *portManager) beginDestroy(containerRef string) func(context.Context, bool) error {
	record, err := m.lockContainer(containerRef)
	if errors.Is(err, errUnmanagedContainer) {
		return func(context.Context, bool) error { return nil }
	}
	if err != nil {
		return func(context.Context, bool) error { return err }
	}
	var once sync.Once
	return func(ctx context.Context, destroyed bool) (retErr error) {
		once.Do(func() {
			defer record.mu.Unlock()
			if !destroyed {
				inspect, inspectErr := m.engine.inspect(ctx, record.id)
				if inspectErr != nil {
					retErr = inspectErr
					return
				}
				destroyed = inspect == nil
			}
			if destroyed {
				retErr = m.releaseRecordLocked(ctx, record)
				m.removeRecord(record)
			}
		})
		return retErr
	}
}

func (m *portManager) releaseIfConfirmedStoppedLocked(ctx context.Context, record *containerLease) error {
	inspect, err := m.engine.inspect(ctx, record.id)
	if err != nil {
		return fmt.Errorf("sandbox %q container %q: inspect after lifecycle failure: %w", m.runtime.name, record.id, err)
	}
	if inspect == nil {
		err = m.releaseRecordLocked(ctx, record)
		m.removeRecord(record)
		return err
	}
	if !inspect.Running {
		return m.releaseRecordLocked(ctx, record)
	}
	return nil
}

func (m *portManager) lockContainer(ref string) (*containerLease, error) {
	m.mu.Lock()
	record := m.resolveContainerLocked(ref)
	m.mu.Unlock()
	if record == nil {
		ctx, cancel := context.WithTimeout(context.Background(), portOperationTimeout)
		inspect, err := m.engine.inspect(ctx, ref)
		cancel()
		if err != nil {
			return nil, fmt.Errorf("sandbox %q container %q: resolve canonical ID: %w", m.runtime.name, ref, err)
		}
		if inspect != nil {
			m.mu.Lock()
			record = m.containers[inspect.ID]
			if record == nil {
				token := inspect.Labels[portLeaseLabel]
				record = m.creations[token]
				if record != nil {
					delete(m.creations, token)
					record.id, record.name = inspect.ID, strings.TrimPrefix(inspect.Name, "/")
					m.containers[inspect.ID] = record
				}
			}
			m.mu.Unlock()
		}
		if record == nil {
			return nil, fmt.Errorf("sandbox %q container %q: %w", m.runtime.name, ref, errUnmanagedContainer)
		}
	}
	record.mu.Lock()
	if record.removed {
		record.mu.Unlock()
		return nil, fmt.Errorf("sandbox %q container %q has already been removed: %w", m.runtime.name, ref, errUnmanagedContainer)
	}
	return record, nil
}

func (m *portManager) resolveContainerLocked(ref string) *containerLease {
	if record := m.containers[ref]; record != nil {
		return record
	}
	clean := strings.TrimPrefix(ref, "/")
	var matched *containerLease
	for id, record := range m.containers {
		idMatch := len(clean) >= 12 && strings.HasPrefix(id, clean)
		if record.name != clean && !idMatch {
			continue
		}
		if matched != nil && matched != record {
			return nil
		}
		matched = record
	}
	return matched
}

func (m *portManager) releaseRecord(ctx context.Context, record *containerLease) error {
	record.mu.Lock()
	defer record.mu.Unlock()
	return m.releaseRecordLocked(ctx, record)
}

func (m *portManager) releaseRecordLocked(ctx context.Context, record *containerLease) error {
	var errs []error
	var published []portPublication
	var listErr error
	listedPublished := false
	for i := range record.ports {
		lease := &record.ports[i]
		if !lease.live {
			continue
		}
		if !lease.exact {
			if !listedPublished {
				published, listErr = m.listPublished(ctx)
				listedPublished = true
			}
			if listErr != nil {
				lease.orphan = true
				errs = append(errs, fmt.Errorf("sandbox %q container %q port %s: verify uncertain publication before release: %w", m.runtime.name, record.id, lease.desired.spec(), listErr))
				continue
			}
			if !containsPublication(published, lease.desired) {
				lease.live, lease.orphan = false, false
				continue
			}
			lease.exact = true
		}
		if err := m.unpublish(ctx, lease.desired); err != nil {
			// The unpublish command may have taken effect even when its command
			// or authoritative follow-up list failed. Force the next start to
			// prove presence or republish instead of trusting stale live state.
			lease.orphan, lease.exact = true, false
			errs = append(errs, err)
			continue
		}
		lease.live, lease.orphan = false, false
	}
	return errors.Join(errs...)
}

func (m *portManager) releaseIndices(ctx context.Context, record *containerLease, indices []int) error {
	var errs []error
	for _, index := range indices {
		lease := &record.ports[index]
		if err := m.unpublish(ctx, lease.desired); err != nil {
			lease.orphan, lease.exact = true, false
			errs = append(errs, err)
			continue
		}
		lease.live = false
	}
	return errors.Join(errs...)
}

func (m *portManager) retainOrphans(record *containerLease) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.retainOrphansLocked(record)
}

func (m *portManager) retainOrphansLocked(record *containerLease) {
	for _, lease := range record.ports {
		if lease.live {
			lease.orphan = true
			m.nextOrphanID++
			m.orphans = append(m.orphans, orphanLease{id: m.nextOrphanID, token: record.token, port: lease})
		}
	}
}

func (m *portManager) removeRecord(record *containerLease) {
	record.removed = true
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.containers, record.id)
	delete(m.creations, record.token)
	m.retainOrphansLocked(record)
}

func (m *portManager) forbiddenPortsLocked() map[portReservation]bool {
	ports := map[portReservation]bool{{protocol: "tcp", port: 2375}: true}
	for _, record := range m.creations {
		for i := range record.ports {
			ports[record.ports[i].desired.reservation()] = true
		}
	}
	for _, record := range m.containers {
		for i := range record.ports {
			ports[record.ports[i].desired.reservation()] = true
		}
	}
	for _, orphan := range m.orphans {
		ports[orphan.port.desired.reservation()] = true
	}
	return ports
}

func (m *portManager) reconcile(ctx context.Context) error {
	m.mu.Lock()
	records := make([]*containerLease, 0, len(m.containers))
	for _, record := range m.containers {
		records = append(records, record)
	}
	creations := make([]*containerLease, 0, len(m.creations))
	for _, record := range m.creations {
		if record.ambiguousCreate {
			creations = append(creations, record)
		}
	}
	m.mu.Unlock()
	var errs []error
	for _, record := range creations {
		containers, err := m.engine.listByLease(ctx, record.token)
		if err != nil {
			errs = append(errs, fmt.Errorf("sandbox %q creation %s: reconcile ambiguous create: %w", m.runtime.name, record.token, err))
			continue
		}
		if len(containers) == 1 {
			if err := m.bindCreate(record.token, containers[0].ID, containers[0].Name); err != nil {
				errs = append(errs, err)
			}
		} else if len(containers) > 1 {
			errs = append(errs, fmt.Errorf("sandbox %q creation %s: ambiguous create matched %d containers", m.runtime.name, record.token, len(containers)))
		}
	}
	for _, record := range records {
		if err := m.reconcileRecord(ctx, record); err != nil {
			errs = append(errs, err)
		}
	}
	if err := m.reconcileOrphans(ctx); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (m *portManager) reconcileRecord(ctx context.Context, record *containerLease) error {
	record.mu.Lock()
	defer record.mu.Unlock()
	if record.removed {
		return nil
	}
	inspect, err := m.engine.inspect(ctx, record.id)
	if err != nil {
		return fmt.Errorf("sandbox %q container %q: reconcile inspect: %w", m.runtime.name, record.id, err)
	}
	if inspect == nil {
		err = m.releaseRecordLocked(ctx, record)
		m.removeRecord(record)
		return err
	}
	m.mu.Lock()
	record.name = strings.TrimPrefix(inspect.Name, "/")
	m.mu.Unlock()
	if !inspect.Running && (record.firstStartAttempted || recordHasOrphanedPublication(record)) {
		return m.releaseRecordLocked(ctx, record)
	}
	return nil
}

func recordHasOrphanedPublication(record *containerLease) bool {
	for _, lease := range record.ports {
		if lease.live && lease.orphan {
			return true
		}
	}
	return false
}

func (m *portManager) reconcileOrphans(ctx context.Context) error {
	m.allocationMu.Lock()
	defer m.allocationMu.Unlock()

	m.mu.Lock()
	if len(m.orphans) == 0 {
		m.mu.Unlock()
		return nil
	}
	m.ensureOrphanIDsLocked()
	captured := append([]orphanLease(nil), m.orphans...)
	m.mu.Unlock()
	published, err := m.listPublished(ctx)
	if err != nil {
		return err
	}
	type orphanResult struct {
		keep bool
	}
	results := make(map[uint64]orphanResult, len(captured))
	var errs []error
	for _, orphan := range captured {
		if !containsPublication(published, orphan.port.desired) {
			results[orphan.id] = orphanResult{}
			continue
		}
		unpublishErr := m.unpublish(ctx, orphan.port.desired)
		results[orphan.id] = orphanResult{keep: unpublishErr != nil}
		if unpublishErr != nil {
			errs = append(errs, fmt.Errorf("sandbox %q creation %s orphan port %s: %w", m.runtime.name, orphan.token, orphan.port.desired.spec(), unpublishErr))
		}
	}
	m.mu.Lock()
	kept := make([]orphanLease, 0, len(m.orphans))
	for _, orphan := range m.orphans {
		result, wasCaptured := results[orphan.id]
		if !wasCaptured {
			// An ambiguous publication appended after our list snapshot must
			// remain owned and forbidden until a later reconciliation sees it.
			kept = append(kept, orphan)
			continue
		}
		if result.keep {
			// The pre-unpublish list proved this tuple existed, but the failed
			// attempt leaves its current presence ambiguous again.
			orphan.port.exact = false
			kept = append(kept, orphan)
		}
	}
	m.orphans = kept
	m.mu.Unlock()
	return errors.Join(errs...)
}

func (m *portManager) ensureOrphanIDsLocked() {
	for _, orphan := range m.orphans {
		if orphan.id > m.nextOrphanID {
			m.nextOrphanID = orphan.id
		}
	}
	for i := range m.orphans {
		if m.orphans[i].id != 0 {
			continue
		}
		m.nextOrphanID++
		m.orphans[i].id = m.nextOrphanID
	}
}

func containsPublication(ports []portPublication, want portPublication) bool {
	for _, port := range ports {
		if port == want {
			return true
		}
	}
	return false
}

func (m *portManager) listPublished(ctx context.Context) ([]portPublication, error) {
	commandCtx, cancel := boundedContext(ctx, portCommandTimeout)
	defer cancel()
	output, err := m.runtime.run(commandCtx, "ports", m.runtime.name, "--format", "json")
	if err != nil {
		return nil, fmt.Errorf("sandbox %q: list published ports: %w", m.runtime.name, err)
	}
	var ports []portPublication
	if err := json.Unmarshal(output, &ports); err != nil {
		return nil, fmt.Errorf("sandbox %q: decode published ports %q: %w", m.runtime.name, string(output), err)
	}
	return ports, nil
}

func (m *portManager) close(ctx context.Context) error {
	m.closeMu.Lock()
	defer m.closeMu.Unlock()
	m.mu.Lock()
	m.closed = true
	m.mu.Unlock()
	m.watchCancel()
	done := make(chan struct{})
	go func() { m.watchWG.Wait(); close(done) }()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}

	// Watchers may need allocationMu to finish orphan reconciliation.
	// Drain them before waiting for in-flight allocations and their rollbacks.
	m.allocationMu.Lock()
	defer m.allocationMu.Unlock()

	m.mu.Lock()
	records := make([]*containerLease, 0, len(m.creations)+len(m.containers))
	for _, record := range m.creations {
		records = append(records, record)
	}
	for _, record := range m.containers {
		records = append(records, record)
	}
	orphans := append([]orphanLease(nil), m.orphans...)
	m.mu.Unlock()
	var errs []error
	for _, record := range records {
		if err := m.releaseRecord(ctx, record); err != nil {
			errs = append(errs, err)
		}
	}
	for _, orphan := range orphans {
		if orphan.port.live && orphan.port.exact {
			if err := m.unpublish(ctx, orphan.port.desired); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

func freshPortOperationContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), portOperationTimeout)
}
