package main

import (
	"fmt"
	"image"
	"net/http"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"git.iglou.eu/adrien/inadl/ina"
)

type Config struct {
	url  string
	path fyne.ListableURI
	icon fyne.Resource
}

const (
	programVs   string = "V 0.0.1"
	programID   string = "eu.git.iglou.adrien.inadl"
	programName string = "INA video backup"
	programBy   string = "Open-source software by Iglou.eu"
)

func main() {
	var video ina.MRSS
	var err error

	// Init
	a := app.NewWithID(programID)
	a.SetIcon(fyne.NewStaticResource("Icon", appIcon))

	w := a.NewWindow(programName)
	w.CenterOnScreen()
	w.Resize(fyne.NewSize(1000, 700))

	// Core section
	var cCoreImageData image.Image
	cCoreImage := canvas.NewImageFromImage(cCoreImageData)
	cCoreImage.FillMode = canvas.ImageFillContain
	cCoreImage.ScaleMode = canvas.ImageScaleSmooth
	cCoreImage.SetMinSize(fyne.NewSize(300, 300))

	var wCoreQualSelect *widget.Select
	wCoreQualSelect = widget.NewSelect(nil, func(s string) { fmt.Println(s) })
	wCoreQualSelect.PlaceHolder = "Meilleure qualité possible"

	var wCoreDownload *widget.Button
	wCoreDownload = widget.NewButton(
		"Télécharger",
		func() {
			wCoreDownload.Text = "En cours ..."
			wCoreDownload.Disable()
			wCoreDownload.Refresh()
		},
	)

	wCoreDesc := widget.NewMultiLineEntry()
	wCoreDesc.Wrapping = fyne.TextWrapWord
	wCoreDescBox := widget.NewCard("", "", wCoreDesc)

	cCoreOption := container.NewVBox(wCoreQualSelect, wCoreDownload)
	cCoreContent := container.New(
		layout.NewFormLayout(),
		widget.NewCard("", "", fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, cCoreOption, nil, nil), cCoreOption, cCoreImage)),
		wCoreDescBox,
	)

	wCoreLoading := widget.NewTextGrid()
	cCoreLoad := container.NewCenter(wCoreLoading)

	wCoreBox := container.NewMax(cCoreLoad)

	// Header section
	wHeadUrlLabel := widget.NewLabel("Collez votre url ina.fr: ")
	wHeadUrlLabel.TextStyle.Bold = true
	wHeadUrlEntry := widget.NewEntry()
	wHeadUrlEntry.PlaceHolder = "https://www.ina.fr/video/PUB232175070/playstation-lancement-video.html"
	wHeadUrlEntry.TextStyle.Monospace = true
	wHeadUrlEntry.OnChanged = func(s string) {
		// Reset
		wCoreBox.Objects = nil
		wCoreBox.Add(cCoreLoad)
		wCoreBox.Refresh()

		// Loading
		wCoreLoading.SetText("Chargement . . .")
		wCoreLoading.Refresh()

		// Get video data
		video, err = ina.MediaNew(s)
		if err != nil {
			wCoreLoading.SetText(fmt.Sprint(err))
			return
		}

		// Get/Set image
		if len(video.Channel.Item.Content.Thumbnail) > 0 {
			resp, err := http.Get(video.Channel.Item.Content.Thumbnail[0].URL)
			if err != nil {
				wCoreLoading.SetText("Impossible de recuperer l'image")
				return
			}
			defer resp.Body.Close()

			cCoreImageData, _, err = image.Decode(resp.Body)
			if err != nil {
				wCoreLoading.SetText("Impossible de decoder l'image")
				return
			}

			cCoreImage.Image = cCoreImageData
			cCoreImage.Refresh()
		}

		// Set quality
		var qlSet []string

		if video.Channel.Item.Content.Hq.URL != "" {
			qlSet = append(qlSet, "Haute qualité")
		}
		if video.Channel.Item.Content.Mq.URL != "" {
			qlSet = append(qlSet, "Moyenne qualité")
		}
		if video.Channel.Item.Content.Bq.URL != "" {
			qlSet = append(qlSet, "Basse qualité")
		}

		if qlSet == nil {
			wCoreQualSelect.PlaceHolder = "Téléchargement indisponible"
			wCoreDownload.Disable()
		} else {
			wCoreQualSelect.Options = qlSet
			wCoreQualSelect.Refresh()
		}

		// Set informations
		if video.Channel.Description == "" {
			video.Channel.Description = "<Pas de description disponible>"
		}

		wCoreDesc.SetText(fmt.Sprintf(
			"Categorie: %s\nPublié le: %s\n---\n\n%s\n",
			video.Channel.Category,
			video.Channel.PubDate,
			video.Channel.Description,
		))

		// Refresh box
		wCoreDescBox.Title = video.Channel.Title
		wCoreDescBox.Refresh()

		wCoreBox.Objects = nil
		wCoreBox.Add(cCoreContent)
		wCoreBox.Refresh()
	}

	wHeadDirLabel := widget.NewLabel("Emplacement de sauvegarde: ")
	wHeadDirLabel.TextStyle.Bold = true
	var wHeadDirSelect *widget.Button
	wHeadDirSelect = widget.NewButton(
		"Selectionner",
		func() {
			dialog.NewFolderOpen(
				func(list fyne.ListableURI, err error) {
					if err != nil {
						fyne.LogError("Error on selecting dir to send", err)
						dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
						return
					} else if list == nil {
						return
					}

					config.path = list
					wHeadDirSelect.Text = list.Path()
					wHeadDirSelect.Refresh()
				},
				w,
			).Show()
		},
	)

	cHead := container.New(
		layout.NewFormLayout(),
		container.NewVBox(wHeadUrlLabel, wHeadDirLabel),
		container.NewVBox(wHeadUrlEntry, wHeadDirSelect),
	)
	wHeadBox := widget.NewCard("Outil de sauvegarde ina.fr", "", cHead)

	// Footer section
	wFooterV := widget.NewLabel(programVs)
	wFooterV.Alignment = fyne.TextAlignTrailing
	wFooterBy := widget.NewLabel(programBy)

	cFooter := container.NewGridWithColumns(2, wFooterBy, wFooterV)
	wFooterBox := widget.NewCard("", "", cFooter)

	// Render
	w.SetContent(
		container.NewPadded(
			container.NewBorder(
				wHeadBox,
				wFooterBox,
				nil,
				nil,
				wCoreBox,
			),
		),
	)

	w.SetMaster()
	w.ShowAndRun()
}
