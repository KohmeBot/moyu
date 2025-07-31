package moyu

type Config struct {
	// 发送的cron表达式
	SendCron string `yaml:"send_cron"`
	// 提示语
	Tips []string `yaml:"tips"`
	// 是否使用AI
	UseAI bool `yaml:"use_ai"`
}
