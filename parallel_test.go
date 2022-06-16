package testcontainers

import (
	"context"
	"testing"
)

func TestParallelContainers(t *testing.T) {
	tests := []struct {
		name      string
		reqs      []GenericContainerRequest
		resLen    int
		expErrors int
	}{
		{
			name: "running two containers (success)",
			reqs: []GenericContainerRequest{
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
			resLen: 2,
		},
		{
			name: "running two containers (one error)",
			reqs: []GenericContainerRequest{
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, err := ParallelContainers(context.Background(), tc.reqs, ParallelContainersOptions{})

			if err != nil && tc.expErrors > 0 {
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
