import {describe as suite,it,expect} from 'vitest';
import {decode,describe,parseImage,display} from './types';
suite('lazy image representation',()=>{
 it('round trips signed values, shared nodes and cyclic references',()=>{const source=JSON.stringify({root:3,nodes:[{tag:'INT',value:-2147483648},{tag:'INT',value:2},{tag:'ADD',a:0,b:1},{tag:'THUNK',a:2}]});const i=parseImage(source);expect(decode(i.nodes[0]).value).toBe(-2147483648);expect(parseImage(describe(i))).toEqual(i);expect(display(13*2**36+3)).toBe('ERROR(CYCLIC_THUNK)');expect(parseImage('{"root":0,"nodes":[{"tag":"IND","a":0}]}').nodes).toHaveLength(1)});
 it('rejects externally created claims, invalid ranges and noninteger roots',()=>{for(const source of ['{"root":0,"nodes":[{"tag":"BLACKHOLE"}]}','{"root":0,"nodes":[{"tag":"INT","value":2147483648}]}','{"root":0.5,"nodes":[{"tag":"INT","value":1}]}','{"root":0,"nodes":[{"tag":"ADD","a":0,"b":-1}]}'])expect(()=>parseImage(source)).toThrow()});
});
