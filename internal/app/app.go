package app

import (
	"fmt"
	"log"

	"github.com/getlantern/systray"

	"masker/internal/audio"
	"masker/internal/config"
	"masker/internal/noise"
)

type App struct {
	engine    *audio.Engine
	generator *noise.Generator
}

func New() *App {
	generator := noise.NewGenerator()
	return &App{
		engine:    audio.NewEngine(generator),
		generator: generator,
	}
}

func Run() error {
	preparePlatform()
	maskerApp := New()
	systray.Run(maskerApp.onReady, maskerApp.onExit)
	return nil
}

func (a *App) onReady() {
	systray.SetIcon(trayIcon)
	systray.SetTitle("")
	systray.SetTooltip("Masker noise masking")

	if err := a.engine.Start(); err != nil {
		log.Fatalf("start audio engine: %v", err)
	}

	status := systray.AddMenuItem(a.statusText(), "Current masking status")
	status.Disable()

	systray.AddSeparator()

	brown := systray.AddMenuItemCheckbox("Brown", "Low rumble / HVAC / travel", true)
	pink := systray.AddMenuItemCheckbox("Pink", "General ambient masking", false)
	speech := systray.AddMenuItemCheckbox("Speech-shaped", "Target the speech band more directly", false)

	systray.AddSeparator()

	volumeUp := systray.AddMenuItem("Volume +", "Increase masker volume")
	volumeDown := systray.AddMenuItem("Volume -", "Decrease masker volume")

	systray.AddSeparator()

	quit := systray.AddMenuItem("Quit", "Quit the app")

	a.updateChecks(brown, pink, speech)

	go func() {
		for {
			select {
			case <-brown.ClickedCh:
				a.generator.SetMode(noise.ModeBrown)
			case <-pink.ClickedCh:
				a.generator.SetMode(noise.ModePink)
			case <-speech.ClickedCh:
				a.generator.SetMode(noise.ModeSpeech)
			case <-volumeUp.ClickedCh:
				a.generator.SetVolume(a.generator.Volume() + config.VolumeStep)
			case <-volumeDown.ClickedCh:
				a.generator.SetVolume(a.generator.Volume() - config.VolumeStep)
			case <-quit.ClickedCh:
				systray.Quit()
				return
			}

			a.updateChecks(brown, pink, speech)
			status.SetTitle(a.statusText())
		}
	}()
}

func (a *App) onExit() {
	a.engine.Stop()
}

func (a *App) updateChecks(brown, pink, speech *systray.MenuItem) {
	mode := a.generator.Mode()
	brown.Uncheck()
	pink.Uncheck()
	speech.Uncheck()

	switch mode {
	case noise.ModeBrown:
		brown.Check()
	case noise.ModePink:
		pink.Check()
	case noise.ModeSpeech:
		speech.Check()
	}
}

func (a *App) statusText() string {
	return fmt.Sprintf("Mode: %s | Vol: %.3f", a.generator.Mode(), a.generator.Volume())
}
