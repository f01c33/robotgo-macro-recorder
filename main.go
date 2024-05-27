package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/template"
	"time"

	hook "github.com/robotn/gohook"
)

type Export struct {
	Code string
}

//go:embed template.go.tmpl
var tmpl string

func main() {
	evChan := hook.Start()
	defer hook.End()

	tmpCode := []string{}
	f, err := os.Create(fmt.Sprintf("%d.go", time.Now().Unix()))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	c := make(chan os.Signal, 5)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM|syscall.SIGABRT)

	t := template.Must(template.New("template.go.tmpl").Parse(tmpl))

	go func() {
		<-c
		err := t.Execute(f, Export{Code: strings.Join(tmpCode, "\n")})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}()

	prevTime, currTime := time.Now(), time.Now()
	tm := int64(0)
	addSleep := func() {
		tm = currTime.Sub(prevTime).Abs().Milliseconds()
		if int(tm) != 0 {
			tmpCode = append(tmpCode, "\t\n")
			tmpCode = append(tmpCode, fmt.Sprintf("\ttime.Sleep(time.Millisecond*%d)", tm))
		}
	}
	pressed := map[string]bool{}
	allPressed := func() string {
		out := ""
		for k, v := range pressed {
			if v {
				out += "\"" + k + "\","
			}
		}
		if len(out) > 0 {
			return out[:len(out)-1]
		}
		return out
	}
	for ev := range evChan {
		if ev.Kind == hook.KeyDown {
			pressed[keyNames[uint32(ev.Rawcode)]] = true
			// addSleep()
			// fmt.Println(keyNames[uint32(ev.Rawcode)], string(ev.Keychar), ev.Rawcode, ev.Mask)
			currTime = ev.When
			prevTime = currTime
		} else if ev.Kind == hook.KeyUp {
			currTime = ev.When
			addSleep()
			tmpCode = append(tmpCode, fmt.Sprintf("\trobotgo.KeyTap(%s)", allPressed()))
			pressed[keyNames[uint32(ev.Rawcode)]] = false
			prevTime = currTime

		} else if ev.Kind == hook.MouseDown {
			currTime = ev.When
			addSleep()
			tmpCode = append(tmpCode, fmt.Sprintf("\trobotgo.Move(%d,%d)", ev.X, ev.Y))
			button := "left"
			if ev.Button == 1 {
				button = "left"
			} else if ev.Button == 2 {
				button = "middle"
			} else if ev.Button == 3 {
				button = "right"
			}
			tmpCode = append(tmpCode, fmt.Sprintf("\trobotgo.Click(\"%s\")", button))

			// fmt.Println(ev.Button, ev.X, ev.Y)
			prevTime = currTime
		}
	}
}
