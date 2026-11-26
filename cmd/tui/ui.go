package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TUI struct {
	app          *tview.Application
	client       *APIClient
	tcpAddr      string
	rootPages    *tview.Pages
	contentPages *tview.Pages
	statusBar    *tview.TextView
	navBar       *tview.TextView
	libraryTable *tview.Table
	username     string
	tcpConnected bool
}

func NewTUI(client *APIClient, tcpAddr string) *TUI {
	ui := &TUI{
		app:          tview.NewApplication(),
		client:       client,
		tcpAddr:      tcpAddr,
		rootPages:    tview.NewPages(),
		contentPages: tview.NewPages(),
		statusBar:    tview.NewTextView().SetDynamicColors(true),
		navBar:       tview.NewTextView().SetDynamicColors(true),
		libraryTable: tview.NewTable().SetBorders(true).SetSelectable(true, false),
	}
	ui.setupUI()
	return ui
}

func (ui *TUI) Run() error {
	return ui.app.SetRoot(ui.rootPages, true).EnableMouse(true).Run()
}

func (ui *TUI) setStatus(msg string) {
	ui.app.QueueUpdateDraw(func() {
		ui.statusBar.SetText(" " + msg)
	})
}

func (ui *TUI) setupUI() {
	ui.statusBar.SetText(" Welcome to MangaHub - Please log in or register")

	var usernameInput, passwordInput string
	authForm := tview.NewForm().
		AddInputField("Username", "", 24, nil, func(t string) { usernameInput = strings.TrimSpace(t) }).
		AddPasswordField("Password", "", 24, '*', func(t string) { passwordInput = t }).
		AddButton("Login", func() {
			if usernameInput == "" || passwordInput == "" {
				ui.statusBar.SetText(" [red]Username and password are required")
				return
			}
			ui.statusBar.SetText(" [yellow]Logging in...")
			go func(u, p string) {
				if err := ui.client.Login(u, p); err != nil {
					ui.setStatus("[red]Login failed: " + err.Error())
					return
				}
				ui.onAuthSuccess(u)
			}(usernameInput, passwordInput)
		}).
		AddButton("Register", func() {
			if usernameInput == "" || passwordInput == "" {
				ui.statusBar.SetText(" [red]Username and password are required")
				return
			}
			ui.statusBar.SetText(" [yellow]Registering...")
			go func(u, p string) {
				if err := ui.client.Register(u, p); err != nil {
					ui.setStatus("[red]Register failed: " + err.Error())
					return
				}
				ui.onAuthSuccess(u)
			}(usernameInput, passwordInput)
		}).
		AddButton("Quit", func() { ui.app.Stop() })
	authForm.SetBorder(true).SetTitle(" MangaHub Authentication ")

	authContainer := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(authForm, 0, 1, true).
		AddItem(ui.statusBar, 1, 0, false)

	ui.navBar.SetText(" [yellow][1][-] My Library  [yellow][2][-] Catalog (Pending)  [yellow][3][-] Chat (Pending)  [yellow][4][-] Alerts (Pending)  [yellow][5][-] Status  [yellow][q][-] Quit")
	ui.navBar.SetBorder(true).SetTitle(" MangaHub ")

	ui.contentPages.AddPage("library", ui.buildLibraryScreen(), true, true)
	ui.contentPages.AddPage("catalog", ui.buildPlaceholder("Manga Catalog", "Waiting for gRPC Catalog Service (port 9092)..."), true, false)
	ui.contentPages.AddPage("chat", ui.buildPlaceholder("Discussion Rooms", "Waiting for WebSocket Hub (/ws)..."), true, false)
	ui.contentPages.AddPage("notifications", ui.buildPlaceholder("Chapter Notifications", "Waiting for UDP Listener (port 9091)..."), true, false)
	ui.contentPages.AddPage("status", ui.buildStatusScreen(), true, false)

	mainLayout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(ui.navBar, 3, 0, false).
		AddItem(ui.contentPages, 0, 1, true).
		AddItem(ui.statusBar, 1, 0, false)

	mainLayout.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if _, isInput := ui.app.GetFocus().(*tview.InputField); isInput {
			return event
		}
		switch event.Rune() {
		case '1':
			ui.contentPages.SwitchToPage("library")
			go ui.refreshLibrary()
		case '2':
			ui.contentPages.SwitchToPage("catalog")
		case '3':
			ui.contentPages.SwitchToPage("chat")
		case '4':
			ui.contentPages.SwitchToPage("notifications")
		case '5':
			ui.contentPages.SwitchToPage("status")
		case 'q', 'Q':
			ui.app.Stop()
		}
		return nil
	})

	ui.rootPages.AddPage("auth", authContainer, true, true)
	ui.rootPages.AddPage("main", mainLayout, true, false)
}

func (ui *TUI) onAuthSuccess(username string) {
	ui.username = username
	ui.app.QueueUpdateDraw(func() {
		ui.rootPages.SwitchToPage("main")
		ui.statusBar.SetText(fmt.Sprintf(" [green]Logged in as %s", username))
	})
	go ui.refreshLibrary()
	go ui.startTCPSync()
}

func (ui *TUI) buildLibraryScreen() tview.Primitive {
	ui.libraryTable.SetBorder(true).SetTitle(" My Library ")
	ui.renderLibraryHeaders()

	var mangaIDStr, statusStr, chapStr string
	statusOptions := []string{"reading", "plan_to_read", "completed", "dropped"}
	statusStr = statusOptions[0]

	actionForm := tview.NewForm().
		AddInputField("Manga ID", "", 12, nil, func(t string) { mangaIDStr = strings.TrimSpace(t) }).
		AddDropDown("Status", statusOptions, 0, func(option string, _ int) { statusStr = option }).
		AddInputField("Chapter", "", 12, nil, func(t string) { chapStr = strings.TrimSpace(t) }).
		AddButton("Add Manga", func() {
			mID, err := strconv.ParseInt(mangaIDStr, 10, 64)
			if err != nil || mID <= 0 {
				ui.statusBar.SetText(" [red]Enter a valid numeric Manga ID")
				return
			}
			go func() {
				if _, err := ui.client.AddToLibrary(mID, statusStr); err != nil {
					ui.setStatus("[red]" + err.Error())
					return
				}
				ui.setStatus(fmt.Sprintf("[green]Added Manga #%d (%s)", mID, statusStr))
				ui.refreshLibrary()
			}()
		}).
		AddButton("Update Chapter", func() {
			mID, err := strconv.ParseInt(mangaIDStr, 10, 64)
			if err != nil || mID <= 0 {
				ui.statusBar.SetText(" [red]Enter a valid numeric Manga ID")
				return
			}
			chap, err := strconv.Atoi(chapStr)
			if err != nil || chap < 0 {
				ui.statusBar.SetText(" [red]Enter a valid non-negative Chapter")
				return
			}
			go func() {
				if _, err := ui.client.UpdateProgress(mID, chap); err != nil {
					ui.setStatus("[red]" + err.Error())
					return
				}
				ui.setStatus(fmt.Sprintf("[green]Updated Manga #%d to Chapter %d", mID, chap))
				ui.refreshLibrary()
			}()
		}).
		AddButton("Delete Manga", func() {
			mID, err := strconv.ParseInt(mangaIDStr, 10, 64)
			if err != nil || mID <= 0 {
				ui.statusBar.SetText(" [red]Enter a valid numeric Manga ID")
				return
			}
			go func() {
				if err := ui.client.DeleteManga(mID); err != nil {
					ui.setStatus("[red]" + err.Error())
					return
				}
				ui.setStatus(fmt.Sprintf("[green]Removed Manga #%d from library", mID))
				ui.refreshLibrary()
			}()
		}).
		AddButton("Refresh", func() { go ui.refreshLibrary() })
	actionForm.SetBorder(true).SetTitle(" Manage Library ")

	return tview.NewFlex().
		AddItem(ui.libraryTable, 0, 2, false).
		AddItem(actionForm, 0, 1, true)
}

func (ui *TUI) renderLibraryHeaders() {
	headers := []string{"Manga ID", "Status", "Current Chapter", "Sync Version"}
	for col, h := range headers {
		ui.libraryTable.SetCell(0, col, tview.NewTableCell(h).
			SetTextColor(tcell.ColorYellow).
			SetSelectable(false).
			SetExpansion(1))
	}
}

func (ui *TUI) refreshLibrary() {
	items, err := ui.client.GetLibrary()
	ui.app.QueueUpdateDraw(func() {
		if err != nil {
			ui.statusBar.SetText(" [red]Library error: " + err.Error())
			return
		}
		ui.libraryTable.Clear()
		ui.renderLibraryHeaders()
		for row, item := range items {
			values := []string{
				strconv.FormatInt(item.MangaID, 10),
				item.Status,
				strconv.Itoa(item.CurrentChapter),
				strconv.Itoa(item.Version),
			}
			for col, value := range values {
				ui.libraryTable.SetCell(row+1, col, tview.NewTableCell(value).SetExpansion(1))
			}
		}
	})
}

func (ui *TUI) buildPlaceholder(title, message string) tview.Primitive {
	view := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("\n\n " + message)
	view.SetBorder(true).SetTitle(" " + title + " ")
	return view
}

func (ui *TUI) buildStatusScreen() tview.Primitive {
	view := tview.NewTextView().SetDynamicColors(true)
	view.SetBorder(true).SetTitle(" System Status ")
	view.SetText("\n [yellow]Backend API[-]  http://localhost:8080\n [yellow]TCP Sync[-]     Connecting...")
	return view
}

func (ui *TUI) startTCPSync() {
	conn, err := net.DialTimeout("tcp", ui.tcpAddr, 5*time.Second)
	if err != nil {
		ui.setStatus("[red]TCP sync unavailable: " + err.Error())
		return
	}
	defer conn.Close()

	authPayload, err := json.Marshal(map[string]string{
		"type":  "auth",
		"token": ui.client.Token,
	})
	if err != nil {
		ui.setStatus("[red]TCP authentication failed")
		return
	}
	if _, err := conn.Write(append(authPayload, '\n')); err != nil {
		ui.setStatus("[red]TCP authentication failed: " + err.Error())
		return
	}
	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return
	}

	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		ui.setStatus("[red]TCP sync disconnected")
		return
	}
	var authResponse struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(scanner.Bytes(), &authResponse); err != nil || authResponse.Type != "auth_ok" {
		message := authResponse.Message
		if message == "" {
			message = "authentication rejected"
		}
		ui.setStatus("[red]TCP sync: " + message)
		return
	}
	_ = conn.SetReadDeadline(time.Time{})
	ui.tcpConnected = true
	ui.setStatus("[green]TCP sync connected")

	var event struct {
		Type           string `json:"type"`
		MangaID        int64  `json:"manga_id"`
		CurrentChapter int    `json:"chapter"`
		Version        int    `json:"version"`
	}
	for scanner.Scan() {
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}
		if event.Type == "progress_updated" {
			ui.setStatus(fmt.Sprintf("[green]Progress synced: Manga #%d Chapter %d (v%d)", event.MangaID, event.CurrentChapter, event.Version))
			ui.refreshLibrary()
		}
	}

	ui.tcpConnected = false
	ui.setStatus("[yellow]TCP sync disconnected")
}
