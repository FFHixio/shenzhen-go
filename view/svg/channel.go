// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"github.com/gopherjs/gopherjs/js"
)

type Channel struct {
	Type string
	Cap  int

	Pins map[*Pin]struct{}

	steiner *js.Object // symbol representing the channel itself, not used if channel is simple
	x, y    float64    // centre of steiner point, for snapping
	tx, ty  float64    // temporary centre of steiner point, for display
	l, c    *js.Object // for dragging to more pins
	p       *Pin       // considering attaching to this pin
}

func newChannel(p, q *Pin) *Channel {
	ch := &Channel{
		Type: p.Type,
		Pins: map[*Pin]struct{}{
			p: struct{}{},
			q: struct{}{},
		},
	}
	ch.makeElements()
	return ch
}

func (c *Channel) makeElements() {
	c.steiner = makeSVGElement("circle")
	c.l = makeSVGElement("line")
	c.c = makeSVGElement("circle")

	diagramSVG.Call("appendChild", c.steiner)
	diagramSVG.Call("appendChild", c.l)
	diagramSVG.Call("appendChild", c.c)

	c.steiner.Call("setAttribute", "r", pinRadius)
	c.steiner.Call("addEventListener", "mousedown", c.dragStart)

	c.l.Call("setAttribute", "stroke-width", lineWidth)
	c.l.Call("setAttribute", "display", "none")
	c.c.Call("setAttribute", "r", pinRadius)
	c.c.Call("setAttribute", "fill", "transparent")
	c.c.Call("setAttribute", "stroke-width", lineWidth)
	c.c.Call("setAttribute", "display", "none")
}

func (c *Channel) Pt() (x, y float64) { return c.x, c.y }

func (c *Channel) commit() { c.x, c.y = c.tx, c.ty }

func (c *Channel) dragStart(e *js.Object) {
	dragItem = c

	c.steiner.Call("setAttribute", "display", "")
	c.setColour(activeColour)

	x, y := cursorPos(e)
	c.reposition(ephemeral{x, y})
	c.l.Call("setAttribute", "x1", x)
	c.l.Call("setAttribute", "y1", y)
	c.l.Call("setAttribute", "x2", c.tx)
	c.l.Call("setAttribute", "y2", c.ty)
	c.c.Call("setAttribute", "cx", x)
	c.c.Call("setAttribute", "cy", y)
	c.c.Call("setAttribute", "display", "")
	c.l.Call("setAttribute", "display", "")
}

func (c *Channel) drag(e *js.Object) {
	x, y := cursorPos(e)
	c.steiner.Call("setAttribute", "display", "")
	c.l.Call("setAttribute", "x1", x)
	c.l.Call("setAttribute", "y1", y)
	c.c.Call("setAttribute", "cx", x)
	c.c.Call("setAttribute", "cy", y)
	d, q := graph.nearestPoint(x, y)
	p, _ := q.(*Pin)

	if p != nil && p == c.p && d < snapQuad {
		return
	}

	if c.p != nil && (c.p != p || d >= snapQuad) {
		c.p.disconnect()
		c.p.circ.Call("setAttribute", "fill", normalColour)
		c.p.l.Call("setAttribute", "display", "none")
		c.p = nil
	}

	noSnap := func() {
		c.c.Call("setAttribute", "display", "")
		c.l.Call("setAttribute", "display", "")
		c.reposition(ephemeral{x, y})
	}

	if d >= snapQuad || q == c || (p != nil && p.ch == c) {
		noSnap()
		c.setColour(activeColour)
		return
	}

	if p == nil || p.ch != nil {
		noSnap()
		c.setColour(errorColour)
		return
	}

	if err := p.connectTo(c); err != nil {
		noSnap()
		c.setColour(errorColour)
		return
	}

	// Let's snap!
	c.p = p
	p.l.Call("setAttribute", "display", "")
	c.setColour(activeColour)
	c.l.Call("setAttribute", "display", "none")
	c.c.Call("setAttribute", "display", "none")
}

func (c *Channel) drop(e *js.Object) {
	c.reposition(nil)
	c.commit()
	c.setColour(normalColour)
	if c.p != nil {
		c.p = nil
		return
	}
	c.c.Call("setAttribute", "display", "none")
	c.l.Call("setAttribute", "display", "none")
	if len(c.Pins) <= 2 {
		c.steiner.Call("setAttribute", "display", "none")
	}
}

func (c *Channel) reposition(additional Point) {
	np := len(c.Pins)
	if additional != nil {
		np++
	}
	if np < 2 {
		// Not actually a channel anymore - hide.
		c.steiner.Call("setAttribute", "display", "none")
		for t := range c.Pins {
			t.c.Call("setAttribute", "display", "none")
			t.l.Call("setAttribute", "display", "none")
		}
		return
	}
	c.tx, c.ty = 0, 0
	if additional != nil {
		c.tx, c.ty = additional.Pt()
	}
	for t := range c.Pins {
		c.tx += t.x
		c.ty += t.y
	}
	n := float64(np)
	c.tx /= n
	c.ty /= n
	c.steiner.Call("setAttribute", "cx", c.tx)
	c.steiner.Call("setAttribute", "cy", c.ty)
	c.l.Call("setAttribute", "x2", c.tx)
	c.l.Call("setAttribute", "y2", c.ty)
	for t := range c.Pins {
		t.l.Call("setAttribute", "x2", c.tx)
		t.l.Call("setAttribute", "y2", c.ty)
	}
	disp := ""
	if np <= 2 {
		disp = "none"
	}
	c.steiner.Call("setAttribute", "display", disp)
}

func (c *Channel) setColour(col string) {
	c.steiner.Call("setAttribute", "fill", col)
	c.c.Call("setAttribute", "stroke", col)
	c.l.Call("setAttribute", "stroke", col)
	for t := range c.Pins {
		t.circ.Call("setAttribute", "fill", col)
		t.l.Call("setAttribute", "stroke", col)
	}
}
