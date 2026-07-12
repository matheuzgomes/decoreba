package tray

import (
	"fmt"
	"log"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"
)

func Available() bool {
	if sniProbe() == nil {
		return true
	}
	if xembedProbe() {
		return true
	}
	return false
}

func New(showCh chan<- bool, quitCh chan<- struct{}) (*Tray, error) {
	t, err := newSNI(showCh, quitCh)
	if err == nil {
		log.Printf("tray: using SNI (StatusNotifierItem)")
		return t, nil
	}
	log.Printf("tray: SNI unavailable (%v), trying XEmbed fallback", err)

	t, err = newXEmbed(showCh, quitCh)
	if err == nil {
		log.Printf("tray: using XEmbed (_NET_SYSTEM_TRAY)")
		return t, nil
	}
	log.Printf("tray: XEmbed also unavailable (%v)", err)

	return nil, fmt.Errorf("no tray protocol available: SNI=%v, XEmbed=%v", err, err)
}

type Tray struct {
	closeFn func() error
}

func (t *Tray) Close() error {
	if t.closeFn != nil {
		return t.closeFn()
	}
	return nil
}

// ------- SNI backend -------

const (
	sniBusName = "org.decoreba.Decoreba"
	sniObjPath = "/org/decoreba/StatusNotifierItem"
	sniIface   = "org.kde.StatusNotifierItem"
)

type sniHandler struct {
	showCh chan<- bool
	quitCh chan<- struct{}
}

func (h *sniHandler) Activate(x int32, y int32) *dbus.Error {
	select {
	case h.showCh <- true:
	default:
	}
	return nil
}

func (h *sniHandler) SecondaryActivate(x int32, y int32) *dbus.Error {
	select {
	case h.quitCh <- struct{}{}:
	default:
	}
	return nil
}

func (h *sniHandler) ContextMenu(x int32, y int32) *dbus.Error {
	return nil
}

type sniToolTip struct {
	IconName    string
	IconPixmap  []sniToolTipPixmap
	Title       string
	Description string
}

type sniToolTipPixmap struct {
	Width  int32
	Height int32
	Data   []byte
}

func newSNI(showCh chan<- bool, quitCh chan<- struct{}) (*Tray, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, fmt.Errorf("session bus: %w", err)
	}

	reply, err := conn.RequestName(sniBusName, dbus.NameFlagDoNotQueue)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("request name: %w", err)
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		conn.Close()
		return nil, fmt.Errorf("name %s already taken", sniBusName)
	}

	h := &sniHandler{showCh: showCh, quitCh: quitCh}
	if err := conn.Export(h, sniObjPath, sniIface); err != nil {
		conn.Close()
		return nil, fmt.Errorf("export: %w", err)
	}

	propsSpec := map[string]map[string]*prop.Prop{
		sniIface: {
			"Category":    {Value: "ApplicationStatus", Writable: false, Emit: prop.EmitTrue},
			"Id":          {Value: "decoreba", Writable: false, Emit: prop.EmitTrue},
			"Title":       {Value: "decoreba", Writable: false, Emit: prop.EmitTrue},
			"Status":      {Value: "Active", Writable: false, Emit: prop.EmitTrue},
			"WindowId":    {Value: int32(0), Writable: false, Emit: prop.EmitTrue},
			"IconName":    {Value: "utilities-terminal", Writable: false, Emit: prop.EmitTrue},
			"ItemIsMenu":  {Value: false, Writable: false, Emit: prop.EmitTrue},
			"ToolTip": {
				Value: sniToolTip{
					IconName:    "utilities-terminal",
					Title:       "decoreba",
					Description: "Ctrl+Shift+D to toggle",
				},
				Writable: false,
				Emit:     prop.EmitTrue,
			},
		},
	}

	_, err = prop.Export(conn, sniObjPath, propsSpec)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("prop export: %w", err)
	}

	watcher := conn.Object("org.kde.StatusNotifierWatcher", "/StatusNotifierWatcher")
	call := watcher.Call("org.kde.StatusNotifierWatcher.RegisterStatusNotifierItem", 0, sniBusName)
	if call.Err != nil {
		conn.Close()
		return nil, fmt.Errorf("register watcher: %w", call.Err)
	}

	return &Tray{closeFn: conn.Close}, nil
}

func sniProbe() error {
	conn, err := dbus.SessionBus()
	if err != nil {
		return err
	}
	defer conn.Close()

	watcher := conn.Object("org.kde.StatusNotifierWatcher", "/StatusNotifierWatcher")
	return watcher.Call("org.freedesktop.DBus.Peer.Ping", 0).Err
}
