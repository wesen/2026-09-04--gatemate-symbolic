import {configureStore,createSlice,PayloadAction} from '@reduxjs/toolkit';
import {createApi,fetchBaseQuery} from '@reduxjs/toolkit/query/react';
import type {Frame,Operation,State,Validation} from './types';
export const api=createApi({reducerPath:'dataflowApi',baseQuery:fetchBaseQuery({baseUrl:new URL('/api/dataflow',window.location.href).href}),tagTypes:['State','Projects'],endpoints:build=>({
 state:build.query<State,void>({query:()=>'/state',providesTags:['State']}),
 examples:build.query<Record<string,string>,void>({query:()=>'/examples'}),
 history:build.query<Frame,{id:number;generation:number}>({query:({id,generation})=>`/history/${id}?generation=${generation}`}),
 control:build.mutation<State,{expectedId:number;operation:Operation}>({query:body=>({url:'/control',method:'POST',body}),invalidatesTags:['State']}),
 pause:build.mutation<State,void>({query:()=>({url:'/pause',method:'POST',body:{}}),invalidatesTags:['State']}),
 validate:build.mutation<Validation,string>({query:source=>({url:'/scenario/validate',method:'POST',body:{source}})}),
 run:build.mutation<State,{source:string;mode:'step'|'run';expectedId:number}>({query:body=>({url:'/scenario/run',method:'POST',body}),invalidatesTags:['State']}),
 projects:build.query<string[],void>({query:()=>'/projects',providesTags:['Projects']}),
 project:build.query<{source:string},string>({query:id=>`/projects/${encodeURIComponent(id)}`}),
 save:build.mutation<{saved:boolean},{id:string;source:string}>({query:({id,source})=>({url:`/projects/${encodeURIComponent(id)}`,method:'PUT',body:{source}}),invalidatesTags:['Projects']}),
})});
interface Workspace {source:string;context:number;node:number;history:{id:number;generation:number}|null}
const workspace=createSlice({name:'dataflowWorkspace',initialState:{source:'',context:0,node:0,history:null} as Workspace,reducers:{
 setSource(s,a:PayloadAction<string>){s.source=a.payload},setContext(s,a:PayloadAction<number>){s.context=a.payload},setNode(s,a:PayloadAction<number>){s.node=a.payload},setHistory(s,a:PayloadAction<Workspace['history']>){s.history=a.payload},
}});
export const actions=workspace.actions;
export const makeStore=()=>configureStore({reducer:{workspace:workspace.reducer,[api.reducerPath]:api.reducer},middleware:getDefault=>getDefault().concat(api.middleware)});
export const store=makeStore();export type RootState=ReturnType<typeof store.getState>;export type AppDispatch=typeof store.dispatch;
