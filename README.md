ProtoBuf-Code-Formatter
=======================

Code Formatter for Protocol Buffer.  Should be used as a plugin for protoc, but will be extended as a plugin for sublime.

For use in protoc:
Have the location of the plugin binary in your PATH variable, then run the following on the command line:

$ protoc --CF_out="location of output" "unformatted .proto file" 
