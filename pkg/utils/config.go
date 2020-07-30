package utils

var C *Config

type Config struct {
	*Base
	*Redis
	*Mysql
	*Etcd
	*Grpc
}

type Base struct {
	Name    string
	Version string
	Webhost string
}

type Redis struct {
	Host     string
	Password string
	Database int
}

type Mysql struct {
	Host     string
	User     string
	Password string
	Database string
}

type Etcd struct {
	Host []string
}

type Grpc struct {
	Name string
	Host string
	Port int
}
