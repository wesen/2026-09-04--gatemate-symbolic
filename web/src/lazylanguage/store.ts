import {configureStore,createSlice,PayloadAction} from '@reduxjs/toolkit';
import {createApi,fetchBaseQuery} from '@reduxjs/toolkit/query/react';
import type {ArtifactView,CompileResult,Frame,Operation,State} from './types';
export const api=createApi({reducerPath:'languageApi',baseQuery:fetchBaseQuery({baseUrl:'/api/language/'}),tagTypes:['State'],endpoints:build=>({
 state:build.query<State,void>({query:()=> 'state',providesTags:['State']}),
 examples:build.query<Record<string,string>,void>({query:()=> 'examples'}),
 artifact:build.query<ArtifactView,string>({query:id=>`artifacts/${id}`}),
 history:build.query<Frame,number>({query:id=>`history/${id}`}),
 compile:build.mutation<CompileResult,{source:string;clientRevision:number}>({query:body=>({url:'compile',method:'POST',body})}),
 control:build.mutation<State,Operation>({query:body=>({url:'control',method:'POST',body}),invalidatesTags:['State']}),
})});
const editor=createSlice({name:'editor',initialState:{source:'',revision:0,compiledID:'',compiledRevision:-1,diagnostics:[] as CompileResult['diagnostics'],selectedRef:-1,selectedSpan:0,historyID:0},reducers:{
 edit(state,action:PayloadAction<string>){state.source=action.payload;state.revision++;state.diagnostics=[];},
 compiled(state,action:PayloadAction<CompileResult>){if(action.payload.clientRevision!==state.revision)return;state.compiledID=action.payload.artifactId;state.compiledRevision=state.revision;state.diagnostics=action.payload.diagnostics;},
 selectRef(state,action:PayloadAction<number>){state.selectedRef=action.payload;},
 selectSpan(state,action:PayloadAction<number>){state.selectedSpan=action.payload;},
 history(state,action:PayloadAction<number>){state.historyID=action.payload;state.selectedRef=-1;state.selectedSpan=0;},
}});
export const actions=editor.actions;
export const editorReducer=editor.reducer;
export const store=configureStore({reducer:{editor:editor.reducer,[api.reducerPath]:api.reducer},middleware:getDefault=>getDefault({serializableCheck:false}).concat(api.middleware)});
export type RootState=ReturnType<typeof store.getState>;
