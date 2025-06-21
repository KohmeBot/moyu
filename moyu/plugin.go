package moyu

import (
	"fmt"
	"github.com/kohmebot/chatai/chatai/chataisdk"
	"github.com/kohmebot/chatai/chatai/model"
	"github.com/kohmebot/pkg/command"
	"github.com/kohmebot/pkg/version"
	"github.com/kohmebot/plugin"
	"github.com/robfig/cron/v3"
	"github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"io"
	"math/rand/v2"
	"net/http"
	"sync"
)

type PluginMoyu struct {
	conf  Config
	env   plugin.Env
	batch model.Batch

	on chan model.OnResponse
}

func NewPluginMoyu() plugin.Plugin {
	return new(PluginMoyu)
}

func (p *PluginMoyu) Init(engine *zero.Engine, env plugin.Env) error {
	p.on = make(chan model.OnResponse, 1)
	err := env.GetConf(&p.conf)
	if err != nil {
		return err
	}
	p.env = env
	if p.conf.UseAI {
		p.batch, err = chataisdk.NewBatch(p.env, p.OnAIResponse)
	}
	return nil
}

func (p *PluginMoyu) OnAIResponse(ctx *zero.Ctx, request *model.Request, response *model.Response, err error) {
	select {
	case on := <-p.on:
		on(ctx, request, response, err)
	default:
	}
}

func (p *PluginMoyu) GetTips(ctx *zero.Ctx) string {
	var text string
	if len(p.conf.Tips) > 0 {
		text = p.conf.Tips[rand.IntN(len(p.conf.Tips))]
	}
	if len(text) <= 0 || !p.conf.UseAI {
		return text
	}
	// use AI
	wg := sync.WaitGroup{}
	wg.Add(1)
	p.on <- func(ctx *zero.Ctx, request *model.Request, response *model.Response, err error) {
		defer wg.Done()
		if err != nil {
			p.env.Error(ctx, err)
			return
		}
		if len(response.ErrorMsg) > 0 {
			p.env.Error(ctx, fmt.Errorf(response.ErrorMsg))
			return
		}
		text = response.Answer
	}
	p.batch.Submit(ctx, model.Key{}, []string{text})
	wg.Wait()

	return text
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
	return uint64(version.NewVersion(0, 0, 20))
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

		imgMsg := message.ImageBytes(img)

		for ctx := range p.env.RangeBot {
			text := p.GetTips(ctx)
			textMsg := message.Text(text)
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
