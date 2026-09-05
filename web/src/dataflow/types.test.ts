import {expect,it} from 'vitest';
import {intValue,valueText,valueType} from './types';
it('preserves signed payloads and discriminates tagged values',()=>{
 expect(valueText(intValue(-2))).toBe('-2');expect(valueText(2**36)).toBe('false');expect(valueText(2**36+1)).toBe('true');expect(valueText(13*2**36+2)).toBe('ERROR 2');expect(valueType(13*2**36+2)).toBe('ERROR');expect(()=>intValue(2**31)).toThrow();expect(()=>intValue(1.5)).toThrow();
});
