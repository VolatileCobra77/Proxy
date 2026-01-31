package main

import (
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

func main() {
	go func() {
		window := new(app.Window)
		window.Option(app.Title("Proxy GUI"))
		window.Option(app.Size(unit.Dp(600), unit.Dp(400)))

		var ops op.Ops

		var toggleButton widget.Clickable

		currentTheme := material.NewTheme()

		for {
			event := window.Event()

			switch eventType := event.(type) {

			case app.FrameEvent:
				context := app.NewContext(&ops, eventType)

				layout.Flex{
					Axis:    layout.Vertical,
					Spacing: layout.SpaceStart,
				}.Layout(
					context,
					layout.Rigid(
						func(context layout.Context) layout.Dimensions {
							button := material.Button(currentTheme, &toggleButton, "Start Proxy")
							return button.Layout(context)
						},
					),
					layout.Rigid(
						// The height of the spacer is 25 Device independent pixels
						layout.Spacer{Height: unit.Dp(25)}.Layout,
					),
				)

				eventType.Frame(context.Ops)

			case app.DestroyEvent:
				os.Exit(0)

			}

		}
	}()
	app.Main()
}
