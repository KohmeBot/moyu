package moyu

import (
	"fmt"
	"github.com/kohmebot/chatai/chatai/chataisdk"
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

	invoker *chataisdk.ChatAIInvoker
}

func NewPlugin() plugin.Plugin {
	return new(PluginMoyu)
}

func (p *PluginMoyu) Init(engine *zero.Engine, env plugin.Env) error {
	err := env.GetConf(&p.conf)
	if err != nil {
		return err
	}
	p.env = env
	if p.conf.UseAI {
		p.invoker, err = chataisdk.NewChatAIInvoker(env)
	}
	return err
}

func (p *PluginMoyu) GetTips(ctx *zero.Ctx) string {
	var text string
	if len(p.conf.Tips) > 0 {
		text = p.conf.Tips[rand.IntN(len(p.conf.Tips))]
	}
	if len(text) <= 0 || !p.conf.UseAI {
		return text
	}

	// useAi
	resp, err := p.invoker.DoRequest(text)
	if err != nil {
		p.env.Error(ctx, err)
		return text
	}
	return resp
}

func (p *PluginMoyu) GetImage() ([]byte, error) {
	resp, err := http.Get("https://api.vvhan.com/api/moyu")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d: %s", resp.StatusCode, resp.Status)
	}

	img, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return img, nil
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
	return uint64(version.NewVersion(0, 0, 28))
}

func (p *PluginMoyu) OnBoot() {
	sendErr := func(err error) {
		for ctx := range p.env.RangeBot {
			p.env.Error(ctx, err)
		}
	}
	c := cron.New()
	_, err := c.AddFunc(p.conf.SendCron, func() {

		img, err := p.GetImage()
		if err != nil {
			sendErr(err)
		}

		for ctx := range p.env.RangeBot {
			text := p.GetTips(ctx)
			for gid := range p.env.Groups().RangeGroup {
				if len(text) > 0 {
					ctx.SendGroupMessage(gid, message.Text(text))
				}
				if len(img) > 0 {
					ctx.SendGroupMessage(gid, message.ImageBytes(img))
				}

			}
		}

	})
	if err != nil {
		sendErr(err)
		return
	}
	c.Start()
}
