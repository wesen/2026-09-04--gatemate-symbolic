#!/usr/bin/env python3
from pathlib import Path
ROOT=next(p for p in Path(__file__).resolve().parents if (p/'symbolic_eval').is_dir())/'symbolic_eval'
p=ROOT/'sim/tb_top.sv';s=p.read_text()
s=s.replace('  integer max_cycles = 300000;', '  integer max_cycles = 300000;\n  integer restart_cycle = 0;\n  logic restarted = 0;')
s=s.replace('    fork\n', '    void\x27($value$plusargs("restart_cycle=%d", restart_cycle));\n\n    fork\n',1)
old='''          @(negedge uart_tx_pin);
          #(BIT_NS + BIT_NS/2);
          for (int j = 0; j < 8; j = j + 1) begin
            b[j] = uart_tx_pin;
            #BIT_NS;
          end
          $display("UARTBYTE %02x", b);'''
new='''          wait (dut.rst_n === 1'b1);
          fork : frame_or_reset
            begin
              @(negedge uart_tx_pin);
              #(BIT_NS + BIT_NS/2);
              for (int j = 0; j < 8; j = j + 1) begin
                b[j] = uart_tx_pin;
                #BIT_NS;
              end
              if (dut.rst_n) $display("UARTBYTE %02x", b);
            end
            begin
              @(negedge dut.rst_n);
            end
          join_any
          disable frame_or_reset;'''
assert s.count(old)==1;s=s.replace(old,new)
s=s.replace('        wait (dut.halted || dut.fault_valid);','        if (restart_cycle > 0) wait (restarted);\n        wait (dut.halted || dut.fault_valid);')
s=s.replace('    $display("TOPDONE', '    if (dut.halted && user_led !== 1\x27b0) $display("FAIL: halted LED is not lit");\n    $display("TOPDONE')
s=s.replace('endmodule','''  initial begin
    #1;
    if (restart_cycle > 0) begin
      repeat (restart_cycle) @(posedge clk_10m);
      fpga_but <= 1'b0;
      repeat (8) @(posedge clk_10m);
      $display("RESTART");
      fpga_but <= 1'b1;
      restarted <= 1'b1;
    end
  end
endmodule''')
p.write_text(s)
p=ROOT/'sim/test_top.py';s=p.read_text();s=s.replace('def _run_top(hex_path, max_cycles=300000):','def _run_top(hex_path, max_cycles=300000, restart_cycle=0):')
s=s.replace('f"+max_cycles={max_cycles}"],','f"+max_cycles={max_cycles}", f"+restart_cycle={restart_cycle}"],')
s+='''

@pytest.mark.parametrize("restart_cycle", [30, 1500])
def test_restart_during_compute_or_uart(restart_cycle):
    stdout = _run_top(_ensure_hex("countdown"), restart_cycle=restart_cycle)
    assert "RESTART" in stdout
    restarted = stdout.split("RESTART", 1)[1]
    got = _bytes_to_str(_uart_lines(restarted))
    expected = "".join(f"T{v.tag:x}:{v.payload:08X}\\r\\n" for v in _model_emits("countdown"))
    assert got == expected
    assert "TOPDONE 1 0 12 0 0" in restarted
''';p.write_text(s)
