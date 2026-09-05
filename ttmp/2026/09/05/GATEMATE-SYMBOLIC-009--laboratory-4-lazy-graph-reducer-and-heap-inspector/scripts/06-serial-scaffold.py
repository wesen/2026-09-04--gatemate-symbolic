#!/usr/bin/env python3
"""Retain the proven bounded UART exchange implementation for the new protocol."""
from pathlib import Path
repo=Path(__file__).resolve().parents[6]
s=(repo/'pkg/dataflow/serial.go').read_text();s=s[:s.index('func (s *Serial) Execute')]
s=s.replace('package dataflow','package lazy').replace('"encoding/binary"','"encoding/binary"\n"encoding/hex"').replace('dataflow UART','lazy UART').replace('dataflow request','lazy request').replace('dataflow response','lazy response')
s=s.replace('if err != nil && !errors.Is(err, ErrFull) && !errors.Is(err, ErrBlocked) {','if err != nil {')
start=s.index('\t\t\tswitch line {');end=s.index('\t\t\tif strings.HasPrefix',start);s=s[:start]+s[end:]
(repo/'pkg/lazy/serial.go').write_text(s)
