package main

import (
	"Proxy/cmd/client/core"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

var isRunning bool = false
var proxyManager core.ProxyManager

var (
	toggleButton widget.Clickable
	statusText   = "Status: Offline"
	buttonText   = "Start"
)

func main() {
	go func() {
		window := new(app.Window)
		proxyManager = core.ProxyManager{}
		window.Option(app.Title("Proxy GUI"))
		window.Option(app.Size(unit.Dp(600), unit.Dp(400)))
		if err := draw(window); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func draw(window *app.Window) error {
	var ops op.Ops

	currentTheme := material.NewTheme()

	for {
		event := window.Event()

		switch eventType := event.(type) {

		case app.FrameEvent:
			context := app.NewContext(&ops, eventType)
			if toggleButton.Clicked(context) {
				log.Println("CLICKED")
				if !isRunning {
					config, err := core.LoadConfig("clientConfigs.json")
					if err != nil {
						log.Fatal(err)
					}
					err = proxyManager.Start(config)
					if err != nil {
						return err
					}
					statusText = "Status: Online"
					buttonText = "Stop"
					isRunning = true
				} else {
					err := proxyManager.Stop()
					if err != nil {
						return err
					}
					statusText = "Status: Offline"
					buttonText = "Start"
					isRunning = false
				}
			}
			layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceStart,
			}.Layout(
				context,
				layout.Rigid(
					func(context layout.Context) layout.Dimensions {
						margins := layout.Inset{
							Top:    unit.Dp(25),
							Bottom: unit.Dp(25),
							Right:  unit.Dp(35),
							Left:   unit.Dp(35),
						}
						// TWO: ... then we lay out those margins ...
						return margins.Layout(context,
							// THREE: ... and finally within the margins, we ddefine and lay out the button
							func(gtx layout.Context) layout.Dimensions {
								btn := material.Button(currentTheme, &toggleButton, buttonText)
								return btn.Layout(gtx)
							},
						)
					},
				),
				layout.Rigid(
					// The height of the spacer is 25 Device independent pixels
					layout.Spacer{Height: unit.Dp(25)}.Layout,
				),
				layout.Rigid(
					func(context layout.Context) layout.Dimensions {
						margins := layout.Inset{
							Top:    unit.Dp(25),
							Bottom: unit.Dp(25),
							Right:  unit.Dp(35),
							Left:   unit.Dp(35),
						}
						return margins.Layout(context,
							func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(currentTheme, unit.Sp(32), statusText)
								return lbl.Layout(context)
							})

					},
				),
			)

			eventType.Frame(context.Ops)

		case app.DestroyEvent:
			os.Exit(0)

		}

	}
}
