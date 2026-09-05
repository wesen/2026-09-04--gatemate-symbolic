import {configureStore,createSlice,PayloadAction} from '@reduxjs/toolkit';
import {createApi,fetchBaseQuery} from '@reduxjs/toolkit/query/react';
import type {Image,Operation,State} from './types';
export const api=createApi({reducerPath:'lazyApi',baseQuery:fetchBaseQuery({baseUrl:'/api/lazy'}),tagTypes:['State'],endpoints:b=>({state:b.query<State,void>({query:()=>'/state',providesTags:['State']}),examples:b.query<Record<string,Image>,void>({query:()=>'/examples'}),control:b.mutation<State,{expectedId:number;operation:Operation}>({query:body=>({url:'/control',method:'POST',body}),invalidatesTags:['State']})})});
const view=createSlice({name:'lazyView',initialState:{selected:0,history:0},reducers:{select:(s,a:PayloadAction<number>)=>{s.selected=a.payload},history:(s,a:PayloadAction<number>)=>{s.history=a.payload}}});
export const actions=view.actions;
export const store=configureStore({reducer:{lazyView:view.reducer,[api.reducerPath]:api.reducer},middleware:g=>g().concat(api.middleware)});
export type RootState=ReturnType<typeof store.getState>;
