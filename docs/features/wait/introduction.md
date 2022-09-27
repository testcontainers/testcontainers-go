# Wait Strategies

!!! info "Wait strategies vs Startup strategies"

    **Wait strategy:** is the container in a state that is useful for testing. This is generally approximated as 'can we talk to this container over the network'. However, there are quite a few variations and nuances.
    
    **Startup strategy:** did a container reach the desired running state. *Almost always* this just means 'wait until the container is running' - for a daemon process in a container this is the goal. Sometimes we need to wait until the container reaches a running state and then exits - this is the 'one shot startup' strategy, only used for cases where we need to run a one off command in a container but not a daemon.

When defining a wait strategy, Testcontainers will create a cancel context with 60 seconds defined as timeout.

If the default 60s timeout is not sufficient, it can be updated with the `WithStartupTimeout(startupTimeout time.Duration)` function, present at each wait struct.

Below you can find a list of the available wait strategies that you can use:

- [Exec](./wait/exec.md)
- [HostPort](./wait/host_port.md)
- [HTTP](./wait/http.md)
