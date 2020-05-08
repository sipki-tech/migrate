package zergrepo

// Option to connect to the database.
type Option func(*Config)

// Name sets the connection parameters.
func Name(name string) Option {
	return func(config *Config) {
		config.DBName = name
	}
}

// User sets the connection parameters.
func User(user string) Option {
	return func(config *Config) {
		config.User = user
	}
}

// Pass sets the connection parameters.
func Pass(pass string) Option {
	return func(config *Config) {
		config.Password = pass
	}
}

// Host sets the connection parameters.
func Host(host string) Option {
	return func(config *Config) {
		config.Host = host
	}
}

// SSLMode sets the connection parameters.
func SSLMode(sslMode string) Option {
	return func(config *Config) {
		config.SSLMode = sslMode
	}
}

// Port sets the connection parameters.
func Port(port int) Option {
	return func(config *Config) {
		config.Port = port
	}
}
