// Graph coloring with runtime adjacency and lossless semantic event backpressure.
// Seeded from the queens laboratory; initial singleton propagation permits level zero.
// Every observable write has one owner; trail persistence precedes publication.
`default_nettype none
module graph_core #(
  parameter bit USE_TRAIL=1,
  parameter int TRAIL_CAPACITY=64,
  parameter int CHOICE_CAPACITY=8
) (
  input logic clk, rst_n,
  input logic [63:0] initial_domains, adjacency,
  input logic [7:0] active_vertices,
  input logic first_only, trace_ready,
  output wire [39:0] choice_debug,
  output wire [19:0] trail_debug,
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
    S_RESTORE,S_RETRY,S_UPDATE_WRITE,S_OUTPUT,S_COMPLETE,S_STOP,
    S_LOG_WRITE,S_APPLY,S_TCHECK,S_TWAIT,S_TCAPTURE,S_TAPPLY,S_INIT} state_t;
  state_t state_q, write_resume_q;
  logic [7:0] domains_q[0:7];
  logic [7:0] propagated_q;
  logic [3:0] choice_top_q;
  logic [103:0] cp_q;
  logic [2:0] source_q,row_q,target_q,mut_col_q;
  logic [7:0] mut_mask_q;
  logic [23:0] pending_q;
  logic [6:0] trail_top_q, trail_base_q;
  logic [7:0] old_mask_q;
  logic [19:0] trail_entry_q;
  logic zero_found, source_found, unresolved_found;
  logic [2:0] next_source,next_variable;
  localparam integer CP_WIDTH=USE_TRAIL ? 40 : 104;
  wire [CP_WIDTH-1:0] cp_ram_data;
  wire [103:0] cp_rd_data={{(104-CP_WIDTH){1'b0}},cp_ram_data};
  wire advance = !trace_valid || trace_ready;
  wire cp_wr_en = advance && (state_q==S_CP_WRITE || state_q==S_UPDATE_WRITE);
  wire [2:0] cp_wr_addr = state_q==S_CP_WRITE ? choice_top_q[2:0] : 3'(choice_top_q-1'b1);
  wire [2:0] cp_rd_addr = choice_top_q != 0 ? 3'(choice_top_q-1'b1) : 3'd0;
  sync_sdp_ram #(.DEPTH(8),.WIDTH(CP_WIDTH)) u_choices(
    .clk(clk),.wr_en(cp_wr_en),.wr_addr(cp_wr_addr),.wr_data(cp_q[CP_WIDTH-1:0]),
    .rd_addr(cp_rd_addr),.rd_data(cp_ram_data));
  wire [19:0] trail_rd_data;
  wire [5:0] trail_rd_addr=trail_top_q!=0 ? 6'(trail_top_q-1'b1) : 6'd0;
  wire [19:0] trail_wr_data={mut_col_q,old_mask_q,1'b0,choice_top_q,4'd0};
  sync_sdp_ram #(.DEPTH(64),.WIDTH(20)) u_trail(
    .clk(clk),.wr_en(advance && USE_TRAIL && state_q==S_LOG_WRITE),.wr_addr(trail_top_q[5:0]),
    .wr_data(trail_wr_data),.rd_addr(trail_rd_addr),.rd_data(trail_rd_data));

  assign choice_debug=cp_q[39:0];
  assign trail_debug=trace_kind==queens_types_pkg::E_RESTORE ? trail_entry_q : trail_wr_data;
  assign propagated_o=propagated_q;
  assign choice_top_o=choice_top_q;
  assign trail_top_o=trail_top_q;
  assign trail_base_o=trail_base_q;
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
      state_q<=S_INIT; write_resume_q<=S_SCAN;
      for(c=0;c<8;c=c+1) domains_q[c]<=8'hff;
      propagated_q<=0; choice_top_q<=0; cp_q<=0;
      trail_top_q<=0;trail_base_q<=0;old_mask_q<=0;trail_entry_q<=0;
      source_q<=0;row_q<=0;target_q<=0;mut_col_q<=0;mut_mask_q<=0;pending_q<=0;
      done<=0;fault_valid<=0;fault_code<=0;solution_count<=0;
      trace_valid<=0;trace_kind<=0;
      cycles<=0;domain_writes<=0;history_writes<=0;history_reads<=0;
      choice_writes<=0;result_stalls<=0;
    end else if(advance) begin
      trace_valid<=0;
      if(state_q!=S_STOP) cycles<=cycles+1;
      if(cp_wr_en) choice_writes<=choice_writes+1;
      case(state_q)
        S_INIT: begin
          for(c=0;c<8;c=c+1) domains_q[c]<=initial_domains[8*c+:8];
          propagated_q<=~active_vertices;state_q<=S_SCAN;
        end
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
          else if(USE_TRAIL && trail_top_q>=TRAIL_CAPACITY) fault(queens_types_pkg::F_TRAIL_FULL);
          else begin
            cp_q<={(USE_TRAIL ? 64'd0 : domains_o),next_variable,
              (domains_q[next_variable] & ~queens_types_pkg::first_bit(domains_q[next_variable])),
              trail_top_q,propagated_q,14'd0};
            mut_col_q<=next_variable;
            mut_mask_q<=queens_types_pkg::first_bit(domains_q[next_variable]);
            write_resume_q<=S_SCAN;state_q<=S_CP_WRITE;
          end
        end
        S_CP_WRITE: begin
          state_q<=S_CP_PUBLISH;
          if(!USE_TRAIL) history_writes<=history_writes+1;
        end
        S_CP_PUBLISH: begin
          choice_top_q<=choice_top_q+1'b1;trace_valid<=1;trace_kind<=queens_types_pkg::E_CREATE;
          state_q<=S_MUTATE;
        end
        S_MUTATE: begin
          if((mut_mask_q & ~domains_q[mut_col_q])!=0) fault(queens_types_pkg::F_TRAIL_INTEGRITY);
          else if(domains_q[mut_col_q]==mut_mask_q) state_q<=S_AFTER_WRITE;
          else if(USE_TRAIL && trail_top_q>=TRAIL_CAPACITY) fault(queens_types_pkg::F_TRAIL_FULL);
          else begin old_mask_q<=domains_q[mut_col_q];state_q<=USE_TRAIL ? S_LOG_WRITE : S_APPLY;end
        end
        S_LOG_WRITE: begin history_writes<=history_writes+1;state_q<=S_APPLY;end
        S_APPLY: begin
          domains_q[mut_col_q]<=mut_mask_q;
          propagated_q<=propagated_q & ~(8'b1<<mut_col_q);
          if(USE_TRAIL) trail_top_q<=trail_top_q+1'b1;
          domain_writes<=domain_writes+1;
          trace_valid<=1;trace_kind<=queens_types_pkg::E_WRITE;
          state_q<=S_AFTER_WRITE;
        end
        S_AFTER_WRITE: state_q<=write_resume_q;
        S_PROP: begin
          if(!queens_types_pkg::singleton(domains_q[source_q])) fault(queens_types_pkg::F_BAD_ONEHOT);
          else if(target_q==source_q) state_q<=S_PROP_ADV;
          else begin
            mut_col_q<=target_q;
            mut_mask_q<=domains_q[target_q] & ~(adjacency[8*int'(source_q)+int'(target_q)] ? domains_q[source_q] : 8'd0);
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
          else begin
            state_q<=S_CP_WAIT;
            if(!USE_TRAIL) history_reads<=history_reads+1;
          end
        end
        S_CP_WAIT: state_q<=S_CP_CAPTURE;
        S_CP_CAPTURE: begin
          if(cp_rd_data[13:0]!=0) fault(queens_types_pkg::F_TRAIL_INTEGRITY);
          else begin cp_q<=cp_rd_data;state_q<=USE_TRAIL ? S_TCHECK : S_RESTORE;end
        end
        S_TCHECK: begin
          if(cp_q[28:22]>trail_top_q || cp_q[28:22]<trail_base_q)
            fault(queens_types_pkg::F_TRAIL_INTEGRITY);
          else if(trail_top_q==cp_q[28:22]) state_q<=S_RESTORE;
          else begin history_reads<=history_reads+1;state_q<=S_TWAIT;end
        end
        S_TWAIT: state_q<=S_TCAPTURE;
        S_TCAPTURE: begin trail_entry_q<=trail_rd_data;state_q<=S_TAPPLY;end
        S_TAPPLY: begin
          if(trail_top_q<=trail_base_q || trail_entry_q[3:0]!=0 ||
             trail_entry_q[8:4]>choice_top_q ||
             trail_entry_q[16:9]==0 ||
             (domains_q[trail_entry_q[19:17]] & ~trail_entry_q[16:9])!=0)
            fault(queens_types_pkg::F_TRAIL_INTEGRITY);
          else begin
            domains_q[trail_entry_q[19:17]]<=trail_entry_q[16:9];
            trail_top_q<=trail_top_q-1'b1;
            trace_valid<=1;trace_kind<=queens_types_pkg::E_RESTORE;state_q<=S_TCHECK;
          end
        end
        S_RESTORE: begin
          if(!USE_TRAIL) for(c=0;c<8;c=c+1) domains_q[c]<=cp_q[40+8*c+:8];
          propagated_q<=cp_q[21:14];
          trace_valid<=1;trace_kind<=queens_types_pkg::E_RESTORED;state_q<=S_RETRY;
        end
        S_RETRY: begin
          if(cp_q[36:29]==0) begin
            choice_top_q<=choice_top_q-1'b1;trace_valid<=1;
            trace_kind<=queens_types_pkg::E_POP;state_q<=S_BACK;
          end else if(USE_TRAIL && trail_top_q>=TRAIL_CAPACITY) fault(queens_types_pkg::F_TRAIL_FULL);
          else begin
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
            if(first_only) begin choice_top_q<=0;trail_base_q<=trail_top_q;state_q<=S_COMPLETE;end
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
    if(TRAIL_CAPACITY<0 || TRAIL_CAPACITY>64) $fatal(1,"invalid trail capacity");
  end
endmodule
`default_nettype wire
