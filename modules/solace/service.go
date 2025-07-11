package solace

type Service struct {
	Name       string
	Port       int
	Protocol   string
	SupportSSL bool
}

var (
	ServiceAMQP = Service{
		Name:       "amqp",
		Port:       5672,
		Protocol:   "amqp",
		SupportSSL: false,
	}
	ServiceMQTT = Service{
		Name:       "mqtt",
		Port:       1883,
		Protocol:   "tcp",
		SupportSSL: false,
	}
	ServiceREST = Service{
		Name:       "rest",
		Port:       9000,
		Protocol:   "http",
		SupportSSL: false,
	}
	ServiceManagement = Service{
		Name:       "management",
		Port:       8080,
		Protocol:   "http",
		SupportSSL: false,
	}
	ServiceSMF = Service{
		Name:       "smf",
		Port:       55555,
		Protocol:   "tcp",
		SupportSSL: true,
	}
	ServiceSMFSSL = Service{
		Name:       "smf",
		Port:       55443,
		Protocol:   "tcps",
		SupportSSL: true,
	}
)
