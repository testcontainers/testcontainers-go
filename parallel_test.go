package testcontainers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
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
				defer c.Terminate(context.Background())
			}

			if len(res) != tc.resLen {
				t.Fatalf("expected containers: %d, got: %d\n", tc.resLen, len(res))
			}
		})
	}
}
