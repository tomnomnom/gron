# Output Grammar

```
Input ::= '--'* Statement (Statement | '--')*
Statement ::= Path Space* "=" Space* Value ";" "\n"
Path ::= (BareWord) ("." BareWord | ("[" Key "]"))*
BareWord ::= (UnicodeLu | UnicodeLl | UnicodeLm | UnicodeLo | UnicodeNl | '$' | '_') (UnicodeLu | UnicodeLl | UnicodeLm | UnicodeLo | UnicodeNl | UnicodeMn | UnicodeMc | UnicodeNd | UnicodePc | '$' | '_')*
Key ::= [0-9]+ | String
Value ::= String | Number | "true" | "false" | "null" | "[]" | "{}"
String ::= '"' (UnescapedRune | ("\" (["\/bfnrt] | ('u' Hex))))* '"'
UnescapedRune ::= [^#x0-#x1f"\]
```
