package kredis

type Config struct {
	Addrs        []string `yaml:"addrs"`
	DialTimeout  int64    `yaml:"dialTimeOut"`  // in ms
	ReadTimeout  int64    `yaml:"readTimeOut"`  // in ms
	WriteTimeout int64    `yaml:"writeTimeOut"` // in ms
}
