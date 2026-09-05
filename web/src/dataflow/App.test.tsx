import {afterEach,beforeEach,expect,it,vi} from 'vitest';
import {render,screen,waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {Provider} from 'react-redux';
import App from './App';
import {api,makeStore} from './store';
import type {State,Snapshot} from './types';
const source=JSON.stringify({version:1,name:'Test',actions:[{kind:'reset'},{kind:'tick',ticks:6}]});
const snap=():Snapshot=>({graph:{count:1,descriptors:[{Op:4,Required:1,Count:0,Final:true,Destinations:[{Node:0,Port:0},{Node:0,Port:0}]}]},debug:{halted:false,reason:0,stopCycle:0,mask:0,node:255,context:255,dropped:0,events:[]},source:'serial',config:{InputDepth:8,CompletionDepth:8,OutputDepth:8,MulLatency:4,EpochBits:8},epochs:[0,0,0,0],closed:[false,false,false,false],slots:Array.from({length:28},(_,n)=>({context:Math.floor(n/7),node:n%7,values:[0,0],valid:0,issued:false,pending:false})),issue:null,router:null,mul:[null,null,null,null],alu:[null],input:[],completion:[],output:[],errors:[null,null,null,null],counters:{cycles:6,activations:1},quiescent:false});
let state:State;let store:ReturnType<typeof makeStore>;let failControl:boolean;let controlCalls:number;
beforeEach(()=>{
 state={current:{program:null,id:2,generation:1,label:'tick',cycle:6,at:'2026-09-04T00:00:00Z',snapshot:snap()},history:[{id:1,generation:1,label:'reset',cycle:0,at:''},{id:2,generation:1,label:'tick',cycle:6,at:''}],results:[],events:[],running:false,scenario:'Test',step:0,actions:2,error:'',needsReset:false};failControl=false;controlCalls=0;
 vi.stubGlobal('fetch',vi.fn(async(input:RequestInfo|URL,init?:RequestInit)=>{
  const req=input instanceof Request?input:new Request(input,init);const path=new URL(req.url).pathname;let data:unknown=state;
  if(path.endsWith('/program/compile')){const {source}=await req.json() as {source:string};data={diagnostics:[],program:{source,graph:snap().graph,inputs:[],constants:[],nodes:[{node:0,name:'output',line:1,type:'int16'}]}}}
  else if(path.endsWith('/examples'))data={book:source,copy:source,fault:source,cancel:source};
  else if((path.endsWith('/projects')||path.endsWith('/programs')))data=[];
  else if(path.endsWith('/validate')){const body=await req.json() as {source:string};let valid=true;try{JSON.parse(body.source)}catch{valid=false}data={diagnostics:valid?[]:[{action:-1,message:'Invalid JSON'}],formatted:source,actions:2}}
  else if(path.includes('/history/'))data={...state.current,id:1,label:'reset',cycle:0,snapshot:{...snap(),counters:{cycles:0}}};
  else if(path.endsWith('/control')){controlCalls++;if(failControl)return new Response(JSON.stringify({error:'UART state uncertain; reset required'}),{status:502,headers:{'Content-Type':'application/json'}})}
  return new Response(JSON.stringify(data),{status:200,headers:{'Content-Type':'application/json'}});
 }));store=makeStore();
});
afterEach(()=>{store.dispatch(api.util.resetApiState());vi.unstubAllGlobals()});
const mount=()=>{const result=render(<Provider store={store}><App/></Provider>);screen.getByText("Low-level scenario experiments").closest("details")!.open=true;return result};
it('labels physical state and makes history strictly observational',async()=>{
 const user=userEvent.setup();mount();await screen.findByText('PHYSICAL FPGA');await waitFor(()=>expect(screen.getByRole('button',{name:'Run from start'})).toBeEnabled());
 await user.selectOptions(screen.getByLabelText('Snapshot history'),'1');await screen.findByText('HISTORICAL SNAPSHOT 1');
 expect(screen.getByRole('button',{name:'Step 1 cycle'})).toBeDisabled();expect(screen.getByRole('button',{name:'Reset engine'})).toBeDisabled();expect(screen.getByRole('button',{name:'Run from start'})).toBeDisabled();expect(controlCalls).toBe(0);
 await user.click(screen.getByRole('button',{name:'Return to live'}));expect(screen.getByRole('button',{name:'Step 1 cycle'})).toBeEnabled();
});
it('blocks invalid source and displays device failures',async()=>{
 const user=userEvent.setup();mount();await screen.findByText('PHYSICAL FPGA');await user.clear(screen.getByRole('textbox',{name:'Scenario JSON source'}));await user.type(screen.getByRole('textbox',{name:'Scenario JSON source'}),'invalid');
 await screen.findByText('Invalid JSON');expect(screen.getByRole('button',{name:'Run from start'})).toBeDisabled();
 failControl=true;await user.click(screen.getByRole('button',{name:'Step 1 cycle'}));expect(await screen.findByRole('alert')).toHaveTextContent('UART state uncertain');
});
it('offers model configuration only for the software source',async()=>{
 state.current.snapshot.source='model';mount();await screen.findByText('TRANSACTION MODEL');expect(await screen.findByText('Model configuration')).toBeInTheDocument();
});

it('requires a fresh compilation after edits and exposes trace loss',async()=>{
 const user=userEvent.setup();state.current.snapshot.debug.dropped=3;state.current.snapshot.debug.halted=true;state.current.snapshot.debug.reason=1;state.current.snapshot.debug.stopCycle=6;
 mount();await screen.findByText('PHYSICAL FPGA');
 expect(screen.getByRole('button',{name:'Load program'})).toBeDisabled();
 await user.click(screen.getByRole('button',{name:'Compile program'}));
 await waitFor(()=>expect(screen.getByRole('button',{name:'Load program'})).toBeEnabled());
 await user.type(screen.getByRole('textbox',{name:'Typed program source'}),' ');
 expect(screen.getByRole('button',{name:'Load program'})).toBeDisabled();
 expect(screen.getByText(/Incomplete trace: 3 events/)).toBeInTheDocument();
 expect(screen.getByRole('button',{name:'Resume execution'})).toBeEnabled();
 await user.selectOptions(screen.getByLabelText('Snapshot history'),'1');
 await waitFor(()=>expect(screen.getByRole('button',{name:'Apply breakpoint'})).toBeDisabled());
});
