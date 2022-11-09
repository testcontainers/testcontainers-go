package testcontainers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestParallelContainers(t *testing.T) {
	tests := []struct {
		name      string
		reqs      ParallelContainerRequest
		resLen    int
		expErrors int
	}{
		{
			name: "running two containers (one error)",
			reqs: ParallelContainerRequest{
				{
					ContainerRequest: ContainerRequest{

						Image: "nginx",
						ExposedPorts: []string{
							"10080/tcp",
						},
					},
					Started: true,
				},
				{
					ContainerRequest: ContainerRequest{

						Image: "bad bad bad",
						ExposedPorts: []string{
							"10081/tcp",
						},
					},
					Started: true,
				},
			},
			resLen:    1,
			expErrors: 1,
		},
		{
			name: "running two containers (all errors)",
			reqs: ParallelContainerRequest{
				{
					ContainerRequest: ContainerRequest{

						Image: "bad bad bad",
						ExposedPorts: []string{
							"10081/tcp",
						},
					},
					Started: true,
				},
				{
					ContainerRequest: ContainerRequest{

						Image: "bad bad bad",
						ExposedPorts: []string{
							"10081/tcp",
						},
					},
					Started: true,
				},
			},
			resLen:    0,
			expErrors: 2,
		},
		{
			name: "running two containers (success)",
			reqs: ParallelContainerRequest{
				{
					ContainerRequest: ContainerRequest{

						Image: "nginx",
						ExposedPorts: []string{
							"10080/tcp",
						},
					},
					Started: true,
				},
				{
					ContainerRequest: ContainerRequest{

						Image: "nginx",
						ExposedPorts: []string{
							"10081/tcp",
						},
					},
					Started: true,
				},
			},
			resLen:    2,
			expErrors: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, err := ParallelContainers(context.Background(), tc.reqs, ParallelContainersOptions{})
			if err != nil {
				require.NotZero(t, tc.expErrors)
				e, _ := err.(ParallelContainersError)

				if len(e.Errors) != tc.expErrors {
					t.Fatalf("expected erorrs: %d, got: %d\n", tc.expErrors, len(e.Errors))
				}
			}

			for _, c := range res {
				c := c
				terminateContainerOnEnd(t, context.Background(), c)
			}

			if len(res) != tc.resLen {
				t.Fatalf("expected containers: %d, got: %d\n", tc.resLen, len(res))
			}
		})
	}
}

func TestParallelContainersWithReuse(t *testing.T) {
	const (
		postgresPort     = 5432
		postgresPassword = "test"
		postgresUser     = "test"
		postgresDb       = "test"
	)

	natPort := fmt.Sprintf("%d/tcp", postgresPort)

	req := GenericContainerRequest{
		ContainerRequest: ContainerRequest{
			Image:        "postgis/postgis",
			Name:         "test-postgres",
			ExposedPorts: []string{natPort},
			Env: map[string]string{
				"POSTGRES_PASSWORD": postgresPassword,
				"POSTGRES_USER":     postgresUser,
				"POSTGRES_DATABASE": postgresDb,
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections").
				WithPollInterval(100 * time.Millisecond).
				WithOccurrence(2),
		},
		Started: true,
		Reuse:   true,
	}

	parallelRequest := ParallelContainerRequest{
		req,
		req,
		req,
	}

	ctx := context.Background()

	res, err := ParallelContainers(ctx, parallelRequest, ParallelContainersOptions{})
	if err != nil {
		e, _ := err.(ParallelContainersError)
		t.Fatalf("expected errors: %d, got: %d\n", 0, len(e.Errors))
	}
	// Container is reused, only terminate first container
	terminateContainerOnEnd(t, ctx, res[0])
}
