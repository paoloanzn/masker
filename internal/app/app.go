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
	engine        *audio.Engine
	generator     *noise.Generator
	status        *systray.MenuItem
	focus         *systray.MenuItem
	brown         *systray.MenuItem
	pink          *systray.MenuItem
	speech        *systray.MenuItem
	densityLow    *systray.MenuItem
	densityMedium *systray.MenuItem
	densityHigh   *systray.MenuItem
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
	systray.SetTooltip("Masker focus audio")

	if err := a.engine.Start(); err != nil {
		log.Fatalf("start audio engine: %v", err)
	}

	a.status = systray.AddMenuItem(a.statusText(), "Current masking status")
	a.status.Disable()

	systray.AddSeparator()

	a.focus = systray.AddMenuItemCheckbox("Focus (Recommended)", "Recommended slow-beat mode for general productivity", true)
	a.brown = systray.AddMenuItemCheckbox("Brown", "Masking option for low rumble, HVAC, or travel", false)
	a.pink = systray.AddMenuItemCheckbox("Pink", "Masking option for general ambient coverage", false)
	a.speech = systray.AddMenuItemCheckbox("Speech-shaped", "Masking option that targets the speech band more directly", false)

	systray.AddSeparator()

	a.densityLow = systray.AddMenuItemCheckbox("Density: Low", "Focus only: sparse pad with a very soft timing anchor", false)
	a.densityMedium = systray.AddMenuItemCheckbox("Density: Medium", "Focus only: recommended preset with a soft harmonic bed", true)
	a.densityHigh = systray.AddMenuItemCheckbox("Density: High", "Focus only: add very subtle background motion", false)

	systray.AddSeparator()

	volumeUp := systray.AddMenuItem("Volume +", "Increase masker volume")
	volumeDown := systray.AddMenuItem("Volume -", "Decrease masker volume")

	systray.AddSeparator()

	quit := systray.AddMenuItem("Quit", "Quit the app")

	a.syncUI()
	installTrackCommandHandlers(a.nextMode, a.previousMode, a.play, a.pause, a.togglePlayPause)

	go func() {
		for {
			select {
			case <-a.focus.ClickedCh:
				a.generator.SetMode(noise.ModeFocus)
			case <-a.brown.ClickedCh:
				a.generator.SetMode(noise.ModeBrown)
			case <-a.pink.ClickedCh:
				a.generator.SetMode(noise.ModePink)
			case <-a.speech.ClickedCh:
				a.generator.SetMode(noise.ModeSpeech)
			case <-a.densityLow.ClickedCh:
				a.generator.SetDensity(noise.DensityLow)
			case <-a.densityMedium.ClickedCh:
				a.generator.SetDensity(noise.DensityMedium)
			case <-a.densityHigh.ClickedCh:
				a.generator.SetDensity(noise.DensityHigh)
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
	a.focus.Uncheck()
	a.brown.Uncheck()
	a.pink.Uncheck()
	a.speech.Uncheck()

	switch mode {
	case noise.ModeFocus:
		a.focus.Check()
	case noise.ModeBrown:
		a.brown.Check()
	case noise.ModePink:
		a.pink.Check()
	case noise.ModeSpeech:
		a.speech.Check()
	}

	density := a.generator.Density()
	a.densityLow.Uncheck()
	a.densityMedium.Uncheck()
	a.densityHigh.Uncheck()

	switch density {
	case noise.DensityLow:
		a.densityLow.Check()
	case noise.DensityMedium:
		a.densityMedium.Check()
	case noise.DensityHigh:
		a.densityHigh.Check()
	}

	if mode == noise.ModeFocus {
		a.densityLow.Enable()
		a.densityMedium.Enable()
		a.densityHigh.Enable()
		return
	}

	a.densityLow.Disable()
	a.densityMedium.Disable()
	a.densityHigh.Disable()
}

func (a *App) syncUI() {
	if a.focus != nil && a.brown != nil && a.pink != nil && a.speech != nil && a.densityLow != nil && a.densityMedium != nil && a.densityHigh != nil {
		a.updateChecks()
	}
	if a.status != nil {
		a.status.SetTitle(a.statusText())
	}
	updateTrackCommandState(a.generator.Mode().String(), a.generator.Paused())
}

func (a *App) nextMode() {
	a.generator.SetMode(a.generator.Mode().Next())
	a.syncUI()
}

func (a *App) previousMode() {
	a.generator.SetMode(a.generator.Mode().Previous())
	a.syncUI()
}

func (a *App) play() {
	a.generator.SetPaused(false)
	a.syncUI()
}

func (a *App) pause() {
	a.generator.SetPaused(true)
	a.syncUI()
}

func (a *App) togglePlayPause() {
	a.generator.TogglePaused()
	a.syncUI()
}

func (a *App) statusText() string {
	state := "Playing"
	if a.generator.Paused() {
		state = "Paused"
	}
	mode := a.generator.Mode().String()
	if a.generator.Mode() == noise.ModeFocus {
		mode = fmt.Sprintf("%s (%s density)", mode, a.generator.Density().String())
	}
	return fmt.Sprintf("%s | Mode: %s | Vol: %.3f", state, mode, a.generator.Volume())
}
