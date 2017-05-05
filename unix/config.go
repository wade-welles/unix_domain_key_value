package unix

//Config ...
type Config struct {
	Daemon DaemonConf
}

//DaemonConf ...
type DaemonConf struct {
	Socket         string
	Cpus           int
	Updateinterval int
	Sourceendpoint string
	Apitoken       string
}
