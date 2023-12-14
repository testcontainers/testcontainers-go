package rabbitmq_test

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/testcontainers/testcontainers-go"
)

// The following structs are added as a demonstration for the RabbitMQ management API therefore,
// they are not used in the RabbitMQ module.
// All of them implement the testcontainers.Executable interface, which is used to generate
// the command that will be executed, with the "AsCommand" method.
// Please be aware that they could be outdated, as they are not actively maintained, just here for reference.

// --------- Bindings ---------

type Binding struct {
	testcontainers.ExecOptions
	VHost           string
	Source          string
	Destination     string
	DestinationType string
	RoutingKey      string
	// additional arguments, that will be serialized to JSON when passed to the container
	Args map[string]interface{}
}

func NewBinding(source string, destination string) Binding {
	return Binding{
		Source:      source,
		Destination: destination,
	}
}

func NewBindingWithVHost(vhost string, source string, destination string) Binding {
	return Binding{
		VHost:       vhost,
		Source:      source,
		Destination: destination,
	}
}

func (b Binding) AsCommand() []string {
	cmd := []string{"rabbitmqadmin"}

	if b.VHost != "" {
		cmd = append(cmd, fmt.Sprintf("--vhost=%s", b.VHost))
	}

	cmd = append(cmd, "declare", "binding", fmt.Sprintf("source=%s", b.Source), fmt.Sprintf("destination=%s", b.Destination))

	if b.DestinationType != "" {
		cmd = append(cmd, fmt.Sprintf("destination_type=%s", b.DestinationType))
	}
	if b.RoutingKey != "" {
		cmd = append(cmd, fmt.Sprintf("routing_key=%s", b.RoutingKey))
	}

	if len(b.Args) > 0 {
		bytes, err := json.Marshal(b.Args)
		if err != nil {
			return cmd
		}

		cmd = append(cmd, "arguments="+string(bytes))
	}

	return cmd
}

// --------- Bindings ---------

// --------- Exchange ---------

type Exchange struct {
	testcontainers.ExecOptions
	Name       string
	VHost      string
	Type       string
	AutoDelete bool
	Internal   bool
	Durable    bool
	Args       map[string]interface{}
}

func (e Exchange) AsCommand() []string {
	cmd := []string{"rabbitmqadmin"}

	if e.VHost != "" {
		cmd = append(cmd, "--vhost="+e.VHost)
	}

	cmd = append(cmd, "declare", "exchange", fmt.Sprintf("name=%s", e.Name), fmt.Sprintf("type=%s", e.Type))

	if e.AutoDelete {
		cmd = append(cmd, "auto_delete=true")
	}
	if e.Internal {
		cmd = append(cmd, "internal=true")
	}
	if e.Durable {
		cmd = append(cmd, fmt.Sprintf("durable=%t", e.Durable))
	}

	if len(e.Args) > 0 {
		bytes, err := json.Marshal(e.Args)
		if err != nil {
			return cmd
		}

		cmd = append(cmd, "arguments="+string(bytes))
	}

	return cmd
}

// --------- Exchange ---------

// --------- OperatorPolicy ---------

type OperatorPolicy struct {
	testcontainers.ExecOptions
	Name       string
	Pattern    string
	Definition map[string]interface{}
	Priority   int
	ApplyTo    string
}

func (op OperatorPolicy) AsCommand() []string {
	cmd := []string{"rabbitmqadmin", "declare", "operator_policy", fmt.Sprintf("name=%s", op.Name), fmt.Sprintf("pattern=%s", op.Pattern)}

	if op.Priority > 0 {
		cmd = append(cmd, fmt.Sprintf("priority=%d", op.Priority))
	}
	if op.ApplyTo != "" {
		cmd = append(cmd, fmt.Sprintf("apply-to=%s", op.ApplyTo))
	}

	if len(op.Definition) > 0 {
		bytes, err := json.Marshal(op.Definition)
		if err != nil {
			return cmd
		}

		cmd = append(cmd, "definition="+string(bytes))
	}

	return cmd
}

// --------- OperatorPolicy ---------

// --------- Parameter ---------

type Parameter struct {
	testcontainers.ExecOptions
	Component string
	Name      string
	Value     string
}

func NewParameter(component string, name string, value string) Parameter {
	return Parameter{
		Component: component,
		Name:      name,
		Value:     value,
	}
}

func (p Parameter) AsCommand() []string {
	return []string{
		"rabbitmqadmin", "declare", "parameter",
		fmt.Sprintf("component=%s", p.Component), fmt.Sprintf("name=%s", p.Name), fmt.Sprintf("value=%s", p.Value),
	}
}

// --------- Parameter ---------

// --------- Permission ---------

type Permission struct {
	testcontainers.ExecOptions
	VHost     string
	User      string
	Configure string
	Write     string
	Read      string
}

func NewPermission(vhost string, user string, configure string, write string, read string) Permission {
	return Permission{
		VHost:     vhost,
		User:      user,
		Configure: configure,
		Write:     write,
		Read:      read,
	}
}

func (p Permission) AsCommand() []string {
	return []string{
		"rabbitmqadmin", "declare", "permission",
		fmt.Sprintf("vhost=%s", p.VHost), fmt.Sprintf("user=%s", p.User),
		fmt.Sprintf("configure=%s", p.Configure), fmt.Sprintf("write=%s", p.Write), fmt.Sprintf("read=%s", p.Read),
	}
}

// --------- Permission ---------

// --------- Plugin ---------

type Plugin struct {
	testcontainers.ExecOptions
	Name string
}

func (p Plugin) AsCommand() []string {
	return []string{"rabbitmq-plugins", "enable", p.Name}
}

// --------- Plugin ---------

// --------- Policy ---------

type Policy struct {
	testcontainers.ExecOptions
	VHost      string
	Name       string
	Pattern    string
	Definition map[string]interface{}
	Priority   int
	ApplyTo    string
}

func (p Policy) AsCommand() []string {
	cmd := []string{"rabbitmqadmin"}

	if p.VHost != "" {
		cmd = append(cmd, "--vhost="+p.VHost)
	}

	cmd = append(cmd, "declare", "policy", fmt.Sprintf("name=%s", p.Name), fmt.Sprintf("pattern=%s", p.Pattern))

	if p.Priority > 0 {
		cmd = append(cmd, fmt.Sprintf("priority=%d", p.Priority))
	}
	if p.ApplyTo != "" {
		cmd = append(cmd, fmt.Sprintf("apply-to=%s", p.ApplyTo))
	}

	if len(p.Definition) > 0 {
		bytes, err := json.Marshal(p.Definition)
		if err != nil {
			return cmd
		}

		cmd = append(cmd, "definition="+string(bytes))
	}

	return cmd
}

// --------- Policy ---------

// --------- Queue ---------

type Queue struct {
	testcontainers.ExecOptions
	Name       string
	VHost      string
	AutoDelete bool
	Durable    bool
	Args       map[string]interface{}
}

func (q Queue) AsCommand() []string {
	cmd := []string{"rabbitmqadmin"}

	if q.VHost != "" {
		cmd = append(cmd, "--vhost="+q.VHost)
	}

	cmd = append(cmd, "declare", "queue", fmt.Sprintf("name=%s", q.Name))

	if q.AutoDelete {
		cmd = append(cmd, "auto_delete=true")
	}
	if q.Durable {
		cmd = append(cmd, fmt.Sprintf("durable=%t", q.Durable))
	}

	if len(q.Args) > 0 {
		bytes, err := json.Marshal(q.Args)
		if err != nil {
			return cmd
		}

		cmd = append(cmd, "arguments="+string(bytes))
	}

	return cmd
}

// --------- Queue ---------

// --------- User ---------

type User struct {
	testcontainers.ExecOptions
	Name     string
	Password string
	Tags     []string
}

func (u User) AsCommand() []string {
	tagsMap := make(map[string]bool)
	for _, tag := range u.Tags {
		tagsMap[tag] = true
	}

	uniqueTags := make([]string, 0, len(tagsMap))
	for tag := range tagsMap {
		uniqueTags = append(uniqueTags, tag)
	}

	return []string{
		"rabbitmqadmin", "declare", "user",
		fmt.Sprintf("name=%s", u.Name), fmt.Sprintf("password=%s", u.Password),
		fmt.Sprintf("tags=%s", strings.Join(uniqueTags, ",")),
	}
}

// --------- User ---------

// --------- Virtual Hosts --------

type VirtualHost struct {
	testcontainers.ExecOptions
	Name    string
	Tracing bool
}

func (v VirtualHost) AsCommand() []string {
	cmd := []string{"rabbitmqadmin", "declare", "vhost", fmt.Sprintf("name=%s", v.Name)}

	if v.Tracing {
		cmd = append(cmd, "tracing=true")
	}

	return cmd
}

type VirtualHostLimit struct {
	testcontainers.ExecOptions
	VHost string
	Name  string
	Value int
}

func (v VirtualHostLimit) AsCommand() []string {
	return []string{"rabbitmqadmin", "declare", "vhost_limit", fmt.Sprintf("vhost=%s", v.VHost), fmt.Sprintf("name=%s", v.Name), fmt.Sprintf("value=%d", v.Value)}
}

// --------- Virtual Hosts ---------
