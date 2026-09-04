import { configureStore, createSlice, type PayloadAction } from '@reduxjs/toolkit';
import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type { Action, Graph, Snapshot, State } from './types';

export const api = createApi({
  reducerPath: 'instrument',
  baseQuery: fetchBaseQuery({ baseUrl: new URL('/api/', window.location.href).href }),
  tagTypes: ['State'],
  endpoints: (builder) => ({
    state: builder.query<State, void>({ query: () => 'state', providesTags: ['State'] }),
    load: builder.mutation<State, Graph>({ query: (body) => ({ url: 'graph', method: 'POST', body }), invalidatesTags: ['State'] }),
    control: builder.mutation<State, Action>({ query: (action) => ({ url: 'control', method: 'POST', body: { action } }), invalidatesTags: ['State'] }),
    event: builder.query<Snapshot, { sequence: number; generation: number }>({ query: ({ sequence, generation }) => `events/${sequence}?generation=${generation}` }),
  }),
});
const initialUI = { sequence: null as number | null, vertex: null as number | null };
export const ui = createSlice({
  name: 'view', initialState: initialUI,
  reducers: {
    inspect: (state, action: PayloadAction<number>) => { state.sequence = action.payload },
    selectVertex: (state, action: PayloadAction<number>) => { state.vertex = action.payload },
    live: (state) => { state.sequence = null },
    reset: () => initialUI,
  },
});
export const makeStore = () => configureStore({ reducer: { [api.reducerPath]: api.reducer, view: ui.reducer }, middleware: (getDefault) => getDefault().concat(api.middleware) });
export type AppStore = ReturnType<typeof makeStore>;
export type RootState = ReturnType<AppStore['getState']>;
export type AppDispatch = AppStore['dispatch'];
