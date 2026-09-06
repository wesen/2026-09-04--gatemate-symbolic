import {describe,it,expect} from 'vitest';
import {sourceSlice,frameFields} from './types';
import {actions,editorReducer} from './store';
describe('language source identity',()=>{
 it('maps UTF-8 byte spans without treating them as UTF-16 offsets',()=>{expect(sourceSlice('// λ\ndef main',{start:6,end:9})).toBe('def');});
 it('ignores compilation responses for older editor revisions',()=>{let state=editorReducer(undefined,actions.edit('a'));state=editorReducer(state,actions.edit('b'));state=editorReducer(state,actions.compiled({clientRevision:1,artifactId:'old',diagnostics:[]}));expect(state.compiledID).toBe('');state=editorReducer(state,actions.compiled({clientRevision:2,artifactId:'new',diagnostics:[]}));expect(state.compiledID).toBe('new');});
 it('decodes only bounded fields from wide frames',()=>{expect(frameFields('06001234000000000000000056780000')).toEqual({kind:6,op:0,a:0x1234,b:0,c:0,saved:0,span:0x5678});});
});
