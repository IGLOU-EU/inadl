package main

import (
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"git.iglou.eu/adrien/inadl/ina"
)

// Config is a config struct ...
type Config struct {
	url  string
	name string
	path fyne.ListableURI
}

const (
	programVs   string = "V 0.0.1"
	programID   string = "eu.git.iglou.adrien.inadl"
	programName string = "INA video backup"
	programBy   string = "Open-source software by Iglou.eu"
)

var config Config

func main() {
	var video ina.MRSS
	var err error

	// Init
	a := app.NewWithID(programID)
	a.SetIcon(fyne.NewStaticResource("Icon", appIcon))

	w := a.NewWindow(programName)
	w.CenterOnScreen()
	w.Resize(fyne.NewSize(1000, 600))

	// Core section
	var cCoreImageData image.Image
	cCoreImage := canvas.NewImageFromImage(cCoreImageData)
	cCoreImage.FillMode = canvas.ImageFillContain
	cCoreImage.ScaleMode = canvas.ImageScaleSmooth
	cCoreImage.SetMinSize(fyne.NewSize(300, 300))

	var wCoreQualSelect *widget.Select
	wCoreQualSelect = widget.NewSelect(nil, func(s string) {
		config.url = s
	})
	wCoreQualSelect.PlaceHolder = "Meilleure qualité possible"

	var wCoreDownload *widget.Button
	wCoreDownload = widget.NewButton(
		"Télécharger",
		func() {
			if config.path == nil {
				dialog.NewError(fmt.Errorf("Vous n'avez pas sélectionné d'emplacement de sauvegarde"), w).Show()
				return
			}

			if config.url == "" {
				dialog.NewError(fmt.Errorf("Vous n'avez pas sélectionné de qualité ou la vidéo n'est pas disponible au téléchargement"), w).Show()
				return
			}

			wCoreDownload.Text = "Téléchargement en cour ..."
			wCoreDownload.Disable()
			wCoreDownload.Refresh()

			if err := download(config.fileName(), config.url, config.path.Path()); err != nil {
				dialog.NewError(fmt.Errorf("Une erreur c'est produite pendant la tentative de téléchargement\n%s", err), w).Show()
			}

			wCoreDownload.Text = "Télécharger"
			wCoreDownload.Enable()
			wCoreDownload.Refresh()
		},
	)

	wCoreDesc := widget.NewMultiLineEntry()
	wCoreDesc.Wrapping = fyne.TextWrapWord
	wCoreDescBox := widget.NewCard("", "", wCoreDesc)

	cCoreOption := container.NewVBox(wCoreQualSelect, wCoreDownload)
	cCoreContent := container.New(
		layout.NewFormLayout(),
		widget.NewCard(
			"",
			"",
			fyne.NewContainerWithLayout(
				layout.NewBorderLayout(
					nil,
					cCoreOption,
					nil,
					nil,
				),
				cCoreOption,
				cCoreImage,
			),
		),
		wCoreDescBox,
	)

	wCoreLoading := widget.NewTextGrid()
	cCoreLoad := container.NewCenter(wCoreLoading)

	wCoreBox := container.NewMax(cCoreLoad)

	// Header section
	wHeadURLLabel := widget.NewLabel("Collez votre url ina.fr: ")
	wHeadURLLabel.TextStyle.Bold = true
	wHeadURLEntry := widget.NewEntry()
	wHeadURLEntry.PlaceHolder = "https://www.ina.fr/video/PUB232175070/playstation-lancement-video.html"
	wHeadURLEntry.TextStyle.Monospace = true
	wHeadURLEntry.OnChanged = func(s string) {
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
			config.url = video.Channel.Item.Content.Hq.URL
		}
		if video.Channel.Item.Content.Mq.URL != "" {
			qlSet = append(qlSet, "Moyenne qualité")
			if config.url == "" {
				config.url = video.Channel.Item.Content.Mq.URL
			}
		}
		if video.Channel.Item.Content.Bq.URL != "" {
			qlSet = append(qlSet, "Basse qualité")
			if config.url == "" {
				config.url = video.Channel.Item.Content.Bq.URL
			}
		}

		if qlSet == nil {
			wCoreQualSelect.PlaceHolder = "Téléchargement indisponible"
			wCoreDownload.Disable()
		} else {
			wCoreQualSelect.Options = qlSet
			wCoreQualSelect.Refresh()
		}

		// Set informations
		config.name = video.Channel.Title

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
		container.NewVBox(wHeadURLLabel, wHeadDirLabel),
		container.NewVBox(wHeadURLEntry, wHeadDirSelect),
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

func (c Config) fileName() string {
	var ext string

	s := strings.Split(c.url, ".")
	l := len(s)

	if l > 0 {
		ext = s[l-1]
	} else {
		ext = "mp4"
	}

	return fmt.Sprintf("%s.%s", c.name, ext)
}

func download(name, url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath.Join(path, name))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
