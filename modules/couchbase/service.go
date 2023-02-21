package couchbase

type service struct {
	identifier     string
	minimumQuotaMb int
	ports          []string
}

func (s service) hasQuota() bool {
	return s.minimumQuotaMb > 0
}

var (
	kv = service{
		identifier:     "kv",
		minimumQuotaMb: 256,
		ports:          []string{KV_PORT, KV_SSL_PORT, VIEW_PORT, VIEW_SSL_PORT},
	}

	query = service{
		identifier:     "n1ql",
		minimumQuotaMb: 0,
		ports:          []string{QUERY_PORT, QUERY_SSL_PORT},
	}

	search = service{
		identifier:     "fts",
		minimumQuotaMb: 256,
		ports:          []string{SEARCH_PORT, SEARCH_SSL_PORT},
	}

	index = service{
		identifier:     "index",
		minimumQuotaMb: 256,
	}

	analytics = service{
		identifier:     "cbas",
		minimumQuotaMb: 256,
		ports:          []string{ANALYTICS_PORT, ANALYTICS_SSL_PORT},
	}

	eventing = service{
		identifier:     "eventing",
		minimumQuotaMb: 256,
		ports:          []string{EVENTING_PORT, EVENTING_SSL_PORT},
	}
)
