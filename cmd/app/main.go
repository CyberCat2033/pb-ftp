package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"os"
	"pb-ftp/internal/control"
	"pb-ftp/internal/netutils"
	"pb-ftp/internal/rescan"
	"pb-ftp/internal/selfupdate"
	ui "pb-ftp/internal/ui"
	"strings"
	"sync"
	"syscall"
	"time"

	ink "github.com/dennwc/inkview"
	"github.com/skip2/go-qrcode"
)

const (
	PORT        = "2121"
	CONTROLPORT = "2122"
)

var (
	code      *qrcode.QRCode
	errorText string
	fun       func()
	ftpServer *netutils.FTPServer
	apiServer *control.Server
)

type App struct {
	drawCount          int
	restartAfterUpdate bool
	mu                 sync.Mutex
}

func (a *App) Init() error {
	var err error

	fun, _ = ink.KeepNetwork()
	ink.ConnectDefault()

	ftpServer, err = netutils.StartVSFTPD()
	if err != nil {
		errorText = err.Error()
		code = nil
		return nil
	}

	apiServer, err = control.Start(
		":"+CONTROLPORT,
		control.WithUpdateHandler(a.applyUpdateAndRestart),
	)
	if err != nil {
		_ = ftpServer.Stop()
		ftpServer = nil
		errorText = err.Error()
		code = nil
		return nil
	}

	ip, err := netutils.GetLocalIP()
	if err != nil {
		errorText = err.Error()
		code = nil
		return nil
	}

	output := netutils.GenerateLink(ip, PORT)

	code, err = qrcode.New(output, qrcode.High)
	if err != nil {
		errorText = err.Error()
		code = nil
		return nil
	}

	errorText = ""
	return nil
}

func drawCenteredStringInRect(y int, text string, font *ink.Font, cl color.Color, bounds image.Rectangle, fontSize int) int {
	if font != nil {
		font.SetActive(cl)
	}
	if text == "" {
		return y
	}

	width := bounds.Dx() - 40
	if width < 0 {
		width = 0
	}
	textWidth := ink.StringWidth(text)
	if textWidth <= width || len(text) < 10 {
		x := bounds.Min.X + (bounds.Dx()-textWidth)/2
		if x < bounds.Min.X {
			x = bounds.Min.X
		}
		ink.DrawString(image.Point{X: x, Y: y}, text)
		return y + fontSize
	}

	splitAt := -1
	delimiters := []string{"@", ":", "/", " "}
	for _, d := range delimiters {
		idx := strings.LastIndex(text, d)
		if idx > 5 && idx < len(text)-5 {
			if ink.StringWidth(text[:idx+1]) <= width {
				if splitAt == -1 || idx > splitAt {
					splitAt = idx
				}
				break
			}
		}
	}

	if splitAt == -1 {
		splitAt = len(text) / 2
	}

	y = drawCenteredStringInRect(y, text[:splitAt+1], font, cl, bounds, fontSize)
	return drawCenteredStringInRect(y+fontSize+4, text[splitAt+1:], font, cl, bounds, fontSize)
}

func (a *App) Close() error {
	var firstErr error

	if ftpServer != nil {
		if err := ftpServer.Stop(); err != nil {
			firstErr = err
		}
	}

	if apiServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		if err := apiServer.Stop(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
		cancel()
	}

	if fun != nil {
		fun()
	}

	a.mu.Lock()
	restartAfterUpdate := a.restartAfterUpdate
	a.mu.Unlock()

	if restartAfterUpdate {
		if err := syscall.Exec(
			selfupdate.LauncherPath,
			[]string{selfupdate.LauncherPath},
			os.Environ(),
		); err != nil && firstErr == nil {
			firstErr = err
		}
		return firstErr
	}

	_ = rescan.TriggerDefault()

	return firstErr
}

func (a *App) applyUpdateAndRestart(request control.UpdateRequest) error {
	err := selfupdate.Apply(selfupdate.Request{
		SourcePath:  request.SourcePath,
		VersionName: request.VersionName,
		VersionCode: request.VersionCode,
		ReleasedAt:  request.ReleasedAt,
		BuildID:     request.BuildID,
		SHA256:      request.SHA256,
	})
	if err != nil {
		return err
	}

	a.mu.Lock()
	a.restartAfterUpdate = true
	a.mu.Unlock()

	go func() {
		time.Sleep(250 * time.Millisecond)
		ink.Exit()
	}()

	return nil
}

func drawCenteredString(y int, text string, font *ink.Font, cl color.Color, screenWidth int, fontSize int) int {
	if font != nil {
		font.SetActive(cl)
	}
	width := ink.StringWidth(text)
	if width <= screenWidth-40 || len(text) < 10 {
		x := (screenWidth - width) / 2
		if x < 0 {
			x = 0
		}
		ink.DrawString(image.Point{X: x, Y: y}, text)
		return y + fontSize
	}

	// Simple wrapping logic
	splitAt := -1
	// Try to split at logical characters for URLs or spaces for text
	delimiters := []string{"@", ":", "/", " "}
	for _, d := range delimiters {
		idx := strings.LastIndex(text, d)
		if idx > 5 && idx < len(text)-5 {
			if ink.StringWidth(text[:idx+1]) <= screenWidth-40 {
				if splitAt == -1 || idx > splitAt {
					splitAt = idx
				}
			}
		}
	}

	if splitAt == -1 {
		// Fallback: split roughly in the middle
		splitAt = len(text) / 2
	}

	// Draw first part
	_ = drawCenteredString(y, text[:splitAt+1], font, cl, screenWidth, fontSize)
	// Draw second part
	return drawCenteredString(y+fontSize+4, text[splitAt+1:], font, cl, screenWidth, fontSize)
}

func (a *App) Draw() {
	ink.ClearScreen()
	size := ink.ScreenSize()
	width := size.X
	height := size.Y

	// Scale font sizes dynamically based on screen height
	titleFontSize := height / 35
	if titleFontSize < 14 {
		titleFontSize = 14
	}
	largeFontSize := height / 25
	if largeFontSize < 18 {
		largeFontSize = 18
	}
	regularFontSize := height / 45
	if regularFontSize < 14 {
		regularFontSize = 14
	}
	smallFontSize := height / 50
	if smallFontSize < 12 {
		smallFontSize = 12
	}

	fontTitle := ink.OpenFont(ink.DefaultFont, titleFontSize, true)
	if fontTitle != nil {
		defer fontTitle.Close()
	}
	fontLarge := ink.OpenFont(ink.DefaultFont, largeFontSize, true)
	if fontLarge != nil {
		defer fontLarge.Close()
	}
	fontRegular := ink.OpenFont(ink.DefaultFont, regularFontSize, true)
	if fontRegular != nil {
		defer fontRegular.Close()
	}
	fontSmall := ink.OpenFont(ink.DefaultFont, smallFontSize, true)
	if fontSmall != nil {
		defer fontSmall.Close()
	}

	// 1. Draw Top Panel (Status Bar)
	topBarHeight := height / 20
	if topBarHeight < 40 {
		topBarHeight = 40
	}

	// Draw device model name on the left
	modelName := ink.DeviceModel()
	if fontSmall != nil {
		fontSmall.SetActive(color.Black)
	}
	ink.DrawString(image.Point{X: 20, Y: (topBarHeight - smallFontSize) / 2}, modelName)

	// Draw battery status on the right
	batteryVal := ink.BatteryPower()
	charging := ink.IsCharging()
	batteryText := fmt.Sprintf("Battery: %d%%", batteryVal)
	if charging {
		batteryText += " ⚡"
	}
	batteryWidth := ink.StringWidth(batteryText)
	ink.DrawString(image.Point{X: width - batteryWidth - 20, Y: (topBarHeight - smallFontSize) / 2}, batteryText)

	// Draw horizontal line separator
	ink.DrawLine(image.Point{X: 0, Y: topBarHeight}, image.Point{X: width, Y: topBarHeight}, color.Black)

	// 2. Draw Bottom Panel (Instructions and Logs)
	bottomHeight := height * 22 / 100 // 22% of screen height
	bottomStartY := height - bottomHeight

	// Draw separator line above the bottom panel
	ink.DrawLine(image.Point{X: 0, Y: bottomStartY}, image.Point{X: width, Y: bottomStartY}, color.Black)

	// 3. Draw QR Code centered in the middle region
	qrRect := image.Rect(0, topBarHeight+10, width, bottomStartY-10)
	if code != nil && errorText == "" {
		_ = ui.DrawQRCenteredIn(code, qrRect)
	} else {
		font := fontRegular
		if font == nil {
			font = fontSmall
		}
		fontSize := regularFontSize
		if font == fontSmall {
			fontSize = smallFontSize
		}
		drawCenteredStringInRect(qrRect.Min.Y+20, errorText, font, color.Black, qrRect, fontSize)
	}

	// 4. Draw contents in the bottom panel
	yCur := bottomStartY + 15

	// Instruction text
	yCur = drawCenteredString(yCur, "Scan QR code or use FTP address to connect:", fontRegular, color.Black, width, regularFontSize)
	yCur += 10

	// FTP URL
	ip, err := netutils.GetLocalIP()
	var ftpURL string
	if err == nil && ip != "" {
		ftpURL = netutils.GenerateLink(ip, PORT)
	} else {
		ftpURL = "No Wi-Fi Connection"
	}

	// Font scaling logic
	selectedFont := fontLarge
	selectedFontSize := largeFontSize

	if fontLarge != nil {
		fontLarge.SetActive(color.Black)
		if ink.StringWidth(ftpURL) > width-40 {
			selectedFont = fontRegular
			selectedFontSize = regularFontSize
			if fontRegular != nil {
				fontRegular.SetActive(color.Black)
				if ink.StringWidth(ftpURL) > width-40 {
					selectedFont = fontSmall
					selectedFontSize = smallFontSize
				}
			}
		}
	}

	yCur = drawCenteredString(yCur, ftpURL, selectedFont, color.Black, width, selectedFontSize)
	yCur += 12

	// Exit instructions
	yCur = drawCenteredString(yCur, "Press any key to exit", fontSmall, color.Black, width, smallFontSize)

	a.mu.Lock()
	a.drawCount++
	drawCount := a.drawCount
	a.mu.Unlock()

	// For the initial draw we do a FullUpdate to clean up screen ghosting, subsequent draws use SoftUpdate
	if drawCount <= 1 {
		ink.FullUpdate()
	} else {
		ink.SoftUpdate()
	}
}

func (a *App) Key(e ink.KeyEvent) bool {
	ink.Exit()
	return true
}

func (a *App) Pointer(e ink.PointerEvent) bool { return false }
func (a *App) Touch(e ink.TouchEvent) bool {
	return false
}

func (a *App) Orientation(o ink.Orientation) bool {
	return false
}

func main() {
	app := &App{}
	ink.Run(app)
}
