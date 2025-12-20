package conf

type Bootstrap struct {
	Server *Server
	Data   *Data
	Auth   *Auth
}

type Auth struct {
	JwtSecret string
	JwtExpiry string
}

type Server struct {
	Http *ServerHTTP
	Grpc *ServerGRPC
}

type ServerHTTP struct {
	Addr    string
	Timeout string
}

type ServerGRPC struct {
	Addr    string
	Timeout string
}

type Data struct {
	Database *Database
	Redis    *Redis
}

type Database struct {
	Driver string
	Source string
}

type Redis struct {
	Addr         string
	Password     string
	ReadTimeout  string
	WriteTimeout string
}
