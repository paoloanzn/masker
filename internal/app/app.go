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
	status    *systray.MenuItem
	brown     *systray.MenuItem
	pink      *systray.MenuItem
	speech    *systray.MenuItem
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

	a.status = systray.AddMenuItem(a.statusText(), "Current masking status")
	a.status.Disable()

	systray.AddSeparator()

	a.brown = systray.AddMenuItemCheckbox("Brown", "Low rumble / HVAC / travel", true)
	a.pink = systray.AddMenuItemCheckbox("Pink", "General ambient masking", false)
	a.speech = systray.AddMenuItemCheckbox("Speech-shaped", "Target the speech band more directly", false)

	systray.AddSeparator()

	volumeUp := systray.AddMenuItem("Volume +", "Increase masker volume")
	volumeDown := systray.AddMenuItem("Volume -", "Decrease masker volume")

	systray.AddSeparator()

	quit := systray.AddMenuItem("Quit", "Quit the app")

	a.syncUI()
	installTrackCommandHandlers(a.nextMode, a.previousMode)

	go func() {
		for {
			select {
			case <-a.brown.ClickedCh:
				a.generator.SetMode(noise.ModeBrown)
			case <-a.pink.ClickedCh:
				a.generator.SetMode(noise.ModePink)
			case <-a.speech.ClickedCh:
				a.generator.SetMode(noise.ModeSpeech)
			case <-volumeUp.ClickedCh:
				a.generator.SetVolume(a.generator.Volume() + config.VolumeStep)
			case <-volumeDown.ClickedCh:
				a.generator.SetVolume(a.generator.Volume() - config.VolumeStep)
			case <-quit.ClickedCh:
				systray.Quit()
				return
			}

			a.syncUI()
		}
	}()
}

func (a *App) onExit() {
	clearTrackCommandHandlers()
	a.engine.Stop()
}

func (a *App) updateChecks() {
	mode := a.generator.Mode()
	a.brown.Uncheck()
	a.pink.Uncheck()
	a.speech.Uncheck()

	switch mode {
	case noise.ModeBrown:
		a.brown.Check()
	case noise.ModePink:
		a.pink.Check()
	case noise.ModeSpeech:
		a.speech.Check()
	}
}

func (a *App) syncUI() {
	if a.brown != nil && a.pink != nil && a.speech != nil {
		a.updateChecks()
	}
	if a.status != nil {
		a.status.SetTitle(a.statusText())
	}
	updateTrackCommandMode(a.generator.Mode().String())
}

func (a *App) nextMode() {
	a.generator.SetMode(a.generator.Mode().Next())
	a.syncUI()
}

func (a *App) previousMode() {
	a.generator.SetMode(a.generator.Mode().Previous())
	a.syncUI()
}

func (a *App) statusText() string {
	return fmt.Sprintf("Mode: %s | Vol: %.3f", a.generator.Mode(), a.generator.Volume())
}
