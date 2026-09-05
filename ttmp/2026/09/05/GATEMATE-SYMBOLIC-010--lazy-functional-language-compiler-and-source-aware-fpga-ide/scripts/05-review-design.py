#!/usr/bin/env python3
from pathlib import Path
root=Path(__file__).resolve().parents[1]
p=root/'sources/design.md';s=p.read_text()
s=s.replace('Tokenization must accept -2147483648 as a signed literal; unary minus on arbitrary expressions is not in version 1. Subtraction remains a binary operator.','The lexer emits a separate minus token. In an expression-atom position, the parser may combine minus with an integer token and range-check the signed magnitude, including -2147483648. In an infix position it denotes subtraction, so n-1 parses identically to n - 1. Unary minus on arbitrary expressions is not in version 1.')
s=s.replace('A[Typed syntax tree]','A[Parsed syntax tree]').replace('B[Lexical binding resolution]','B[Resolved bindings and types]')
s=s.replace('Implementation may resolve names before type checking so that both passes use stable binding IDs; the diagram groups the typed syntax outcome rather than prescribing a circular dependency. Concretely use tokenize, parse, allocate binding IDs, resolve references, type-check, then lower.','Use tokenize, parse, allocate binding IDs, resolve references, type-check, then lower. Both binding analysis and type checking use the same stable binding IDs.')
s=s.replace('Freeze their exact subfields in I2\'s schema file before writing either the serial host or RTL decoder; every field must have a named bit range and golden request/response example.','Appendix B specifies their subfields. Transcribe that map into I2\'s schema file and golden request/response examples before writing either the serial host or RTL decoder.')
s=s.replace('The first version is pure and monomorphic.','Identifiers use ASCII letters, digits and underscore, starting with a letter or underscore; keywords are reserved. Whitespace and // comments to end of line separate tokens, and comments may contain Unicode. The first version is pure and monomorphic.')
p.write_text(s)
