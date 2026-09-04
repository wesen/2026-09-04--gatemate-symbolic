// Deterministic full-snapshot baseline. Every observable write has one owner.
`default_nettype none
module queens_core #(
  parameter int CHOICE_CAPACITY=8,
  parameter bit FIRST_ONLY=0
) (
  input logic clk, rst_n,
  output logic result_valid,
  input logic result_ready,
  output logic [23:0] result_data,
  output logic done, fault_valid,
  output logic [3:0] fault_code,
  output logic [31:0] solution_count,
  output logic trace_valid,
  output logic [3:0] trace_kind,
  output wire [63:0] domains_o,
  output wire [7:0] propagated_o,
  output wire [3:0] choice_top_o,
  output wire [6:0] trail_top_o, trail_base_o,
  output logic [31:0] cycles, domain_writes, history_writes, history_reads,
  output logic [31:0] choice_writes, result_stalls
);
  typedef enum logic [4:0] {S_SCAN,S_CHOOSE,S_CP_WRITE,S_CP_PUBLISH,
    S_MUTATE,S_AFTER_WRITE,S_PROP,S_PROP_ADV,S_BACK,S_CP_WAIT,S_CP_CAPTURE,
    S_RESTORE,S_RETRY,S_UPDATE_WRITE,S_UPDATE_PUBLISH,S_OUTPUT,S_COMPLETE,S_STOP} state_t;
  state_t state_q, write_resume_q;
  logic [7:0] domains_q[0:7];
  logic [7:0] propagated_q;
  logic [3:0] choice_top_q;
  logic [103:0] cp_q;
  logic [2:0] source_q,row_q,target_q,mut_col_q;
  logic [7:0] mut_mask_q;
  logic [23:0] pending_q;
  logic zero_found, source_found, unresolved_found;
  logic [2:0] next_source,next_variable;
  wire [103:0] cp_rd_data;
  wire cp_wr_en = state_q==S_CP_WRITE || state_q==S_UPDATE_WRITE;
  wire [2:0] cp_wr_addr = state_q==S_CP_WRITE ? choice_top_q[2:0] : 3'(choice_top_q-1'b1);
  wire [2:0] cp_rd_addr = choice_top_q != 0 ? 3'(choice_top_q-1'b1) : 3'd0;
  sync_sdp_ram #(.DEPTH(8),.WIDTH(104)) u_choices(
    .clk(clk),.wr_en(cp_wr_en),.wr_addr(cp_wr_addr),.wr_data(cp_q),
    .rd_addr(cp_rd_addr),.rd_data(cp_rd_data));

  assign propagated_o=propagated_q;
  assign choice_top_o=choice_top_q;
  assign trail_top_o=0;
  assign trail_base_o=0;
  assign result_valid=state_q==S_OUTPUT;
  assign result_data=pending_q;
  genvar g;
  generate for(g=0;g<8;g=g+1) begin: flatten_domains
    assign domains_o[8*g+:8]=domains_q[g];
  end endgenerate

  integer c;
  always_comb begin
    zero_found=0; source_found=0; unresolved_found=0;
    next_source=0; next_variable=0;
    for(integer k=0;k<8;k=k+1) begin
      if(domains_q[k]==0) zero_found=1;
      if(queens_types_pkg::singleton(domains_q[k]) && !propagated_q[k] && !source_found) begin
        source_found=1; next_source=3'(k);
      end
      if(!queens_types_pkg::singleton(domains_q[k]) && !unresolved_found) begin
        unresolved_found=1; next_variable=3'(k);
      end
    end
  end

  task automatic fault(input logic [3:0] code);
    fault_valid<=1; fault_code<=code; trace_valid<=1;
    trace_kind<=queens_types_pkg::E_FAULT; state_q<=S_STOP;
  endtask

  always_ff @(posedge clk or negedge rst_n) begin
    if(!rst_n) begin
      state_q<=S_SCAN; write_resume_q<=S_SCAN;
      for(c=0;c<8;c=c+1) domains_q[c]<=8'hff;
      propagated_q<=0; choice_top_q<=0; cp_q<=0;
      source_q<=0;row_q<=0;target_q<=0;mut_col_q<=0;mut_mask_q<=0;pending_q<=0;
      done<=0;fault_valid<=0;fault_code<=0;solution_count<=0;
      trace_valid<=0;trace_kind<=0;
      cycles<=0;domain_writes<=0;history_writes<=0;history_reads<=0;
      choice_writes<=0;result_stalls<=0;
    end else begin
      trace_valid<=0;
      if(state_q!=S_STOP) cycles<=cycles+1;
      if(cp_wr_en) choice_writes<=choice_writes+1;
      case(state_q)
        S_SCAN: begin
          if(zero_found) begin
            trace_valid<=1;trace_kind<=queens_types_pkg::E_CONTRADICTION;state_q<=S_BACK;
          end else if(source_found) begin
            source_q<=next_source;row_q<=queens_types_pkg::row_of(domains_q[next_source]);
            target_q<=0;state_q<=S_PROP;
          end else if(!unresolved_found) begin
            for(c=0;c<8;c=c+1) pending_q[3*c+:3]<=queens_types_pkg::row_of(domains_q[c]);
            state_q<=S_OUTPUT;
          end else state_q<=S_CHOOSE;
        end
        S_CHOOSE: begin
          if(choice_top_q>=CHOICE_CAPACITY) fault(queens_types_pkg::F_CHOICE_FULL);
          else begin
            cp_q<={domains_o,next_variable,
              (domains_q[next_variable] & ~queens_types_pkg::first_bit(domains_q[next_variable])),
              7'd0,propagated_q,14'd0};
            mut_col_q<=next_variable;
            mut_mask_q<=queens_types_pkg::first_bit(domains_q[next_variable]);
            write_resume_q<=S_SCAN;state_q<=S_CP_WRITE;
          end
        end
        S_CP_WRITE: begin state_q<=S_CP_PUBLISH;history_writes<=history_writes+1;end
        S_CP_PUBLISH: begin
          choice_top_q<=choice_top_q+1'b1;trace_valid<=1;trace_kind<=queens_types_pkg::E_CREATE;
          state_q<=S_MUTATE;
        end
        S_MUTATE: begin
          if(domains_q[mut_col_q]!=mut_mask_q) begin
            domains_q[mut_col_q]<=mut_mask_q;
            propagated_q<=propagated_q & ~(8'b1<<mut_col_q);
            domain_writes<=domain_writes+1;
            trace_valid<=1;trace_kind<=queens_types_pkg::E_WRITE;
          end
          state_q<=S_AFTER_WRITE;
        end
        S_AFTER_WRITE: state_q<=write_resume_q;
        S_PROP: begin
          if(target_q==source_q) state_q<=S_PROP_ADV;
          else begin
            mut_col_q<=target_q;
            mut_mask_q<=domains_q[target_q] & ~queens_types_pkg::attack(source_q,row_q,target_q);
            write_resume_q<=S_PROP_ADV;state_q<=S_MUTATE;
          end
        end
        S_PROP_ADV: begin
          if(domains_q[target_q]==0) state_q<=S_SCAN;
          else if(target_q==7) begin
            propagated_q<=propagated_q | (8'b1<<source_q);
            trace_valid<=1;trace_kind<=queens_types_pkg::E_PROPAGATED;state_q<=S_SCAN;
          end else begin target_q<=target_q+1'b1;state_q<=S_PROP;end
        end
        S_BACK: begin
          if(choice_top_q==0) state_q<=S_COMPLETE;
          else begin state_q<=S_CP_WAIT;history_reads<=history_reads+1;end
        end
        S_CP_WAIT: state_q<=S_CP_CAPTURE;
        S_CP_CAPTURE: begin cp_q<=cp_rd_data;state_q<=S_RESTORE;end
        S_RESTORE: begin
          for(c=0;c<8;c=c+1) domains_q[c]<=cp_q[40+8*c+:8];
          propagated_q<=cp_q[21:14];
          trace_valid<=1;trace_kind<=queens_types_pkg::E_RESTORED;state_q<=S_RETRY;
        end
        S_RETRY: begin
          if(cp_q[36:29]==0) begin
            choice_top_q<=choice_top_q-1'b1;trace_valid<=1;
            trace_kind<=queens_types_pkg::E_POP;state_q<=S_BACK;
          end else begin
            mut_col_q<=cp_q[39:37];mut_mask_q<=queens_types_pkg::first_bit(cp_q[36:29]);
            cp_q[36:29]<=cp_q[36:29] & ~queens_types_pkg::first_bit(cp_q[36:29]);
            write_resume_q<=S_SCAN;state_q<=S_UPDATE_WRITE;
          end
        end
        S_UPDATE_WRITE: begin
          trace_valid<=1;trace_kind<=queens_types_pkg::E_UPDATE;state_q<=S_MUTATE;
        end
        S_OUTPUT: begin
          if(result_ready) begin
            solution_count<=solution_count+1;trace_valid<=1;trace_kind<=queens_types_pkg::E_OUTPUT;
            if(FIRST_ONLY) begin choice_top_q<=0;state_q<=S_COMPLETE;end
            else state_q<=S_BACK;
          end else result_stalls<=result_stalls+1;
        end
        S_COMPLETE: begin done<=1;trace_valid<=1;trace_kind<=queens_types_pkg::E_COMPLETE;state_q<=S_STOP;end
        S_STOP: state_q<=S_STOP;
        default: fault(queens_types_pkg::F_TRAIL_INTEGRITY);
      endcase
    end
  end
  initial begin
    if(CHOICE_CAPACITY<0 || CHOICE_CAPACITY>8) $fatal(1,"invalid choice capacity");
  end
endmodule
`default_nettype wire
