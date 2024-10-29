package moyu

import (
	"fmt"
	"github.com/kohmebot/pkg/command"
	"github.com/kohmebot/pkg/version"
	"github.com/kohmebot/plugin"
	"github.com/robfig/cron/v3"
	"github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"io"
	"math/rand/v2"
	"net/http"
)

type PluginMoyu struct {
	conf Config
	env  plugin.Env
}

func NewPluginMoyu() plugin.Plugin {
	return new(PluginMoyu)
}

func (p *PluginMoyu) Init(engine *zero.Engine, env plugin.Env) error {
	err := env.GetConf(&p.conf)
	if err != nil {
		return err
	}
	p.env = env
	return nil
}

func (p *PluginMoyu) Name() string {
	return "moyu"
}

func (p *PluginMoyu) Description() string {
	return "每日发送摸鱼日历"
}

func (p *PluginMoyu) Commands() fmt.Stringer {
	return command.NewCommands()
}

func (p *PluginMoyu) Version() uint64 {
	return uint64(version.NewVersion(0, 0, 11))
}

func (p *PluginMoyu) OnBoot() {
	sendErr := func(err error) {
		for ctx := range p.env.RangeBot {
			p.env.Error(ctx, err)
		}
	}
	c := cron.New()
	_, err := c.AddFunc(p.conf.SendCron, func() {
		var err error
		defer func() {
			if err != nil {
				sendErr(err)
			}
		}()
		resp, err := http.Get("https://api.vvhan.com/api/moyu")
		if err != nil {
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("status code: %d: %s", resp.StatusCode, resp.Status)
			return
		}

		img, err := io.ReadAll(resp.Body)
		if err != nil {
			return
		}
		var text string
		if len(p.conf.Tips) > 0 {
			text = p.conf.Tips[rand.IntN(len(p.conf.Tips))]
		}
		textMsg := message.Text(text)
		imgMsg := message.ImageBytes(img)

		for ctx := range p.env.RangeBot {
			for gid := range p.env.Groups().RangeGroup {
				if len(text) > 0 {
					ctx.SendGroupMessage(gid, textMsg)
				}
				ctx.SendGroupMessage(gid, imgMsg)
			}
		}

	})
	if err != nil {
		sendErr(err)
		return
	}
	c.Start()
}
