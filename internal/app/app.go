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
	engine          *audio.Engine
	generator       *noise.Generator
	status          *systray.MenuItem
	focus           *systray.MenuItem
	adhd            *systray.MenuItem
	brown           *systray.MenuItem
	speech          *systray.MenuItem
	presetLow       *systray.MenuItem
	presetMedium    *systray.MenuItem
	presetHigh      *systray.MenuItem
	presetCognitive *systray.MenuItem
	adhdWhite       *systray.MenuItem
	adhdPink        *systray.MenuItem
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
	a.adhd = systray.AddMenuItemCheckbox("ADHD / attention problems", "Steady-state white or pink noise presets intended for ADHD-like inattentive symptoms", false)
	a.brown = systray.AddMenuItemCheckbox("Brown", "Masking option for low rumble, HVAC, or travel", false)
	a.speech = systray.AddMenuItemCheckbox("Speech-shaped", "Masking option that targets the speech band more directly", false)

	systray.AddSeparator()

	a.presetLow = systray.AddMenuItemCheckbox("Preset: Low", "Focus only: sparse pad plus a soft beat-synced pulse bed", false)
	a.presetMedium = systray.AddMenuItemCheckbox("Preset: Medium", "Focus only: recommended preset with a soft harmonic bed and structured pulse overlay", true)
	a.presetHigh = systray.AddMenuItemCheckbox("Preset: High", "Focus only: add very subtle background motion while preserving the pulse scaffold", false)
	a.presetCognitive = systray.AddMenuItemCheckbox("Preset: High cognitive load", "Focus only: same BPM range with reduced harmonic motion and lower melodic novelty", false)
	a.adhdWhite = systray.AddMenuItemCheckbox("ADHD preset: White", "ADHD only: steady broadband white noise with flat spectral density; start low", false)
	a.adhdPink = systray.AddMenuItemCheckbox("ADHD preset: Pink", "ADHD only: steady pink noise with an approximate 1/f spectral tilt; start low", false)

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
			case <-a.adhd.ClickedCh:
				a.generator.SetMode(noise.ModeADHD)
			case <-a.brown.ClickedCh:
				a.generator.SetMode(noise.ModeBrown)
			case <-a.speech.ClickedCh:
				a.generator.SetMode(noise.ModeSpeech)
			case <-a.presetLow.ClickedCh:
				a.generator.SetFocusPreset(noise.FocusPresetLow)
			case <-a.presetMedium.ClickedCh:
				a.generator.SetFocusPreset(noise.FocusPresetMedium)
			case <-a.presetHigh.ClickedCh:
				a.generator.SetFocusPreset(noise.FocusPresetHigh)
			case <-a.presetCognitive.ClickedCh:
				a.generator.SetFocusPreset(noise.FocusPresetHighCognitiveLoad)
			case <-a.adhdWhite.ClickedCh:
				a.generator.SetADHDPreset(noise.ADHDPresetWhite)
			case <-a.adhdPink.ClickedCh:
				a.generator.SetADHDPreset(noise.ADHDPresetPink)
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
	a.adhd.Uncheck()
	a.brown.Uncheck()
	a.speech.Uncheck()

	switch mode {
	case noise.ModeFocus:
		a.focus.Check()
	case noise.ModeADHD:
		a.adhd.Check()
	case noise.ModeBrown:
		a.brown.Check()
	case noise.ModeSpeech:
		a.speech.Check()
	}

	focusPreset := a.generator.FocusPreset()
	a.presetLow.Uncheck()
	a.presetMedium.Uncheck()
	a.presetHigh.Uncheck()
	a.presetCognitive.Uncheck()
	a.adhdWhite.Uncheck()
	a.adhdPink.Uncheck()

	switch focusPreset {
	case noise.FocusPresetLow:
		a.presetLow.Check()
	case noise.FocusPresetMedium:
		a.presetMedium.Check()
	case noise.FocusPresetHigh:
		a.presetHigh.Check()
	case noise.FocusPresetHighCognitiveLoad:
		a.presetCognitive.Check()
	}

	switch a.generator.ADHDPreset() {
	case noise.ADHDPresetWhite:
		a.adhdWhite.Check()
	case noise.ADHDPresetPink:
		a.adhdPink.Check()
	}

	if mode == noise.ModeFocus {
		a.presetLow.Enable()
		a.presetMedium.Enable()
		a.presetHigh.Enable()
		a.presetCognitive.Enable()
	} else {
		a.presetLow.Disable()
		a.presetMedium.Disable()
		a.presetHigh.Disable()
		a.presetCognitive.Disable()
	}

	if mode == noise.ModeADHD {
		a.adhdWhite.Enable()
		a.adhdPink.Enable()
		return
	}

	a.adhdWhite.Disable()
	a.adhdPink.Disable()
}

func (a *App) syncUI() {
	if a.focus != nil && a.adhd != nil && a.brown != nil && a.speech != nil && a.presetLow != nil && a.presetMedium != nil && a.presetHigh != nil && a.presetCognitive != nil && a.adhdWhite != nil && a.adhdPink != nil {
		a.updateChecks()
	}
	if a.status != nil {
		a.status.SetTitle(a.statusText())
	}
	updateTrackCommandState(a.currentModeText(), a.generator.Paused())
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
	return fmt.Sprintf("%s | Mode: %s | Vol: %.3f", state, a.currentModeText(), a.generator.Volume())
}

func (a *App) currentModeText() string {
	mode := a.generator.Mode().String()

	switch a.generator.Mode() {
	case noise.ModeFocus:
		return fmt.Sprintf("%s (%s preset)", mode, a.generator.FocusPreset().String())
	case noise.ModeADHD:
		return fmt.Sprintf("%s (%s preset)", mode, a.generator.ADHDPreset().String())
	default:
		return mode
	}
}
