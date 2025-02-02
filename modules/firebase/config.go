package firebase

// Based on https://github.com/firebase/firebase-tools/blob/master/src/firebaseConfig.ts

type emulatorsConfig struct {
	SingleProjectMode bool `json:"singleProjectMode,omitempty"`

	Auth struct {
		Host string `json:"host,omitempty"`
		Port int    `json:"port,omitempty"`
	} `json:"auth,omitempty"`

	Database struct {
		Host string `json:"host,omitempty"`
		Port int    `json:"port,omitempty"`
	} `json:"database,omitempty"`

	Firestore struct {
		Host          string `json:"host,omitempty"`
		Port          int    `json:"port,omitempty"`
		WebsocketPort int    `json:"websocketPort,omitempty"`
	} `json:"firestore,omitempty"`

	Functions struct {
		Host string `json:"host,omitempty"`
		Port int    `json:"port,omitempty"`
	} `json:"functions,omitempty"`

	Hosting struct {
		Host string `json:"host,omitempty"`
		Port int    `json:"port,omitempty"`
	} `json:"hosting,omitempty"`

	AppHosting struct {
		Host          string `json:"host,omitempty"`
		Port          int    `json:"port,omitempty"`
		StartCommand  string `json:"startCommand,omitempty"`
		RootDirectory string `json:"rootDirectory,omitempty"`
	} `json:"apphosting,omitempty"`

	PubSub struct {
		Host string `json:"host,omitempty"`
		Port int    `json:"port,omitempty"`
	} `json:"pubsub,omitempty"`

	Storage struct {
		Host string `json:"host,omitempty"`
		Port int    `json:"port,omitempty"`
	} `json:"storage,omitempty"`

	Logging struct {
		Host string `json:"host,omitempty"`
		Port int    `json:"port,omitempty"`
	} `json:"logging,omitempty"`

	Hub struct {
		Host string `json:"host,omitempty"`
		Port int    `json:"port,omitempty"`
	} `json:"hub,omitempty"`

	UI struct {
		Enabled bool   `json:"enabled,omitempty"`
		Host    string `json:"host,omitempty"`
		Port    int    `json:"port,omitempty"`
	} `json:"ui,omitempty"`

	EventArc struct {
		Host string `json:"host,omitempty"`
		Port int    `json:"port,omitempty"`
	} `json:"eventarc,omitempty"`

	DataConnect struct {
		Host         string `json:"host,omitempty"`
		Port         int    `json:"port,omitempty"`
		PostgresHost string `json:"postgresHost,omitempty"`
		PostgresPort int    `json:"postgresPort,omitempty"`
		DataDir      string `json:"dataDir,omitempty"`
	} `json:"dataconnect,omitempty"`

	Tasks struct {
		Host string `json:"host,omitempty"`
		Port int    `json:"port,omitempty"`
	} `json:"tasks,omitempty"`
}
type partialFirebaseConfig struct {
	Emulators emulatorsConfig `json:"emulators,omitempty"`
}
