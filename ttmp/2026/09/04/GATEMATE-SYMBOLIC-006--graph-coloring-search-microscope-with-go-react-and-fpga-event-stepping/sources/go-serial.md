## README

[![Build Status](https://github.com/bugst/go-serial/workflows/test/badge.svg)](https://github.com/bugst/go-serial/actions?workflow=test)

### go.bug.st/serial

A cross-platform serial port library for Go.

#### Documentation and examples

See the package documentation here: [https://pkg.go.dev/go.bug.st/serial](https://pkg.go.dev/go.bug.st/serial)

#### Credits

✨ Thanks to all awesome [contributors](https://github.com/bugst/go-serial/graphs/contributors)! ✨

This software is released under the [BSD 3-clause license](https://github.com/bugst/go-serial/raw/master/LICENSE).

## Documentation

### Overview

Package serial is a cross-platform serial library for the go language.

The canonical import for this library is go.bug.st/serial so the import line is the following:

```
import "go.bug.st/serial"
```

It is possible to get the list of available serial ports with the GetPortsList function:

```
ports, err := serial.GetPortsList()
if err != nil {
    log.Fatal(err)
}
if len(ports) == 0 {
    log.Fatal("No serial ports found!")
}
for _, port := range ports {
    fmt.Printf("Found port: %v\n", port)
}
```

The serial port can be opened with the Open function:

```
mode := &serial.Mode{
    BaudRate: 115200,
}
port, err := serial.Open("/dev/ttyUSB0", mode)
if err != nil {
    log.Fatal(err)
}
```

The Open function needs a "mode" parameter that specifies the configuration options for the serial port. If not specified the default options are 9600\_N81, in the example above only the speed is changed so the port is opened using 115200\_N81. The following snippets shows how to declare a configuration for 57600\_E71:

```
mode := &serial.Mode{
    BaudRate: 57600,
    Parity: serial.EvenParity,
    DataBits: 7,
    StopBits: serial.OneStopBit,
}
```

The configuration can be changed at any time with the SetMode function:

```
err := port.SetMode(mode)
if err != nil {
    log.Fatal(err)
}
```

The port object implements the io.ReadWriteCloser interface, so we can use the usual Read, Write and Close functions to send and receive data from the serial port:

```
n, err := port.Write([]byte("10,20,30\n\r"))
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Sent %v bytes\n", n)

buff := make([]byte, 100)
for {
    n, err := port.Read(buff)
    if err != nil {
        log.Fatal(err)
        break
    }
    if n == 0 {
        fmt.Println("\nEOF")
        break
    }
    fmt.Printf("%v", string(buff[:n]))
}
```

If a port is a virtual USB-CDC serial port (for example an USB-to-RS232 cable or a microcontroller development board) is possible to retrieve the USB metadata, like VID/PID or USB Serial Number, with the GetDetailedPortsList function in the enumerator package:

```
import "go.bug.st/serial/enumerator"

ports, err := enumerator.GetDetailedPortsList()
if err != nil {
    log.Fatal(err)
}
if len(ports) == 0 {
    fmt.Println("No serial ports found!")
    return
}
for _, port := range ports {
    fmt.Printf("Found port: %s\n", port.Name)
    if port.IsUSB {
        fmt.Printf("   USB ID     %s:%s\n", port.VID, port.PID)
        fmt.Printf("   USB serial %s\n", port.SerialNumber)
    }
}
```

for details on USB port enumeration see the documentation of the specific package.

This library tries to avoid the use of the "C" package (and consequently the need of cgo) to simplify cross compiling. Unfortunately the USB enumeration package for darwin (MacOSX) requires cgo to access the IOKit framework. This means that if you need USB enumeration on darwin you're forced to use cgo.

Example (SendAndReceive) [¶](#example-package-SendAndReceive "Go to Example (SendAndReceive)")

This example prints the list of serial ports and use the first one to send a string "10,20,30" and prints the response on the screen.

```
package main

import (
    "fmt"
    "log"
    "strings"

    "go.bug.st/serial"
)

func main() {

    // Retrieve the port list
    ports, err := serial.GetPortsList()
    if err != nil {
        log.Fatal(err)
    }
    if len(ports) == 0 {
        log.Fatal("No serial ports found!")
    }

    // Print the list of detected ports
    for _, port := range ports {
        fmt.Printf("Found port: %v\n", port)
    }

    // Open the first serial port detected at 9600bps N81
    mode := &serial.Mode{
        BaudRate: 9600,
        Parity:   serial.NoParity,
        DataBits: 8,
        StopBits: serial.OneStopBit,
    }
    port, err := serial.Open(ports[0], mode)
    if err != nil {
        log.Fatal(err)
    }

    // Send the string "10,20,30\n\r" to the serial port
    n, err := port.Write([]byte("10,20,30\n\r"))
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Sent %v bytes\n", n)

    // Read and print the response

    buff := make([]byte, 100)
    for {
        // Reads up to 100 bytes
        n, err := port.Read(buff)
        if err != nil {
            log.Fatal(err)
        }
        if n == 0 {
            fmt.Println("\nEOF")
            break
        }

        fmt.Printf("%s", string(buff[:n]))

        // If we receive a newline stop reading
        if strings.Contains(string(buff[:n]), "\n") {
            break
        }
    }
}
```

### Examples

### Constants

This section is empty.

### Variables

[View Source](https://github.com/bugst/go-serial/blob/v1.8.0/serial.go#L60)
```
var NoTimeout time.Duration = -1
```

NoTimeout should be used as a parameter to SetReadTimeout to disable timeout.

### Functions

#### ¶

```
func GetPortsList() ([]string, error)
```

GetPortsList retrieve the list of available serial ports

Example [¶](#example-GetPortsList "Go to Example")

```
package main

import (
    "fmt"
    "log"

    "go.bug.st/serial"
)

func main() {
    ports, err := serial.GetPortsList()
    if err != nil {
        log.Fatal(err)
    }
    if len(ports) == 0 {
        fmt.Println("No serial ports found!")
    } else {
        for _, port := range ports {
            fmt.Printf("Found port: %v\n", port)
        }
    }
}
```

### Types

#### ¶

```
type Mode struct {
    BaudRate          int              // The serial port bitrate (aka Baudrate)
    DataBits          int              // Size of the character (must be 5, 6, 7 or 8)
    Parity            Parity           // Parity (see Parity type for more info)
    StopBits          StopBits         // Stop bits (see StopBits type for more info)
    InitialStatusBits *ModemOutputBits // Initial output modem bits status (if nil defaults to DTR=true and RTS=true)
}
```

Mode describes a serial port configuration.

#### added in v1.4.0

```
type ModemOutputBits struct {
    RTS bool // ReadyToSend status
    DTR bool // DataTerminalReady status
}
```

ModemOutputBits contains all the modem output bits for a serial port. This is used in the Mode.InitialStatusBits struct to specify the initial status of the bits. Note: Linux and MacOSX (and basically all unix-based systems) can not set the status bits before opening the port, even if the initial state of the bit is set to false they will go anyway to true for a few milliseconds, resulting in a small pulse.

#### ¶

```
type ModemStatusBits struct {
    CTS bool // ClearToSend status
    DSR bool // DataSetReady status
    RI  bool // RingIndicator status
    DCD bool // DataCarrierDetect status
}
```

ModemStatusBits contains all the modem input status bits for a serial port (CTS, DSR, etc...). It can be retrieved with the Port.GetModemStatusBits() method.

#### ¶

```
type Parity int
```

Parity describes a serial port parity setting

```
const (
    // NoParity disable parity control (default)
    NoParity Parity = iota
    // OddParity enable odd-parity check
    OddParity
    // EvenParity enable even-parity check
    EvenParity
    // MarkParity enable mark-parity (always 1) check
    MarkParity
    // SpaceParity enable space-parity (always 0) check
    SpaceParity
)
```

#### ¶

```
type Port interface {
    // SetMode sets all parameters of the serial port
    SetMode(mode *Mode) error

    // Stores data received from the serial port into the provided byte array
    // buffer. The function returns the number of bytes read.
    //
    // The Read function blocks until (at least) one byte is received from
    // the serial port or an error occurs.
    Read(p []byte) (n int, err error)

    // Send the content of the data byte array to the serial port.
    // Returns the number of bytes written.
    Write(p []byte) (n int, err error)

    // Wait until all data in the buffer are sent
    Drain() error

    // ResetInputBuffer Purges port read buffer
    ResetInputBuffer() error

    // ResetOutputBuffer Purges port write buffer
    ResetOutputBuffer() error

    // SetDTR sets the modem status bit DataTerminalReady
    SetDTR(dtr bool) error

    // SetRTS sets the modem status bit RequestToSend
    SetRTS(rts bool) error

    // GetModemStatusBits returns a ModemStatusBits structure containing the
    // modem status bits for the serial port (CTS, DSR, etc...)
    GetModemStatusBits() (*ModemStatusBits, error)

    // SetReadTimeout sets the timeout for the Read operation or use serial.NoTimeout
    // to disable read timeout.
    SetReadTimeout(t time.Duration) error

    // Close the serial port
    Close() error

    // Break sends a break for a determined time
    Break(time.Duration) error
}
```

Port is the interface for a serial Port

#### ¶

```
func Open(portName string, mode *Mode) (Port, error)
```

Open opens the serial port using the specified modes

#### ¶

```
type PortError struct {
    // contains filtered or unexported fields
}
```

PortError is a platform independent error type for serial ports

#### ¶

```
func (e PortError) Code() PortErrorCode
```

Code returns an identifier for the kind of error occurred

#### ¶

```
func (e PortError) EncodedErrorString() string
```

EncodedErrorString returns a string explaining the error code

#### ¶

```
func (e PortError) Error() string
```

Error returns the complete error code with details on the cause of the error

#### ¶

```
type PortErrorCode int
```

PortErrorCode is a code to easily identify the type of error

```
const (
    // PortBusy the serial port is already in used by another process
    PortBusy PortErrorCode = iota
    // PortNotFound the requested port doesn't exist
    PortNotFound
    // InvalidSerialPort the requested port is not a serial port
    InvalidSerialPort
    // PermissionDenied the user doesn't have enough privileges
    PermissionDenied
    // InvalidSpeed the requested speed is not valid or not supported
    InvalidSpeed
    // InvalidDataBits the number of data bits is not valid or not supported
    InvalidDataBits
    // InvalidParity the selected parity is not valid or not supported
    InvalidParity
    // InvalidStopBits the selected number of stop bits is not valid or not supported
    InvalidStopBits
    // InvalidTimeoutValue the timeout value is not valid or not supported
    InvalidTimeoutValue
    // ErrorEnumeratingPorts an error occurred while listing serial port
    ErrorEnumeratingPorts
    // PortClosed the port has been closed while the operation is in progress
    PortClosed
    // FunctionNotImplemented the requested function is not implemented
    FunctionNotImplemented
)
```

#### ¶

```
type StopBits int
```

StopBits describe a serial port stop bits setting

```
const (
    // OneStopBit sets 1 stop bit (default)
    OneStopBit StopBits = iota
    // OnePointFiveStopBits sets 1.5 stop bits
    OnePointFiveStopBits
    // TwoStopBits sets 2 stop bits
    TwoStopBits
)
```

## Directories

| Path | Synopsis |
| --- | --- |
| [enumerator](/go.bug.st/serial@v1.8.0/enumerator)  Package enumerator is a golang cross-platform library for USB serial port discovery. | Package enumerator is a golang cross-platform library for USB serial port discovery. |
| [portlist](/go.bug.st/serial@v1.8.0/portlist) |  |
| [unixutils](/go.bug.st/serial@v1.8.0/unixutils) |  |