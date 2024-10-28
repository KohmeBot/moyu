package moyu

type Config struct {
	// 发送的cron表达式
	SendCron string `mapstructure:"send_cron"`
	// 提示语
	Tips []string `mapstructure:"tips"`
}
