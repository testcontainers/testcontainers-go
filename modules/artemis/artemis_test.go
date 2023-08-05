package artemis_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/go-stomp/stomp/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/artemis"
)

func TestArtemis(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		opts       []testcontainers.ContainerCustomizer
		user, pass string
		hook       func(*testing.T, *artemis.Container)
	}{
		{
			name: "Default",
			user: "artemis",
			pass: "artemis",
		},
		{
			name: "WithCredentials",
			opts: []testcontainers.ContainerCustomizer{
				artemis.WithCredentials("test", "test"),
			},
			user: "test",
			pass: "test",
		},
		{
			name: "WithAnonymous",
			opts: []testcontainers.ContainerCustomizer{
				artemis.WithAnonymousLogin(),
			},
		},
		{
			name: "WithExtraArgs",
			opts: []testcontainers.ContainerCustomizer{
				artemis.WithExtraArgs("--http-host 0.0.0.0 --relax-jolokia --queues ArgsTestQueue"),
			},
			user: "artemis",
			pass: "artemis",
			hook: func(t *testing.T, container *artemis.Container) {
				expectQueue(t, container, "ArgsTestQueue")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			container, err := artemis.RunContainer(ctx, test.opts...)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := container.Terminate(ctx); err != nil {
					t.Fatalf("failed to terminate container: %s", err)
				}
			})

			u, err := container.ConsoleURL(ctx)
			if err != nil {
				t.Fatal(err)
			}

			res, err := http.Get(u)
			if err != nil {
				t.Fatal(err)
			}
			res.Body.Close()

			if res.StatusCode != http.StatusOK {
				t.Error("failed to access console")
			}

			if test.user != "" && container.User() != test.user {
				t.Fatal("unexpected user")
			}

			if test.pass != "" && container.Password() != test.pass {
				t.Fatal("unexpected password")
			}

			host, err := container.BrokerEndpoint(ctx)
			if err != nil {
				t.Fatal(err)
			}

			var opt []func(*stomp.Conn) error
			if test.user != "" || test.pass != "" {
				opt = append(opt, stomp.ConnOpt.Login(test.user, test.pass))
			}

			conn, err := stomp.Dial("tcp", host, opt...)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { conn.Disconnect() })

			sub, err := conn.Subscribe("test", stomp.AckAuto)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { sub.Unsubscribe() })

			err = conn.Send("test", "", []byte("test"))
			if err != nil {
				t.Fatal(err)
			}

			ticker := time.NewTicker(10 * time.Second)
			select {
			case <-ticker.C:
				t.Fatal("timed out waiting for message")
			case msg := <-sub.C:
				if string(msg.Body) != "test" {
					t.Fatal("received unexpected message bytes")
				}
			}

			if test.hook != nil {
				test.hook(t, container)
			}
		})
	}
}

func expectQueue(t *testing.T, container *artemis.Container, queueName string) {
	t.Helper()

	u, err := container.ConsoleURL(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	r, err := http.Get(u + `/jolokia/read/org.apache.activemq.artemis:broker="0.0.0.0"/QueueNames`)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Body.Close()

	var res struct{ Value []string }
	if err = json.NewDecoder(r.Body).Decode(&res); err != nil {
		t.Fatal(err)
	}

	for _, v := range res.Value {
		if v == queueName {
			return
		}
	}

	t.Fatalf("should contain queue %q", queueName)
}
