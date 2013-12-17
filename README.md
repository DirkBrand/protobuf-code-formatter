ProtoBuf-Code-Formatter
=======================

Code Formatter for Protocol Buffer.  Should be used as a plugin for protoc, but will be extended as a plugin for sublime.

For use in protoc:
Have the location of the plugin binary in your PATH variable, then run the following on the command line:

`$ protoc --pretty_out='location of output' 'location of unformatted .proto file' `


Limitations
===========
1. Formatter cannot preserve order of structures

2. For comments, outer `extend' groups are logically grouped together, so inner comments are lost

3. For comments, trailing comments are not stored for group/nested message/message, so those are lost in formatting.

4. Style of comments are not preserved (/* */ vs. //), so single-line comments are shown with `//` and multi-line comments with `/* */`.

5. Top-of-file comment not preserved (unless attached to package decleration or similar).


[![Build Status](https://drone.io/github.com/DirkBrand/protobuf-code-formatter/status.png)](https://drone.io/github.com/DirkBrand/protobuf-code-formatter/latest)
