package couchbase

type Service struct {
	identifier     string
	minimumQuotaMb int
	ports          []string
}

func (s Service) hasQuota() bool {
	return s.minimumQuotaMb > 0
}

var (
	kv = Service{
		identifier:     "kv",
		minimumQuotaMb: 256,
		ports:          []string{KV_PORT, KV_SSL_PORT, VIEW_PORT, VIEW_SSL_PORT},
	}

	query = Service{
		identifier:     "n1ql",
		minimumQuotaMb: 0,
		ports:          []string{QUERY_PORT, QUERY_SSL_PORT},
	}

	search = Service{
		identifier:     "fts",
		minimumQuotaMb: 256,
		ports:          []string{SEARCH_PORT, SEARCH_SSL_PORT},
	}

	index = Service{
		identifier:     "index",
		minimumQuotaMb: 256,
	}

	analytics = Service{
		identifier:     "cbas",
		minimumQuotaMb: 256,
		ports:          []string{ANALYTICS_PORT, ANALYTICS_SSL_PORT},
	}

	eventing = Service{
		identifier:     "eventing",
		minimumQuotaMb: 256,
		ports:          []string{EVENTING_PORT, EVENTING_SSL_PORT},
	}
)
