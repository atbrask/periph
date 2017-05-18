// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package pcf8575 controls a Texas Instruments PCF8575 device over I²C.
//
// Interrupt pin on device is not supported.
//
// Datasheet
//
// http://www.ti.com/lit/ds/symlink/pcf8575.pdf

package pcf8575

import (
    "errors"
    "fmt"

    "periph.io/x/periph/conn"
    "periph.io/x/periph/conn/i2c"
    "periph.io/x/periph/devices"
)

// New returns an object that communicates over I²C to a PCF8575 I/O expander.
//
// All outputs are initialized as high (the device's default power-on state).
func New(i i2c.Bus, addr uint16) (*Dev, error) {
    d := &Dev{c: &i2c.Dev{Bus: i, Addr: addr}, lowPins: 0xff, highPins: 0xff}
    err := d.updateState()
    if err != nil {
        return nil, err
    }
    
    return d, nil
}

// Dev is a handle to a pcf8575.
type Dev struct {
    c        conn.Conn // Connection
    lowPins  byte       // State of pins P00-P07
    highPins byte       // State of pins P10-P17
}

func (d *Dev) String() string {
    return fmt.Sprintf("PCF8575{%s}", d.c)
}

func (d *Dev) Halt() error {
    return nil
}

func (d *Dev) WriteOutput(index int, state bool) error {
    if index >= 0 && index < 8 {
        d.lowPins = setBit(d.lowPins, index, state)
    } else if index >= 8 && index < 16 {
        d.highPins = setBit(d.highPins, index - 8, state)
    } else {
        return errors.New(fmt.Sprintf("PCF8575.WriteOutput: Pin index out of range (%d)", index))
    }
    return d.updateState()
}

func (d *Dev) ReadOutput(index int) (bool, error) {
    if index >= 0 && index < 8 {
        return getBit(d.lowPins, index), nil
    } else if index >= 8 && index < 16 {
        return getBit(d.highPins, index - 8), nil
    } else {
        return false, errors.New(fmt.Sprintf("PCF8575.ReadOutput: Pin index out of range (%d)", index))
    }
}

func (d *Dev) ReadInput(index int) (bool, error) {
    s, err := d.readState()
    if err != nil {
        return false, err
    }
    if index >= 0 && index < 8 {
        return getBit(s[0], index), nil
    } else if index >= 8 && index < 16 {
        return getBit(s[1], index - 8), nil
    } else {
        return false, errors.New(fmt.Sprintf("PCF8575.ReadInput: Pin index out of range (%d)", index))
    }
}

func (d *Dev) readState() ([]byte, error) {
    s := []byte {0, 0}
    err := d.c.Tx(nil, s)
    return s, err
}

func (d *Dev) updateState() error {
    return d.c.Tx([]byte{d.lowPins, d.highPins}, nil)
}

func setBit(value byte, index int, state bool) byte {
    if state {
        return value | getMask(index)
    } else {
        return value & ^getMask(index)
    }
}

func getBit(value byte, index int) bool {
    return value & getMask(index) > 0
}

func getMask(index int) byte {
    return 1 << byte(index)
}

var _ devices.Device = &Dev{}
